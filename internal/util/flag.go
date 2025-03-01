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
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/exp/rand"
)

var leetMap = map[rune][]rune{
	'A': {'4', 'A', 'a'},
	'B': {'8', 'B', 'b'},
	'C': {'C', 'c'},
	'D': {'D', 'd'},
	'E': {'3', 'E', 'e'},
	'F': {'F', 'f'},
	'G': {'6', 'G', 'g'},
	'H': {'H', 'h'},
	'I': {'I', 'i', '1'},
	'J': {'J', 'j'},
	'K': {'K', 'k'},
	'L': {'L', 'l', '1'},
	'M': {'M', 'm'},
	'N': {'N', 'n'},
	'O': {'O', 'o', '0'},
	'P': {'P', 'p'},
	'Q': {'Q', 'q'},
	'R': {'R', 'r'},
	'S': {'S', 's', '5', '$'},
	'T': {'T', 't', '7'},
	'U': {'U', 'u'},
	'V': {'V', 'v'},
	'W': {'W', 'w'},
	'X': {'X', 'x'},
	'Y': {'Y', 'y'},
	'Z': {'Z', 'z', '2'},
}

func Leetify(input string, seed uint64) string {
	rand.Seed(seed)

	var result strings.Builder
	for _, char := range input {
		if leetChars, ok := leetMap[unicode.ToUpper(char)]; ok {
			result.WriteRune(leetChars[rand.Intn(len(leetChars))])
		} else {
			result.WriteRune(char)
		}
	}

	return result.String()
}

func GenerateFlagContent(format string, seed string) string {
	// Use [] to wrap the string to be leetified
	// Use {hash} to insert the sha256ed seed
	// Use {uuid} to insert a random uuid

	leetRe := regexp.MustCompile(`\[([^\[\]]+)\]`)

	result := format
	result = strings.ReplaceAll(result, "{hash}", SHA256(seed)[:8])
	result = strings.ReplaceAll(result, "{uuid}", UUID())

	leet := leetRe.FindAllString(result, -1)
	for _, l := range leet {
		result = strings.ReplaceAll(result, l, Leetify(l[1:len(l)-1], SHA256Uint64(seed)))
	}

	return result
}
