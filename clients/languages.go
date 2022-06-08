// Copyright 2021 Security Scorecard Authors
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

// A customized string type `Language` for languages used by clients.
// A language could be a programming language, or more general,
// such as Dockerfile, CMake, HTML, YAML, etc.
type Language string

// TODO: retrieve all languages supported by GitHub.
const (
	// Go: https://go.dev/
	Go Language = "go"

	// Python: https://www.python.org/
	Python Language = "python"

	// JavaScript: https://www.javascript.com/
	JavaScript Language = "javascript"

	// C++: https://cplusplus.com/
	Cpp Language = "c++"

	// C: https://www.open-std.org/jtc1/sc22/wg14/
	C Language = "c"

	// TypeScript: https://www.typescriptlang.org/
	TypeScript Language = "typescript"

	// Java: https://www.java.com/en/
	Java Language = "java"

	// C#: https://docs.microsoft.com/en-us/dotnet/csharp/
	CSharp Language = "c#"

	// Ruby: https://www.ruby-lang.org/
	Ruby Language = "ruby"

	// PHP: https://www.php.net/
	PHP Language = "php"

	// Starlark: https://github.com/bazelbuild/starlark
	StarLark Language = "starlark"

	// Scala: https://www.scala-lang.org/
	Scala Language = "scala"

	// Kotlin: https://kotlinlang.org/
	Kotlin Language = "kotlin"

	// Swift: https://github.com/apple/swift
	Swift Language = "swift"

	// Rust: https://github.com/rust-lang/rust
	Rust Language = "rust"

	// Other indicates other programming languages not listed by the GitHub API.
	Other Language = "other"

	// Add more programming languages here if needed, please use lower cases.
)
