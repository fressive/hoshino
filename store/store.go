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
	"errors"

	"gorm.io/gorm"
	common "rina.icu/hoshino/internal"
	"rina.icu/hoshino/server/config"
	"rina.icu/hoshino/store/sqlite"
)

type Store struct {
	config *config.Config
	db     *gorm.DB
}

func (s *Store) initSetting() {
	for k, v := range common.DefaultSettings {
		s.db.Where("key = ?", k).FirstOrCreate(&Setting{
			Key:      k,
			Value:    v,
			Defaults: v,
		})
	}
}

func (s *Store) migrate() {
	s.db.AutoMigrate(
		&Attachment{},
		&Category{},
		&Challenge{},
		&Container{},
		&Flag{},
		&Game{},
		&User{},
		&Setting{},
		&GameEvent{},
		&Team{},
	)
}

func GetStore(c *config.Config) (*Store, error) {
	var db *gorm.DB
	var err error

	if c.Driver == "sqlite" {
		db, err = sqlite.GetDb(c)
	} else {
		err = errors.New("Unrecognizable database driver " + c.Driver)
	}

	store := &Store{c, db}

	store.migrate()

	// init settings of website here
	store.initSetting()

	if err != nil {
		panic(err)
	}

	return store, err
}
