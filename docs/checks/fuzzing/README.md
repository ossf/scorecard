# Supported Fuzzers

* [LibFuzzer](https://llvm.org/docs/LibFuzzer.html)
  * Detection is based on usages of a function named `LLVMFuzzerTestOneInput` in C, C++, or Swift files.
* [ClusterFuzzLite](https://github.com/google/clusterfuzzlite)
  * Detection is based on a file called `.clusterfuzzlite/Dockerfile`.
* [Native Go Fuzzing](https://go.dev/doc/security/fuzz/)
  * Looks for functions of the form `func FuzzXxx(*testing.F)` in Go files.
* [Jazzer](https://github.com/CodeIntelligenceTesting/jazzer)
  * Detection based on the import of `com.code_intelligence.jazzer.api.FuzzedDataProvider` in Java files.
* [OSS-Fuzz](https://github.com/google/oss-fuzz)
  * Detection based on the presence of integrated projects in the [google/oss-fuzz GitHub repo](https://github.com/google/oss-fuzz/tree/master/projects).
* Property-based Haskell Fuzzers
  * Detected based on imports of various testing frameworks:
    * [QuickCheck](https://hackage.haskell.org/package/QuickCheck)
    * [hedgehog]( https://hedgehog.qa/)
    * [validity](https://github.com/NorfairKing/validity)
    * [smallcheck](https://hackage.haskell.org/package/smallcheck)
    * [hspec](https://hspec.github.io/)
    * [tasty](https://hackage.haskell.org/package/tasty)
* [fast-check](https://github.com/dubzzz/fast-check)
  * Detection based on import statements in JavaScript and TypeScript files.
* [Atheris](https://github.com/google/atheris)
  * Detection based on the presence of `import atheris` in Python files.
* [cargo-fuzz](https://rust-fuzz.github.io/book/cargo-fuzz.html)
  * Detection based on presence of `libfuzzer_sys` in Rust files.
* [FsCheck](https://github.com/fscheck/FsCheck)
  * Detection based on import statements in C# and F# files.

## Add Support

Don't see your fuzzing tool listed?
Search for an existing issue, or create one, to discuss adding support.
