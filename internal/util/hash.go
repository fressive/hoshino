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
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cast"
	"golang.org/x/exp/rand"
)

func UUID() string {
	return uuid.New().String()
}

func SHA256(secret string) string {
	h := sha256.New()
	h.Write([]byte(secret))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func SHA256WithSalt(secret string, salt string) string {
	return SHA256(secret + salt)
}

func GenerateRandomText(length int32) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, length)
	rand.Seed(cast.ToUint64(time.Now().UnixNano()))
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func GenerateRandomNumbers(length int32) string {
	const numberBytes = "0123456789"
	b := make([]byte, length)

	rand.Seed(cast.ToUint64(time.Now().UnixNano()))
	for i := range b {
		b[i] = numberBytes[rand.Intn(len(numberBytes))]
	}
	return string(b)
}
