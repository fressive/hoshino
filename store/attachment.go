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

type Attachment struct {
	gorm.Model
	UUID string `json:"uuid"`
	Name string `json:"name"`

	// If the attachment is multiple, we will distribute it depending on the user's hash
	Multiple     bool   `json:"multiple" priv:"2"`
	SavePath     string `json:"save_path" priv:"2"`
	DownloadName string `json:"download_name"`

	// If the attachment is a flag, we will store the flag
	Flag string `json:"flag" priv:"2"`

	ChallengeID uint       `json:"challenge_id" priv:"2"`
	Challenge   *Challenge `json:"challenge" gorm:"foreignKey:ChallengeID" priv:"2"`

	UploaderID uint  `json:"uploader_id" priv:"2"`
	Uploader   *User `json:"uploader" gorm:"foreignKey:UploaderID" priv:"2"`
}

func (s *Store) CreateAttachment(a *Attachment) error {
	return s.db.Create(a).Error
}

func (s *Store) GetAttachmentByUUID(uuid string) (*Attachment, error) {
	var attachment Attachment
	err := s.db.Where("uuid = ?", uuid).First(&attachment).Error
	return &attachment, err
}

func (s *Store) GetAttachmentsByChallenge(challenge *Challenge) ([]Attachment, error) {
	var attachments []Attachment
	err := s.db.Where("challenge_id = ?", challenge.ID).Find(&attachments).Error
	return attachments, err
}

func (s *Store) GetAttachmentsByName(name string) ([]Attachment, error) {
	var attachments []Attachment
	err := s.db.Where("name = ?", name).Find(&attachments).Error
	return attachments, err
}

func (s *Store) DeleteAttachment(attachment *Attachment) error {
	return s.db.Delete(attachment).Error
}
