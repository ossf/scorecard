// Copyright 2023 OpenSSF Scorecard Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package fnmatch

import (
	"fmt"
	"regexp"
	"strings"
)

func Match(pattern, path string) (bool, error) {
	r := convertToRegex(pattern)
	m, err := regexp.MatchString(r, path)
	if err != nil {
		return false, fmt.Errorf("converted regex invalid: %w", err)
	}
	return m, nil
}

var specialRegexpChars = map[byte]struct{}{
	'.':  {},
	'+':  {},
	'*':  {},
	'?':  {},
	'^':  {},
	'$':  {},
	'(':  {},
	')':  {},
	'[':  {},
	']':  {},
	'{':  {},
	'}':  {},
	'|':  {},
	'\\': {},
}

func convertToRegex(pattern string) string {
	var regexPattern strings.Builder
	regexPattern.WriteRune('^')
	for len(pattern) > 0 {
		matchLen := 1
		switch {
		case len(pattern) > 2 && pattern[:3] == "**/":
			// Matches directories recursively
			regexPattern.WriteString("(?:[^/]+/?)+")
			matchLen = 3
		case len(pattern) > 1 && pattern[:2] == "**":
			// Matches files expansively.
			regexPattern.WriteString("[^/]+")
			matchLen = 2
		case len(pattern) > 1 && pattern[:1] == "\\":
			writePotentialRegexpChar(&regexPattern, pattern[1])
			matchLen = 2
		default:
			switch pattern[0] {
			case '*':
				// Equivalent to ".*"" in regexp, but GitHub uses the File::FNM_PATHNAME flag for the File.fnmatch syntax
				// the * wildcard does not match directory separators (/).
				regexPattern.WriteString("[^/]*")
			case '?':
				// Matches any one character. Equivalent to ".{1}" in regexp, see FNM_PATHNAME note above.
				regexPattern.WriteString("[^/]{1}")
			case '[', ']':
				// "[" and "]" represent character sets in fnmatch too
				regexPattern.WriteByte(pattern[0])
			default:
				writePotentialRegexpChar(&regexPattern, pattern[0])
			}
		}
		pattern = pattern[matchLen:]
	}
	regexPattern.WriteRune('$')
	return regexPattern.String()
}

// Characters with special meaning in regexp may need escaped.
func writePotentialRegexpChar(sb *strings.Builder, b byte) {
	if _, ok := specialRegexpChars[b]; ok {
		sb.WriteRune('\\')
	}
	sb.WriteByte(b)
}
