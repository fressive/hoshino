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
	"github.com/spf13/cast"
	"gorm.io/gorm"
)

type ContainerStatus int

const (
	ContainerStatusUnknown ContainerStatus = iota
	ContainerStatusRunning
	ContainerStatusStopped
	ContainerStatusExpired
)

type Container struct {
	gorm.Model

	CreatorID        uint            `json:"creator_id"`
	Creator          *User           `gorm:"foreignKey:CreatorID" json:"creator"`
	ChallengeID      uint            `json:"challenge_id"`
	Challenge        *Challenge      `gorm:"foreignKey:ChallengeID" json:"challenge"`
	UUID             string          `gorm:"unique" json:"uuid"`
	Status           ContainerStatus `gorm:"default:0" json:"status"`
	ExpireTime       int64           `gorm:"default:0" json:"expire_time"`
	LeftRenewalTimes int             `gorm:"default:0" json:"left_renewal_times"`
}

func (s *Store) CanCreateContainer(user *User) bool {
	var count int64

	s.db.Model(Container{}).Where(&Container{
		Creator: user,
		Status:  ContainerStatusRunning,
	}).Count(&count)

	max := s.GetSettingInt("max_container_per_user")

	return count < cast.ToInt64(max)
}

func (s *Store) CreateContainer(c *Container) error {
	return s.db.Create(c).Error
}

func (s *Store) GetContainerByUUID(uuid string) (*Container, error) {
	var c Container

	err := s.db.Where("uuid = ?", uuid).First(&c).Error

	return &c, err
}

func (s *Store) GetContainerByChallengeAndUser(c *Challenge, u *User) (*Container, error) {
	var container Container

	err := s.db.Where("creator_id = ? AND challenge_id = ?", u.ID, c.ID).First(&container).Error

	return &container, err
}

func (s *Store) UpdateContainer(c *Container) error {
	return s.db.Save(c).Error
}
