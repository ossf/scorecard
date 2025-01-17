// Copyright 2024 OpenSSF Scorecard Authors
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

// Package fuzzers defines the string constants used when identifying supported fuzzer tools.
package fuzzers

const (
	OSSFuzz                 = "OSSFuzz"
	ClusterFuzzLite         = "ClusterFuzzLite"
	BuiltInGo               = "GoBuiltInFuzzer"
	PropertyBasedErlang     = "ErlangPropertyBasedTesting"
	PropertyBasedHaskell    = "HaskellPropertyBasedTesting"
	PropertyBasedElixir     = "ElixirPropertyBasedTesting"
	PropertyBasedGleam      = "GleamPropertyBasedTesting"
	PropertyBasedJavaScript = "JavaScriptPropertyBasedTesting"
	PropertyBasedTypeScript = "TypeScriptPropertyBasedTesting"
	PythonAtheris           = "PythonAtherisFuzzer"
	CLibFuzzer              = "CLibFuzzer"
	CppLibFuzzer            = "CppLibFuzzer"
	SwiftLibFuzzer          = "SwiftLibFuzzer"
	RustCargoFuzz           = "RustCargoFuzzer"
	JavaJazzerFuzzer        = "JavaJazzerFuzzer"
	// TODO: add more fuzzing check supports.
)
