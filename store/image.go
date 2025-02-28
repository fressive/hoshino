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

type Image struct {
	gorm.Model `json:"-"` // ID, CreatedAt, UpdatedAt, DeletedAt

	// Name of the image
	Name string `json:"name"`

	MemoryLimit  string `json:"memory_limit" gorm:"default:'128Mi'"`
	CPULimit     string `json:"cpu_limit" gorm:"default:'100m'"`
	StorageLimit string `json:"storage_limit" gorm:"default:'1Gi'"`

	ExposedPort int `json:"exposed_port"`

	RegistryAccessTokenUUID string `json:"registry_access_token_uuid"`
}
