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
	"reflect"

	"github.com/spf13/cast"
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

func (u *User) UserPriv(s *Store) func(reflect.Type, reflect.Value) int {
	return func(t reflect.Type, v reflect.Value) int {
		switch t {
		case reflect.TypeOf(User{}):
			if u.ID == v.FieldByName("ID").Interface().(uint) {
				// the user self
				return 2
			} else {
				return int(u.Privilege)
			}
		case reflect.TypeOf(Game{}):
			if (v.Convert(t).Interface().(Game)).IsManager(u) {
				// is game manager
				return 2
			} else {
				return int(u.Privilege)
			}
		case reflect.TypeOf(Challenge{}):
			challenge := v.Convert(t).Interface().(Challenge)

			s.db.Preload("Game.Managers").First(&challenge)

			if challenge.Game.IsManager(u) {
				// is game manager
				return 2
			} else {
				return int(u.Privilege)
			}
		default:
			return int(u.Privilege)
		}
	}
}

func FilterFieldsByPrivilege(model interface{}, priv func(reflect.Type, reflect.Value) int) interface{} {
	// field with priv tag will be filtered
	// priv is a function that calculates the privilege level of the current user in specific type

	t := reflect.TypeOf(model)
	v := reflect.ValueOf(model)

	if t.Kind() == reflect.Ptr {
		// deref first
		t = t.Elem()
		v = v.Elem()
	}

	if t.Kind() == reflect.Slice {
		// Recursively filter each element in the slice of structs
		for i := 0; i < v.Len(); i++ {
			elem := v.Index(i)
			nestedModel := FilterFieldsByPrivilege(elem.Interface(), priv)
			elem.Set(reflect.ValueOf(nestedModel))
		}
	}

	if t.Kind() == reflect.Struct {
		// struct, filter fields

		for i := 0; i < t.NumField(); i++ {
			// get fields

			field := t.Field(i)

			if v.Field(i).Kind() == reflect.Ptr && v.Field(i).IsNil() {
				continue
			}

			fieldValue := v.Field(i)
			tag := field.Tag.Get("priv")

			if field.Name == "Model" {
				continue
			}

			if !(tag == "" || cast.ToInt(tag) <= priv(t, v)) {
				// the field is not allowed to be accessed,
				// set the field to zero value and not process the nested fields
				fieldValue.Set(reflect.Zero(field.Type))
				continue
			}

			if field.Type.Kind() == reflect.Struct {
				// struct
				nestedModel := FilterFieldsByPrivilege(fieldValue.Interface(), priv)
				fieldValue.Set(reflect.ValueOf(nestedModel))
			} else if field.Type.Kind() == reflect.Ptr && field.Type.Elem().Kind() == reflect.Struct {
				// *struct
				nestedModel := FilterFieldsByPrivilege(fieldValue.Interface(), priv)
				fieldValue.Set(reflect.ValueOf(nestedModel))
			} else if field.Type.Kind() == reflect.Slice && field.Type.Elem().Kind() == reflect.Struct {
				// []struct

				for i := 0; i < fieldValue.Len(); i++ {
					elem := fieldValue.Index(i)
					nestedModel := FilterFieldsByPrivilege(elem.Interface(), priv)
					elem.Set(reflect.ValueOf(nestedModel))
				}
			} else if field.Type.Kind() == reflect.Slice && field.Type.Elem().Kind() == reflect.Ptr && field.Type.Elem().Elem().Kind() == reflect.Struct {
				// []*struct
				for i := 0; i < fieldValue.Len(); i++ {
					elem := fieldValue.Index(i)
					nestedModel := FilterFieldsByPrivilege(elem.Interface(), priv)
					elem.Set(reflect.ValueOf(nestedModel))
				}
			}
		}
	}

	return model
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
