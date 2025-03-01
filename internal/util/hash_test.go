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

package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSHA256(t *testing.T) {
	secret := "takanashi_hoshino_is_super_kawaii"
	expectedHash := "0a649edaea14f0f0b9f40174f98e51543d5382a58918647ed08dee8dc4ee2a4a"
	hash := SHA256(secret)
	assert.Equal(t, expectedHash, hash, "they should be equal")
}

func TestSHA256WithSalt(t *testing.T) {
	secret := "takanashi_hoshino_is_super_kawaii"
	salt := "random_salt"
	expectedHash := SHA256(secret + salt)
	hash := SHA256WithSalt(secret, salt)
	assert.Equal(t, expectedHash, hash, "they should be equal")
}

func TestGenerateRandomText(t *testing.T) {
	length := int32(10)
	randomText := GenerateRandomText(length)
	assert.Equal(t, length, int32(len(randomText)), "length should be equal to the specified length")
}

func TestSHA256Uint64(t *testing.T) {
	seedToInt := SHA256Uint64("takanashi_hoshino_is_super_kawaii")
	assert.Equal(t, uint64(0xa649edaea14f0f0), seedToInt, "they should be equal")
	seedToInt2 := SHA256Uint64("takanashi_hoshino_is_super_kawaii222")
	assert.NotEqual(t, seedToInt, seedToInt2, "they should not be equal")
}
