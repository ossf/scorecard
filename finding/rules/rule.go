// Copyright 2022 OpenSSF Scorecard Authors
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

package rule

import (
	"embed"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/ossf/scorecard/v4/checker"
	sce "github.com/ossf/scorecard/v4/errors"
)

type jsonRemediation struct {
	Text     []string                  `yaml:"text"`
	Markdown []string                  `yaml:"markdown"`
	Effort   checker.RemediationEffort `yaml:"effort"`
}

type jsonRule struct {
	Short          string          `yaml:"short"`
	Desc           string          `yaml:"desc"`
	Motivation     string          `yaml:"motivation"`
	Implementation string          `yaml:"implementation"`
	Risk           string          `yaml:"risk"`
	Remediation    jsonRemediation `yaml:"remediation"`
}

type Rule struct {
	Name        string
	Short       string
	Desc        string
	Motivation  string
	Risk        checker.Risk
	Remediation *checker.Remediation
}

func RuleNew(loc embed.FS, rule string) (*Rule, error) {
	fmt.Println(os.Getwd())
	content, err := loc.ReadFile(fmt.Sprintf("%s.yml", rule))
	if err != nil {
		return nil, sce.WithMessage(sce.ErrScorecardInternal, err.Error())
	}
	fmt.Print(string(content))

	r, err := parseFromJSON(content)
	if err != nil {
		return nil, err
	}
	// TODO: validate
	return &Rule{
		Name:       rule,
		Short:      r.Short,
		Desc:       r.Desc,
		Motivation: r.Motivation,
		Risk:       checker.Risk(r.Risk),
		Remediation: &checker.Remediation{
			Text:         strings.Join(r.Remediation.Text, "\n"),
			TextMarkdown: strings.Join(r.Remediation.Markdown, "\n"),
			Effort:       r.Remediation.Effort,
		},
	}, nil
}

func parseFromJSON(content []byte) (*jsonRule, error) {
	r := jsonRule{}

	err := yaml.Unmarshal(content, &r)
	if err != nil {
		return nil, sce.WithMessage(sce.ErrScorecardInternal, err.Error())
	}
	return &r, nil
}

// https://abhinavg.net/2021/02/24/flexible-yaml/
// func (r *risk) UnmarshalYAML(n *yaml.Node) error {
// 	var str string
// 	if err := n.Decode(&str); err != nil {
// 		return fmt.Errorf("%w", err)
// 	}

// 	switch n.Value {
// 	default:
// 		return fmt.Errorf("%w: %q", errors.New("bla"), str)
// 	}
// 	return nil
// }
