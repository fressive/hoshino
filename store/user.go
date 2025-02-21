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
	gorm.Model

	// Username
	Username string `gorm:"unique;not null"`

	// Nickname
	Nickname string `gorm:"not null"`

	// SHA256ed Password
	Password string `gorm:"not null"`

	// Salt
	Salt string `gorm:"not null"`

	// Email
	Email string `gorm:"unique;not null"`

	// Is Email Verified
	EmailVerified bool `gorm:"not null"`

	// Email Verification Code
	EmailVerificationCode string

	// Email Verification Code Last Sent
	EmailVerificationCodeLastSent int64

	// Email Verification Code Expire
	EmailVerificationCodeExpire int64

	// Privilege
	Privilege UserPrivilege `gorm:"not null"`

	// Registration IP
	RegistrationIP string `gorm:"not null"`

	// Last Login IP
	LastLoginIP string `gorm:"not null"`

	// Last Login Time
	LastLoginTime int64 `gorm:"not null"`
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

func (user *User) IsInTeam(game *Game) bool {
	for _, team := range game.Teams {
		for _, member := range team.Members {
			if member.ID == user.ID {
				return true
			}
		}
	}
	return false
}
