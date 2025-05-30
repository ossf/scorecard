// Copyright 2024 OpenSSF Scorecard Authors
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
package main

import (
	"fmt"
	"os"
	"path"
	"time"
)

func main() {
	if len(os.Args) != 2 {
		panic("usage: make setup-probe probeName=arg")
	}
	pn := os.Args[1]

	pd := path.Join("probes", pn)
	err := os.Mkdir(pd, 0o700)
	if err != nil {
		panic(err)
	}

	y := time.Now().Year()

	implBoilerplate := fmt.Sprintf(`// Copyright %d OpenSSF Scorecard Authors
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

package %s

import (
	"embed"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/finding"
)

//go:embed *.yml
var fs embed.FS

const (
	Probe = "%s"
)

// If your probe is associated with a Scorecard check, map it like so:
// and create the entry in probes/entries.go
// func init() {
// 	probes.MustRegister(Probe, Run, []probes.CheckName{probes.<ProbeName>})
// }
// If your probe isn't supposed to be part of a Scorecard check, you must
// register it independently:
// func init() {
// 	probes.MustRegisterIndependent(Probe, Run)
// }

func Run(raw *checker.RawResults) ([]finding.Finding, string, error) {
	return nil, "", nil
}
`, y, pn, pn)
	err = os.WriteFile(path.Join(pd, "impl.go"), []byte(implBoilerplate), 0o600)
	if err != nil {
		panic(err)
	}

	implTestBoilerplate := fmt.Sprintf(`// Copyright %d OpenSSF Scorecard Authors
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

package %s
`, y, pn)
	err = os.WriteFile(path.Join(pd, "impl_test.go"), []byte(implTestBoilerplate), 0o600)
	if err != nil {
		panic(err)
	}

	defYmlBoilerplate := fmt.Sprintf(`# Copyright %d OpenSSF Scorecard Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

id: %s
short: A short description of this probe
motivation: >
  What is the motivation for this probe?
implementation: >
  How does this probe work under-the-hood?
outcome:
  -
remediation:
  onOutcome: # Which direction of a probe outcome (True/False) requires a "fix"?
  effort: # High, Medium, Low
  text:
  - 
ecosystem:
  languages:
  -
  clients:
  -
`, y, pn)
	err = os.WriteFile(path.Join(pd, "def.yml"), []byte(defYmlBoilerplate), 0o600)
	if err != nil {
		panic(err)
	}
}
