#!/bin/bash

tmp=$(mktemp -d)
trap "rm -rf $tmp" EXIT
git clone git@github.com:google/oss-fuzz.git --depth=1 $tmp
cat $tmp/projects/*/Dockerfile | grep "git clone" | grep -o "github.com/\S*" | sort | uniq > repos.txt


echo "package checks" > ossfuzz.go
echo "// GENERATED CODE, DO NOT EDIT" >> ossfuzz.go
echo "var fuzzRepos=\`" >> ossfuzz.go
cat repos.txt >> ossfuzz.go
echo "\`" >> ossfuzz.go
