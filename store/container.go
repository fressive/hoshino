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

	CreatedBy  *User           `gorm:"foreignKey:id"`
	UUID       string          `gorm:"unique"`
	Status     ContainerStatus `gorm:"default:0"`
	ExpireTime int64           `gorm:"default:0"`
	Agent      string          `gorm:"default:''"`

	// The number of times the container can be renewed
	LeftRenewalTimes int `gorm:"default:0"`
}

func (s *Store) CanCreateContainer(user *User) bool {
	var count int64

	s.db.Model(Container{}).Where(&Container{
		CreatedBy: user,
		Status:    ContainerStatusRunning,
	}).Count(&count)

	max, err := s.GetSetting("max_container_per_user")
	if err != nil {
		return false
	}

	return count < cast.ToInt64(max)
}
