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

type TeamError struct {
	Msg string
}

var (
	TeamBannedError          = &TeamError{Msg: "Team is banned"}
	TeamFullError            = &TeamError{Msg: "Team is full"}
	TeamNotFoundError        = &TeamError{Msg: "Team not found"}
	MemberAlreadyInError     = &TeamError{Msg: "Member has been in the team"}
	MemberInAnotherTeamError = &TeamError{Msg: "Member has been in another team"}
	MemberNotFoundError      = &TeamError{Msg: "Member not found"}
)

func (e *TeamError) Error() string {
	return e.Msg
}

type Team struct {
	gorm.Model `json:"-"`

	Name string `json:"name"`
	UUID string `gorm:"unique" json:"uuid"`

	// The team is under a game
	GameID uint  `gorm:"not null" json:"-"`
	Game   *Game `gorm:"foreignKey:GameID" json:"-"`

	// Is the team banned
	Banned bool `gorm:"default:false" json:"banned"`

	// The creator of the team
	CreatorID uint  `gorm:"not null" json:"-"`
	Creator   *User `gorm:"foreignKey:CreatorID" json:"creator"`

	// The managers and members of the team
	Managers []*User `gorm:"many2many:team_managers;" json:"managers"`
	Members  []*User `gorm:"many2many:team_members;" json:"members"`
}

func (s *Store) CreateTeam(t *Team) error {
	return s.db.Create(t).Error
}

func (s *Store) UpdateTeam(t *Team) error {
	return s.db.Save(t).Error
}

func (t *Team) HasMember(user *User) bool {
	for _, member := range t.Members {
		if member.ID == user.ID {
			return true
		}
	}
	return false
}

func (t *Team) AddMember(s *Store, user *User) error {
	if t.Banned {
		return TeamBannedError
	}

	if t.HasMember(user) {
		return MemberAlreadyInError
	}

	if t.Game.MaxTeamSize > 0 && len(t.Members) >= t.Game.MaxTeamSize {
		return TeamFullError
	}

	if user.IsInTeam(s, t.Game) {
		return MemberInAnotherTeamError
	}

	return s.db.Model(t).Association("Members").Append(&user)
}

func (t *Team) RemoveMember(db *gorm.DB, user *User) error {
	if !t.HasMember(user) {
		return MemberNotFoundError
	}
	return db.Model(t).Association("Members").Delete(&user)
}

func (t *Team) GetMembers(db *gorm.DB) ([]User, error) {
	var users []User
	err := db.Model(t).Association("Members").Find(&users)
	return users, err
}

func (s *Store) GetTeamByName(name string) (*Team, error) {
	var team Team
	err := s.db.Where("name = ?", name).First(&team).Error
	return &team, err
}

func (s *Store) GetTeamByUUID(uuid string) (*Team, error) {
	var team Team
	err := s.db.Where("uuid = ?", uuid).First(&team).Error
	return &team, err
}

func (t *Team) GetTeamScore(s *Store) int {
	var score int
	s.db.Model(Flag{}).Where("team_id = ? AND state >= 1", t.ID).Select("sum(score)").Row().Scan(&score)
	return score
}

func (t *Team) GetTeamRank(s *Store) int64 {
	var rank int64
	s.db.Model(&Team{}).Where("game_id = ? AND (SELECT sum(score) FROM flags WHERE team_id = teams.id AND state >= 1) > ?", t.GameID, t.GetTeamScore(s)).Count(&rank)
	return rank + 1
}
