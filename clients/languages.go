// Copyright 2021 OpenSSF Scorecard Authors
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

package clients

import "strings"

// LanguageName is the name of a language, a customized type of string.
type LanguageName string

// TODO: retrieve all languages supported by GitHub, or add one manually if needed.
// Currently, this is still an incomplete list of languages.
const (
	// Go: https://go.dev/
	Go LanguageName = "go"

	// Python: https://www.python.org/
	Python LanguageName = "python"

	// JavaScript: https://www.javascript.com/
	JavaScript LanguageName = "javascript"

	// C++: https://cplusplus.com/
	Cpp LanguageName = "c++"

	// C: https://www.open-std.org/jtc1/sc22/wg14/
	C LanguageName = "c"

	// TypeScript: https://www.typescriptlang.org/
	TypeScript LanguageName = "typescript"

	// Java: https://www.java.com/en/
	Java LanguageName = "java"

	// C#: https://docs.microsoft.com/dotnet/csharp/
	CSharp LanguageName = "c#"

	// ObjectiveC: the objective c language.
	ObjectiveC LanguageName = "objectivec"

	// ObjectiveCpp: the objective c++ language.
	ObjectiveCpp LanguageName = "objective-c++"

	// Ruby: https://www.ruby-lang.org/
	Ruby LanguageName = "ruby"

	// PHP: https://www.php.net/
	PHP LanguageName = "php"

	// Starlark: https://github.com/bazelbuild/starlark
	StarLark LanguageName = "starlark"

	// Scala: https://www.scala-lang.org/
	Scala LanguageName = "scala"

	// Kotlin: https://kotlinlang.org/
	Kotlin LanguageName = "kotlin"

	// Swift: https://github.com/apple/swift
	Swift LanguageName = "swift"

	// Rust: https://github.com/rust-lang/rust
	Rust LanguageName = "rust"

	// CMake: https://cmake.org/
	CMake LanguageName = "cmake"

	// Dockerfile: https://docs.docker.com/engine/reference/builder/
	Dockerfile LanguageName = "dockerfile"

	// Erlang: https://www.erlang.org/
	Erlang LanguageName = "erlang"

	// Haskell: https://www.haskell.org/
	Haskell LanguageName = "haskell"

	// Elixir: https://www.elixir.org/
	Elixir LanguageName = "elixir"

	// Gleam: https://www.gleam.org/
	Gleam LanguageName = "gleam"

	// Other indicates other languages not listed by the GitHub API.
	Other LanguageName = "other"

	// All indicates all programming languages.
	All LanguageName = "all"

	// F#: https://learn.microsoft.com/dotnet/fsharp/
	FSharp LanguageName = "f#"

	// Add more languages here if needed,
	// please use lowercases for the LanguageName value.
)

// Language represents a customized struct for languages used by clients.
// A language could be a programming language, or more general,
// such as Dockerfile, CMake, HTML, YAML, etc.
type Language struct {
	// Name is the name of this language.
	Name LanguageName

	// NumLines is the total number of code lines of this language in the repo.
	NumLines int

	// TODO: add more properties for Language.
}

// HasMemoryUnsafeLanguage returns true if any of the languages in the list are memory-unsafe.
// Based on OpenSSF best practices:
// https://github.com/ossf/scorecard/blob/main/docs/checks.md#fuzzing
func HasMemoryUnsafeLanguage(langs []Language) bool {
	memoryUnsafeLanguages := map[LanguageName]bool{
		C:             true,
		Cpp:           true,
		ObjectiveC:    true,
		ObjectiveCpp:  true,
		"objective-c": true, // Robustness for hyphenated Objective-C
	}

	for _, lang := range langs {
		normalizedName := LanguageName(strings.ToLower(string(lang.Name)))
		if memoryUnsafeLanguages[normalizedName] {
			return true
		}
	}
	return false
}
