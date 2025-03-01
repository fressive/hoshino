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

type FlagState int

const (
	FlagUnsolved FlagState = iota
	FlagSolved
	FlagCheated
)

type Flag struct {
	gorm.Model

	// Flag is the flag string
	// The flag may be not unique in the same challenge in fixed flag challenges
	Flag string `gorm:"not null"`

	// 0: unsolved, 1: solved, 2: cheated
	State FlagState `gorm:"default:0"`

	// SolvedAt is the time when the flag was solved
	SolvedAt int64 `gorm:"default:0"`

	// Score is the score of the flag
	Score int `gorm:"default:-1"`

	// ChallengeID is the ID of the challenge that the flag belongs to
	ChallengeID uint
	Challenge   *Challenge `gorm:"foreignKey:ChallengeID"`

	// ContainerID is the ID of the container that the flag belongs to
	ContainerID uint
	Container   *Container `gorm:"foreignKey:ContainerID"`

	// The team that should submit this flag
	TeamID uint
	Team   *Team `gorm:"foreignKey:TeamID"`
}

func (s *Store) CreateFlag(flag *Flag) error {
	return s.db.Create(flag).Error
}

func (s *Store) RemoveFlag(flag *Flag) error {
	return s.db.Delete(flag).Error
}

func (s *Store) UpdateFlag(flag *Flag) error {
	return s.db.Save(flag).Error
}

func (s *Store) GetFlagByID(id uint) (*Flag, error) {
	var flag Flag
	err := s.db.First(&flag, id).Error
	return &flag, err
}

func (s *Store) GetFlagByChallenge(flag string, challenge *Challenge) (*Flag, error) {
	var f Flag
	err := s.db.Preload("Team").Preload("Challenge").Where("flag = ? AND challenge_id = ?", flag, challenge.ID).First(&f).Error
	return &f, err
}

func (s *Store) GetFlagsByChallenge(flag string, challenge *Challenge) ([]*Flag, error) {
	var flags []*Flag
	err := s.db.Preload("Team").Preload("Challenge").Where("flag = ? AND challenge_id = ?", flag, challenge.ID).Find(&flags).Error
	return flags, err
}

func (s *Store) GetFlagByChallengeAndTeam(challenge *Challenge, team *Team) (*Flag, error) {
	var flag Flag
	err := s.db.Preload("Team").Preload("Challenge").Where("challenge_id = ? AND team_id = ?", challenge.ID, team.ID).Order("ID DESC").First(&flag).Error
	return &flag, err
}

func (s *Store) GetSolvedFlagsByChallenge(challenge *Challenge) ([]*Flag, error) {
	var flags []*Flag
	err := s.db.Preload("Team").Preload("Challenge").Where("challenge_id = ? AND state >= 1", challenge.ID).Order("solved_at ASC").Find(&flags).Error
	return flags, err
}
