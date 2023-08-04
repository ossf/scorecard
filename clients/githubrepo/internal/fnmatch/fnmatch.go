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
