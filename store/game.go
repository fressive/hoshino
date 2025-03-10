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
	"slices"

	"gorm.io/gorm"
)

type GameStatus int

const (
	GameStatusInactive GameStatus = iota
	GameStatusActive
)

type Game struct {
	gorm.Model `json:"-"`

	// UUID of the game
	UUID string `gorm:"unique" json:"uuid"`

	// Name of the game
	Name string `gorm:"unique" json:"name"`

	// Description of the game,
	// Support Markdown format
	Description string `gorm:"type:text" json:"description"`

	// Status of the game
	Status GameStatus `gorm:"default:0" json:"status"`

	// Visibility of the game
	Visibility bool `gorm:"default:true" json:"visibility"`

	// Start time of the game
	StartTime int64 `gorm:"default:0" json:"start_time"`

	// End time of the game
	EndTime int64 `gorm:"default:0" json:"end_time"`

	// Max team size of the game
	MaxTeamSize int `gorm:"default:1" json:"max_team_size"`

	// The creators/maintainers of the game
	CreatorID uint  `json:"creator_id"`
	Creator   *User `gorm:"foreignKey:CreatorID" json:"creator"`

	// The managers of the game
	Managers []*User `gorm:"many2many:game_managers;" json:"managers"`

	// Enable the change of member during the game
	EnableChangeMember bool `gorm:"default:false" json:"enable_change_member"`

	// The challenges of the game
	Challenges []*Challenge `gorm:"many2many:game_challenges;" json:"challenges"`

	// Flag prefix of the game
	FlagPrefix string `gorm:"default:flag" json:"flag_prefix" priv:"2"`

	// Auto ban the team when cheating
	AutoBan bool `gorm:"default:false" json:"auto_ban" priv:"2"`
}

func (s *Store) CreateGame(game *Game) error {
	return s.db.Create(game).Error
}

func (s *Store) UpdateGame(game *Game) error {
	return s.db.Save(game).Error
}

func (s *Store) GetGames() ([]*Game, error) {
	var games []*Game
	err := s.db.Preload("Creator").Preload("Managers").Preload("Challenges").Preload("Challenges.Creator").Find(&games).Error
	return games, err
}

func (s *Store) GetGameByUUID(uuid string) (*Game, error) {
	var game Game
	err := s.db.Preload("Creator").Preload("Managers").Preload("Challenges").Preload("Challenges.Creator").Where("uuid = ?", uuid).First(&game).Error
	return &game, err
}

func (g Game) IsManager(user *User) bool {
	return slices.Contains(g.Managers, user)
}

func (g *Game) GetChallenges(withInvisible bool) []*Challenge {
	var challenges []*Challenge
	for _, challenge := range g.Challenges {
		if withInvisible || challenge.State == ChallengeStateVisible {
			challenges = append(challenges, challenge)
		}
	}
	return challenges
}

func (g *Game) GetTeams(s *Store) []*Team {
	var teams []*Team
	s.db.Model(Team{}).Preload("Members").Preload("Managers").Where("game_id = ?", g.ID).Find(&teams)
	return teams
}

func (g *Game) GetTeamByUser(s *Store, user *User) *Team {
	for _, team := range g.GetTeams(s) {
		if team.HasMember(user) {
			return team
		}
	}
	return nil
}
