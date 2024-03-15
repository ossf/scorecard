#!/bin/bash

set -e

if [[ "$#" != 1 ]]; then
  echo "Specify name of the probe: setup-probe.sh [nameOfProbe]"
  exit
fi

export PROBE_NAME=$1

mkdir probes/"$1"
cd probes/"$1"
touch def.yml impl_test.go impl.go

cat << EOF > impl.go
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

//nolint:stylecheck
package $1

import (
	"embed"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/finding"
)

//go:embed *.yml
var fs embed.FS

const (
	Probe = "$1"
)

func Run(raw *checker.RawResults) ([]finding.Finding, string, error) {
	return []finding.Finding{}, "", nil
}

EOF


cat << EOF > impl_test.go
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

//nolint:stylecheck
package $1
EOF

cat << EOF > def.yml
# Copyright 2024 OpenSSF Scorecard Authors
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

id: $1
short: A short description of this probe
motivation: >
    What is the motivation for this probe?
implementation: >
    How does this probe work under-the-hood?
outcome:
  -
remediation:
  effort: # High, Medium, Low
  text:
  - 
ecosystem:
  languages:
    - all
  clients:
    - github
    - gitlab
EOF