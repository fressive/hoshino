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

func TestFlag(t *testing.T) {
	flag1 := GenerateFlagContent("takanashi_hoshino_is_super_kawaii", "random_seed")
	assert.Equal(t, "takanashi_hoshino_is_super_kawaii", flag1, "they should be equal")

	flag2 := GenerateFlagContent("[takanashi_hoshino_is_super_kawaii]_[leet2]", "random_seed")
	println(flag2)
	assert.NotEqual(t, "takanashi_hoshino_is_super_kawaii_leet2", flag2, "they should not be equal")

	flag3 := GenerateFlagContent("[takanashi_hoshino_is_super_kawaii]_[leet2]", "random_seed")
	flag6 := GenerateFlagContent("[takanashi_hoshino_is_super_kawaii]_[leet2]", "random_seeeeeeed")
	println(flag3, flag6)
	assert.Equal(t, flag2, flag3, "they should be equal")
	assert.NotEqual(t, flag3, flag6, "they should not be equal")

	flag4 := GenerateFlagContent("[takanashi_hoshino_is_super_kawaii]_{seed}", "random_seed")
	flag5 := GenerateFlagContent("[takanashi_hoshino_is_super_kawaii_{seed}]", "random_seed")
	println(flag4, flag5)
	assert.NotEqual(t, flag4, flag5, "they should not be equal")

}
