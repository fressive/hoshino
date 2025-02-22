// Copyright 2025 Rina
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1

import (
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"time"

	"github.com/dlclark/regexp2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/spf13/cast"
	"rina.icu/hoshino/internal/util"
	"rina.icu/hoshino/server/context"
	"rina.icu/hoshino/store"
)

type (
	RegisterPayload struct {
		Username string `json:"username" validate:"required"`
		Nickname string `json:"nickname" validate:"required"`
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required"`
		// TODO: Captcha
	}

	LoginPayload struct {
		Username string `json:"username" validate:"required"`
		Password string `json:"password" validate:"required"`
		// TODO: Captcha
	}

	EmailVerifyPayload struct {
		Code string `json:"code" validate:"required"`
	}
)

var (
	username_regex, _ = regexp2.Compile(`^[a-zA-Z0-9_]{4,16}$`, 0)
	password_regex, _ = regexp2.Compile(`^(?:(?=.*[A-Z])(?=.*[a-z])(?=.*\d)|(?=.*[A-Z])(?=.*[a-z])(?=.*[^\w\s])|(?=.*[A-Z])(?=.*\d)(?=.*[^\w\s])|(?=.*[a-z])(?=.*\d)(?=.*[^\w\s])).{8,}$`, 0)
)

func validateUsername(username string) bool {
	// the username can contain only letters, numbers, and underscores
	// and must be between 4 and 16 characters long

	matched, err := username_regex.MatchString(username)

	if err != nil {
		slog.Error(fmt.Sprintf("Failed to match username regex, error: %v", err))
		return false
	}

	return matched
}

func validatePassword(password string) bool {
	// the password must contain at least 8 characters,
	// include 3 types of the following: uppercase letters, lowercase letters, numbers, and special characters

	matched, err := password_regex.MatchString(password)

	if err != nil {
		slog.Error(fmt.Sprintf("Failed to match password regex, error: %v", err))
		return false
	}

	return matched
}

func validateEmail(store *store.Store, email string) bool {
	// the email must match the regex stored in the database

	email_regex, err := store.GetSetting("email_regex")

	if err != nil {
		slog.Error(fmt.Sprintf("Failed to get email regex from store, error: %v", err))
		return false
	}

	matched, err := regexp.MatchString(cast.ToString(email_regex.Value), email)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to match email regex, error: %v", err))
		return false
	}

	if !matched {
		return false
	}

	return true
}

func UserLogin(c echo.Context) error {
	ctx := c.(*context.CustomContext)
	login := new(LoginPayload)

	if err := c.Bind(login); err != nil {
		return Failed(&c, "Invalid request payload")
	}

	user, err := ctx.Store.GetUserByUsernameOrEmail(login.Username)

	if err != nil {
		return Failed(&c, "Login failed")
	}

	if user.Password != string(util.SHA256WithSalt(login.Password, user.Salt)) {
		return Failed(&c, "Login failed")
	}

	// Login successed
	ctx.Store.UpdateLastLogin(user, c.RealIP())

	// set token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": user.Username,
		"exp":      time.Now().Add(time.Hour * 72).Unix(),
	})

	t, err := token.SignedString([]byte(ctx.Config.Secret))
	if err != nil {
		slog.Error("Failed to sign token: ", slog.Any("err", err))
		return ServerError(&c, "Internal Server Error. Contact the administrator for help.")
	}

	c.SetCookie(&http.Cookie{
		Name:    "token",
		Value:   t,
		Expires: time.Now().Add(time.Hour * 72),
	})

	return OK(&c)
}

func EmailVerify(c echo.Context) error {
	ctx := c.(*context.CustomContext)
	payload := new(EmailVerifyPayload)

	if err := c.Bind(payload); err != nil {
		return Failed(&c, "Invalid request payload")
	}

	user, _ := GetUserFromToken(&c)
	if user.EmailVerified {
		return Failed(&c, "Email has already verified")
	}

	if time.Now().Unix() > user.EmailVerificationCodeExpire {
		return Failed(&c, "Verification code has expired")
	}

	if user.EmailVerificationCode != payload.Code {
		return Failed(&c, "Invalid verification code")
	}

	user.EmailVerified = true
	ctx.Store.UpdateUser(user)

	return OK(&c)
}

func ResendVerificationEmail(c echo.Context) error {
	ctx := c.(*context.CustomContext)

	user, _ := GetUserFromToken(&c)

	if user.EmailVerified {
		return Failed(&c, "Email has already verified")
	}

	if time.Now().Unix()-user.EmailVerificationCodeLastSent < 60 {
		return Failed(&c, "Rate limit exceeded, please try again later")
	}

	if ctx.Store.GetSettingBool("need_email_verify") {
		code := fmt.Sprintf("%s{%s}", ctx.Store.GetSettingString("flag_prefix"), util.UUID())
		user.EmailVerificationCode = code
		user.EmailVerificationCodeExpire = time.Now().Add(time.Minute * 5).Unix()
		user.EmailVerificationCodeLastSent = time.Now().Unix()
		ctx.Store.UpdateUser(user)

		util.SendVerificationEmail(ctx.Store, user.Email, user.Nickname, code)
	} else {
		return Failed(&c, "Email verification is not required")
	}

	return OK(&c)
}

func UserRegister(c echo.Context) error {
	ctx := c.(*context.CustomContext)
	reg := new(RegisterPayload)

	// TODO: Captcha verification

	if err := c.Bind(reg); err != nil {
		slog.Debug(err.Error())
		return Failed(&c, "Invalid request payload")
	}

	if !validateUsername(reg.Username) {
		return Failed(&c, "Invalid username")
	}

	if ctx.Store.UsernameExist(reg.Username) {
		return Failed(&c, "Username already exists")
	}

	if ctx.Store.EmailExist(reg.Email) {
		return Failed(&c, "Email already exists")
	}

	if ctx.Store.NicknameExist(reg.Nickname) {
		return Failed(&c, "Nickname already exists")
	}

	if !validateEmail(ctx.Store, reg.Email) {
		return Failed(&c, "Invalid email")
	}

	if !validatePassword(reg.Password) {
		return Failed(&c, "Invalid password")
	}

	salt := util.GenerateRandomText(16)

	user := store.User{
		Username:       reg.Username,
		Nickname:       reg.Nickname,
		Password:       string(util.SHA256WithSalt(reg.Password, salt)),
		Salt:           salt,
		Email:          reg.Email,
		EmailVerified:  false,
		Privilege:      store.UserPrivilegeNormal,
		RegistrationIP: c.RealIP(),
		LastLoginIP:    c.RealIP(),
		LastLoginTime:  time.Now().Unix(),
	}

	if ctx.Store.GetSettingBool("need_email_verify") {
		code := fmt.Sprintf("%s{%s}", ctx.Store.GetSettingString("flag_prefix"), util.UUID())
		user.EmailVerificationCode = code
		user.EmailVerificationCodeExpire = time.Now().Add(time.Minute * 5).Unix()
		user.EmailVerificationCodeLastSent = time.Now().Unix()
		util.SendVerificationEmail(ctx.Store, user.Email, user.Nickname, code)
	}

	ctx.Store.CreateUser(user)

	return c.JSON(http.StatusOK, map[string]any{"message": "OK", "status": "success", "result": true})
}

func CheckUsername(c echo.Context) error {
	ctx := c.(*context.CustomContext)

	username := c.FormValue("username")

	if !validateUsername(username) {
		return Failed(&c, "Invalid username")
	}

	if ctx.Store.UsernameExist(username) {
		return Failed(&c, "Username already exists")
	} else {
		return OK(&c)
	}
}

func CheckEmail(c echo.Context) error {
	ctx := c.(*context.CustomContext)

	email := c.FormValue("email")

	if !validateEmail(ctx.Store, email) {
		return Failed(&c, "Invalid email")
	}

	if ctx.Store.EmailExist(email) {
		return Failed(&c, "Email already exists")
	} else {
		return OK(&c)
	}
}
