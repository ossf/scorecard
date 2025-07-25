# METADATA
# title: No-Binary-Files
# ingest: git
# eval: constraints
# description: The repository does not contain binary files.

package minder

import rego.v1

suspicious_extensions := {
  ".crx",
  ".deb",
  ".dex",
  ".dey",
  ".elf",
  ".o",
  ".a",
  ".so",
  ".macho",
  ".iso",
  ".class",
  ".jar",
  ".bundl",
  ".dylib",
  ".lib",
  ".msi",
  ".dll",
  ".drv",
  ".efi",
  ".exe",
  ".ocx",
  ".pyc",
  ".pyo",
  ".par",
  ".rpm",
  ".wasm",
  ".whl",
}

gradleValidationActions := {
  "gradle/wrapper-validation-action@",
  "gradle/actions/wrapper-validation@",
}

files_in_repo := file.walk(".")

violations contains {"msg": msg} if {
  some current_file in files_in_repo

  http_type := file.http_type(current_file)
  http_type == "application/octet-stream"
  strings.any_suffix_match(current_file, suspicious_extensions)
  not gradleValidatedOk(current_file)

  msg := sprintf("Binary artifact found: %s", [current_file])
}

# Gradle wrapper validation, see ../raw/binary_artifact.go#189
gradleValidatedOk(filename) if {
  strings.any_suffix_match(filename, "gradle-wrapper.jar")

  getValidatingWorkflow(filename)
}

# N.B. this currently does not check that the filename is validated.  /shrug
getValidatingWorkflow(_) := workflowFile if {
  some workflowFile in file.ls("./.github/workflows")

  workflow := yaml.unmarshal(file.read(workflowFile))

  workflow.job[_].steps[_].uses in gradleValidationActions
}