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

package yaml

type Remediation struct {
	OnOutcome string   `yaml:"onOutcome"`
	Effort    string   `yaml:"effort"`
	Text      []string `yaml:"text"`
	Markdown  []string `yaml:"markdown"`
}

type Ecosystem struct {
	Languages []string `yaml:"languages"`
	Clients   []string `yaml:"clients"`
}

type Probe struct {
	Remediation    Remediation `yaml:"remediation"`
	ID             string      `yaml:"id"`
	Short          string      `yaml:"short"`
	Motivation     string      `yaml:"motivation"`
	Lifecycle      string      `yaml:"lifecycle"`
	Implementation string      `yaml:"implementation"`
	Ecosystem      Ecosystem   `yaml:"ecosystem"`
	Outcomes       []string    `yaml:"outcome"`
}
