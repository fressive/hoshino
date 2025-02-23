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
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"rina.icu/hoshino/server/context"
	"rina.icu/hoshino/store"
)

func OK(c *echo.Context) error {
	return (*c).JSON(http.StatusOK, map[string]any{"message": "OK", "status": "success", "result": true})
}

func OKWithData(c *echo.Context, data any) error {
	return (*c).JSON(http.StatusOK, map[string]any{"message": "OK", "status": "success", "result": true, "data": data})
}

func Failed(c *echo.Context, errorMessage string) error {
	return (*c).JSON(http.StatusOK, map[string]any{"message": errorMessage, "status": "error", "result": false})
}

func ServerError(c *echo.Context, errorMessage string) error {
	return (*c).JSON(http.StatusInternalServerError, map[string]any{"message": errorMessage, "status": "error", "result": false})
}

func Unauthorized(c *echo.Context) error {
	return (*c).JSON(http.StatusUnauthorized, map[string]any{"message": "Unauthorized", "status": "error", "result": false})
}

func GetUserFromToken(c *echo.Context) (*store.User, bool) {
	// Warning: Make sure to use this function after the token has been verified

	ctx := (*c).(*context.CustomContext)

	token := (*c).Get("user")
	if token == nil {
		return &store.User{}, false
	}

	claims := token.(*jwt.Token).Claims.(*jwt.MapClaims)
	username := (*claims)["username"].(string)

	user, err := ctx.Store.GetUserByUsername(username)
	if err != nil {
		return &store.User{}, false
	}

	return user, true
}
