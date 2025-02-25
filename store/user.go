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

package store

import (
	"time"

	"golang.org/x/exp/slog"
	"gorm.io/gorm"
)

type UserPrivilege int

const (
	UserPrivilegeBlocked UserPrivilege = iota
	UserPrivilegeNormal
	UserPrivilegeAdministrator
	UserPrivilegeHost
)

type User struct {
	gorm.Model `json:"-"` // ID, CreatedAt, UpdatedAt, DeletedAt

	UUID string `gorm:"unique;not null" json:"uuid"`

	// Username
	Username string `gorm:"unique;not null" json:"username"`

	// Nickname
	Nickname string `gorm:"not null" json:"nickname"`

	// SHA256ed Password
	Password string `gorm:"not null" json:"-"`

	// Salt
	Salt string `gorm:"not null" json:"-"`

	// Email
	Email string `gorm:"unique;not null" json:"email"`

	// Is Email Verified
	EmailVerified bool `gorm:"not null" json:"email_verified" priv:"2"`

	// Email Verification Code
	EmailVerificationCode string `json:"-" priv:"3"`

	// Email Verification Code Last Sent
	EmailVerificationCodeLastSent int64 `json:"-" priv:"3"`

	// Email Verification Code Expire
	EmailVerificationCodeExpire int64 `json:"-" priv:"3"`

	// Privilege
	Privilege UserPrivilege `gorm:"not null" json:"privilege"`

	// Registration IP
	RegistrationIP string `gorm:"not null" json:"registration_ip" priv:"2"`

	// Last Login IP
	LastLoginIP string `gorm:"not null" json:"last_login_ip" priv:"2"`

	// Last Login Time
	LastLoginTime int64 `gorm:"not null" json:"last_login_time" priv:"2"`
}

func (s *Store) CreateUser(user User) error {
	return s.db.Create(&user).Error
}

func (s *Store) GetUserByUsername(username string) (*User, error) {
	var user User
	err := s.db.Model(&User{}).Where("username = ?", username).First(&user).Error
	return &user, err
}

func (s *Store) GetUserByEmail(email string) (*User, error) {
	var user User
	err := s.db.Model(&User{}).Where("email = ?", email).First(&user).Error
	return &user, err
}

func (s *Store) GetUserByUsernameOrEmail(userOrEmail string) (*User, error) {
	var user User
	err := s.db.Model(&User{}).Where("username = ? OR email = ?", userOrEmail, userOrEmail).First(&user).Error
	return &user, err
}

func (s *Store) UpdateUser(user *User) error {
	return s.db.Save(user).Error
}

func (s *Store) UpdateLastLogin(user *User, ip string) error {
	user.LastLoginIP = ip
	user.LastLoginTime = time.Now().Unix()
	return s.UpdateUser(user)
}

func (s *Store) UsernameExist(username string) bool {
	var count int64
	err := s.db.Model(&User{}).Where("username = ?", username).Count(&count).Error

	if err != nil {
		slog.Error("Error when checking user existence: ", err.Error())
	}

	return count > 0
}

func (s *Store) EmailExist(email string) bool {
	var count int64
	err := s.db.Model(&User{}).Where("email = ?", email).Count(&count).Error

	if err != nil {
		slog.Error("Error when checking email existence: ", err.Error())
	}

	return count > 0
}

func (s *Store) NicknameExist(nickname string) bool {
	var count int64
	err := s.db.Model(&User{}).Where("nickname = ?", nickname).Count(&count).Error

	if err != nil {
		slog.Error("Error when checking nickname existence: ", err.Error())
	}

	return count > 0
}

func (user *User) HasPrivilege(privilege UserPrivilege) bool {
	return user.Privilege >= privilege
}

func (user *User) IsInTeam(s *Store, game *Game) bool {
	for _, team := range game.GetTeams(s) {
		for _, member := range team.Members {
			if member.ID == user.ID {
				return true
			}
		}
	}
	return false
}
