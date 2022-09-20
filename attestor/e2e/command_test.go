// Copyright 2022 Security Scorecard Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package e2e

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/ossf/scorecard-attestor/command"
)

func execute(t *testing.T, c *cobra.Command, args ...string) (string, error) {
	t.Helper()

	buf := new(bytes.Buffer)
	c.SetOut(buf)
	c.SetErr(buf)
	c.SetArgs(args)

	err := c.Execute()
	return strings.TrimSpace(buf.String()), err
}

func TestRootCmd(t *testing.T) {
	t.Parallel()
	tt := []struct {
		name string
		args []string
	}{
		{
			name: "test check-only from root",
			args: []string{
				"verify",
				"--policy=../policy/testdata/policy-binauthz.yaml",
				"--repo-url=https://github.com/ossf-tests/scorecard",
			},
		},
	}

	for _, tc := range tt {
		_, err := execute(t, command.RootCmd, tc.args...)
		if err != nil {
			t.Fatalf("%s: %s", tc.name, err)
		}
	}
}
