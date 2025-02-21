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
	"gorm.io/gorm"
)

type EventType int

const (
	GameEventTypeNormal EventType = iota
	GameEventTypeChallengeSolved
	GameEventTypeCheatDetected
)

// Events during the game
type GameEvent struct {
	gorm.Model

	// Content of this event
	Content string `gorm:"type:text" json:"content"`

	// Game of this event
	Game *Game `gorm:"foreignKey:id" json:"game"`

	// Challenge of this event
	Challenge *Challenge `gorm:"foreignKey:id" json:"challenge"`

	// Related teams of this event
	RelatedTeams []*Team `gorm:"many2many:event_teams;" json:"related_teams"`

	// Visibility to the normal players
	Visibility bool `gorm:"default:true" json:"visibility"`

	// Type of the event
	Type EventType `gorm:"default:0" json:"type"`
}

func (s *Store) CreateGameEvent(event *GameEvent) error {
	// TODO: use websocket to broadcast the event
	// TODO: use webhook to notify the event
	return s.db.Create(event).Error
}

func (s *Store) UpdateGameEvent(event *GameEvent) error {
	return s.db.Save(event).Error
}

func (s *Store) DeleteGameEvent(event *GameEvent) error {
	return s.db.Delete(event).Error
}

func (s *Store) GetAllGameEvents(visibility bool) ([]GameEvent, error) {
	var events []GameEvent
	err := s.db.Where("visibility = ?", visibility).Find(&events).Error
	return events, err
}
