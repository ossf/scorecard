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

package data

import "strings"

// CSVStrings is []string with support for CSV formatting.
type CSVStrings []string

// MarshalCSV implements []string -> []byte serialization.
func (s CSVStrings) MarshalCSV() ([]byte, error) {
	return []byte(strings.Join(s, ",")), nil
}

// UnmarshalCSV implements []byte -> []string de-serializtion.
func (s *CSVStrings) UnmarshalCSV(input []byte) error {
	if len(input) == 0 {
		*s = nil
		return nil
	}
	*s = strings.Split(string(input), ",")
	return nil
}

// ToString converts CSVStrings -> []string.
func (s CSVStrings) ToString() []string {
	var ret []string
	for _, i := range s {
		ret = append(ret, i)
	}
	return ret
}

// RepoFormat is used to read input repos.
type RepoFormat struct {
	Repo     string     `csv:"repo"`
	Metadata CSVStrings `csv:"metadata"`
}
