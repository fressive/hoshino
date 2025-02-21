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

// The dynamic settings which may be changed during the runtime
type Setting struct {
	gorm.Model

	Key      string `gorm:"unique;not null"`
	Value    string `gorm:"not null"`
	Defaults string `gorm:"not null"`
}

// use cache to reduce the number of database queries
var cache map[string]string = map[string]string{}

func (s *Store) GetSetting(key string) (Setting, error) {
	if value, ok := cache[key]; ok {
		return Setting{Value: value}, nil
	} else {
		var setting Setting
		err := s.db.Where("key = ?", key).First(&setting).Error
		if err == nil {
			cache[key] = setting.Value
		}
		return setting, err
	}
}

func (s *Store) GetSettingString(key string) string {
	setting, err := s.GetSetting(key)

	if err != nil {
		panic(err)
	}

	return setting.Value
}

func (s *Store) GetSettingBool(key string) bool {
	setting := s.GetSettingString(key)
	return setting == "true"
}

func (s *Store) GetSettingInt(key string) int {
	setting := s.GetSettingString(key)
	return cast.ToInt(setting)
}

func (s *Store) SetSetting(key string, value string) error {
	err := s.db.Model(&Setting{}).Where("key = ?", key).Update("value", value).Error

	if err == nil {
		cache[key] = value
	}
	return err
}

func (s *Store) GetAllSettings() ([]Setting, error) {
	var settings []Setting
	err := s.db.Find(&settings).Error
	return settings, err
}
