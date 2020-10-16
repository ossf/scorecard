#!/bin/bash

tmp=$(mktemp -d)
trap "rm -rf $tmp" EXIT
git clone https://github.com/google/oss-fuzz --depth=1 $tmp
cat $tmp/projects/*/Dockerfile | grep "git clone" | grep -o "github.com/\S*" | sort | uniq > $tmp/repos.txt

ossfuzz_file=checks/ossfuzz.go
echo "package checks" > $ossfuzz_file
echo "// GENERATED CODE, DO NOT EDIT" >> $ossfuzz_file
echo "var fuzzRepos=\`" >> $ossfuzz_file
cat $tmp/repos.txt >> $ossfuzz_file
echo "\`" >> $ossfuzz_file
