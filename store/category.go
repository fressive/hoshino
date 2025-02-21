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

import "gorm.io/gorm"

type Category struct {
	gorm.Model
	Name        string `json:"name"`
	Description string `json:"description"`
	Visibility  bool   `json:"visibility"`
	UUID        string `json:"uuid"`

	GameID     uint         `json:"game_id"`
	Game       *Game        `gorm:"foreignKey:GameID" json:"game"`
	Challenges []*Challenge `gorm:"many2many:category_challenges;" json:"challenges"`
}

func (s *Store) CreateCategory(c *Category) error {
	return s.db.Create(c).Error
}

func (s *Store) GetCategoryByUUID(uuid string) (*Category, error) {
	var category Category
	err := s.db.Where("uuid = ?", uuid).First(&category).Error
	return &category, err
}

func (s *Store) GetCategoriesByGame(game *Game) []*Category {
	categories := []*Category{}
	s.db.Model(game).Association("Categories").Find(&categories)
	return categories
}

func (c *Category) GetChallenges(withInvisible bool) []*Challenge {
	challenges := []*Challenge{}
	for _, challenge := range c.Challenges {
		if withInvisible || challenge.State == ChallengeStateVisible {
			challenges = append(challenges, challenge)
		}
	}
	return challenges
}
