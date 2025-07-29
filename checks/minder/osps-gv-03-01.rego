# METADATA
# title: Contributing-File
# ingest: git
# eval: deny-by-default
# description: Contributor guidance found in a CONTRIBUTING file

package minder

import rego.v1

default allow := false
# We set a floor on the score when using a git forge
default output.score := 3
default message := "No CONTRIBUTING file found"

allow if {
  files := file.ls_glob("./CONTRIBUTING*")

  some name
  content := file.read(files[name])
  "" != content
  output.score = 10
}

allow if {
  files := file.ls_glob("./CONTRIBUTING/*")

  some name
  content := file.read(files[name])
  "" != content
  output.score = 10
}