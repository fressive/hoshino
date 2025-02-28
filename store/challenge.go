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

	"gorm.io/gorm"
	"rina.icu/hoshino/store/types"
)

type ChallengeState int

const (
	ChallengeStateDisabled ChallengeState = iota
	ChallengeStateHidden
	ChallengeStateVisible
)

type AfterExpireOp int

const (
	AfterExpireDisableAll AfterExpireOp = iota
	AfterExpireCreateContainer
	AfterExpireSubmitFlag
)

type ScoreMode int

const (
	ScoreModeStatic ScoreMode = iota
	// The score is static
	ScoreModeDynamic
	// The score is caculated by the number of total successful submissions
	ScoreModeOrdered
	// The score is caculated by the order of the submission
)

type Challenge struct {
	gorm.Model `json:"-"`

	// Name of the challenge
	Name string `json:"name"`

	// Description of the challenge
	// Markdown supported
	Description string `gorm:"type:text" json:"description"`

	// UUID of the challenge
	UUID string `gorm:"unique" json:"uuid"`

	// Game of the challenge
	GameID uint  `json:"-"`
	Game   *Game `gorm:"foreignKey:GameID" json:"game"`

	// State of the challenge
	State ChallengeState `gorm:"default:0" json:"state"`

	// Creators of the challenge
	CreatorID uint  `json:"-"`
	Creator   *User `gorm:"foreignKey:CreatorID" json:"creator"`

	// Attachments of the challenge
	Attachments []*Attachment `gorm:"many2many:challenge_attachments;" json:"attachments"`

	// Tags of the challenge
	Tags types.StringArray `gorm:"type:text" json:"tags"`

	// Category of the challenge
	Category string `json:"category"`

	// StartTime of the challenge
	StartTime int64 `gorm:"default:0" json:"start_time"`

	// ExpireTime of the challenge
	ExpireTime int64 `gorm:"default:0" json:"expire_time"`

	// Is this challenge available (to create container, submit flag etc.) after the deadline
	AfterExpiredOperations AfterExpireOp `gorm:"default:0" json:"after_expired_operations"`

	// Image of the challenge
	ImageID uint   `json:"-"`
	Image   *Image `gorm:"foreignKey:ImageID" json:"image"`

	// Exposed port of the challenge
	ExposedPort int `gorm:"default:0" json:"exposed_port"`

	// Does the challenge need a container
	NoContainer bool `gorm:"default:false" json:"no_container"`

	// Dynamic flag or not
	DynamicFlag bool `gorm:"default:false" json:"dynamic_flag" priv:"2"`

	// Flag template of the challenge
	Flag string `json:"flag" priv:"2"`

	// Score of the challenge (if not dynamic)
	Score int `gorm:"default:0" json:"score"`

	// Score mode of the challenge
	ScoreMode ScoreMode `gorm:"default:0" json:"score_mode"`

	// Score caculation formula of the challenge
	// PS: The score will automatically rounded
	//
	// Parameters:
	// original_score, solved_count, solved_order
	ScoreFormula string `gorm:"type:text" json:"score_formula" priv:"2"`

	// Fake flags of the challenge
	// Set for anti-cheat
	FakeFlag types.StringArray `gorm:"type:text" json:"fake_flag" priv:"2"`

	// Hints of the challenge
	// Markdown supported
	Hints types.StringArray `gorm:"type:text" json:"hints"`
}

func (s *Store) CreateChallenge(challenge *Challenge) error {
	return s.db.Create(challenge).Error
}

func (s *Store) GetChallenges() ([]*Challenge, error) {
	var challenges []*Challenge
	err := s.db.Preload("Game").Preload("Creator").Preload("Attachments").Find(&challenges).Error
	return challenges, err
}

func (s *Store) GetChallengeByID(id int) (*Challenge, error) {
	var challenge Challenge
	err := s.db.First(&challenge, id).Error
	return &challenge, err
}

func (s *Store) GetChallengeByUUID(uuid string) (*Challenge, error) {
	var challenge Challenge
	err := s.db.Preload("Game").Preload("Creator").Preload("Attachments").Where("uuid = ?", uuid).First(&challenge).Error
	return &challenge, err
}

func (c *Challenge) Expired() bool {
	return c.ExpireTime != 0 && c.ExpireTime < time.Now().Unix()
}
