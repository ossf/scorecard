package foo

import "strings"

func SafeFooImports(input string) {
	_ = strings.Contains(input, "foo")
}
