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

package data

import "strings"

type csvStrings []string

func (s csvStrings) MarshalCSV() ([]byte, error) {
	return []byte(strings.Join(s, ",")), nil
}

func (s *csvStrings) UnmarshalCSV(input []byte) error {
	if len(input) == 0 || string(input) == "" {
		*s = nil
		return nil
	}
	*s = strings.Split(string(input), ",")
	return nil
}

type repoFormat struct {
	Repo     string     `csv:"repo"`
	Metadata csvStrings `csv:"metadata"`
}
