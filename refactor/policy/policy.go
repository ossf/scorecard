// Copyright 2022 Allstar Authors

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package policy

import (
	"fmt"
	"os"

	sce "github.com/ossf/scorecard/v4/errors"
	spol "github.com/ossf/scorecard/v4/policy"
)

func FromFile(policyFile string) (*spol.ScorecardPolicy, error) {
	if policyFile != "" {
		data, err := os.ReadFile(policyFile)
		if err != nil {
			return nil, sce.WithMessage(sce.ErrScorecardInternal,
				fmt.Sprintf("os.ReadFile: %v", err))
		}

		sp, err := spol.ParseFromYAML(data)
		if err != nil {
			return nil,
				sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("spol.ParseFromYAML: %v", err))
		}

		return sp, nil
	}

	return nil, nil
}
