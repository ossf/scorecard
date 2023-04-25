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

	// C#: https://docs.microsoft.com/en-us/dotnet/csharp/
	CSharp LanguageName = "c#"

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

	// Haskell: https://www.haskell.org/
	Haskell LanguageName = "haskell"

	// Other indicates other languages not listed by the GitHub API.
	Other LanguageName = "other"

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
