// Copyright 2023 OpenSSF Scorecard Authors
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
package evaluation

import (
	"testing"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/clients"
)

func TestCIIBestPractices(t *testing.T) {
	t.Run("CIIBestPractices", func(t *testing.T) {
		t.Run("in progress", func(t *testing.T) {
			r := &checker.CIIBestPracticesData{
				Badge: clients.InProgress,
			}
			result := CIIBestPractices("CIIBestPractices", nil, r)
			if result.Score != inProgressScore {
				t.Errorf("CIIBestPractices() = %v, want %v", result.Score, inProgressScore)
			}
		})
		t.Run("passing", func(t *testing.T) {
			r := &checker.CIIBestPracticesData{
				Badge: clients.Passing,
			}
			result := CIIBestPractices("CIIBestPractices", nil, r)
			if result.Score != passingScore {
				t.Errorf("CIIBestPractices() = %v, want %v", result.Score, passingScore)
			}
		})
		t.Run("silver", func(t *testing.T) {
			r := &checker.CIIBestPracticesData{
				Badge: clients.Silver,
			}
			result := CIIBestPractices("CIIBestPractices", nil, r)
			if result.Score != silverScore {
				t.Errorf("CIIBestPractices() = %v, want %v", result.Score, silverScore)
			}
		})
		t.Run("gold", func(t *testing.T) {
			r := &checker.CIIBestPracticesData{
				Badge: clients.Gold,
			}
			result := CIIBestPractices("CIIBestPractices", nil, r)
			if result.Score != checker.MaxResultScore {
				t.Errorf("CIIBestPractices() = %v, want %v", result.Score, checker.MaxResultScore)
			}
		})
		t.Run("not found", func(t *testing.T) {
			r := &checker.CIIBestPracticesData{
				Badge: clients.NotFound,
			}
			result := CIIBestPractices("CIIBestPractices", nil, r)
			if result.Score != checker.MinResultScore {
				t.Errorf("CIIBestPractices() = %v, want %v", result.Score, checker.MinResultScore)
			}
		})
		t.Run("error", func(t *testing.T) {
			r := &checker.CIIBestPracticesData{
				Badge: clients.Unknown,
			}
			result := CIIBestPractices("CIIBestPractices", nil, r)
			if result.Score != -1 {
				t.Errorf("CIIBestPractices() = %v, want %v", result.Score, -1)
			}
		})
		t.Run("nil response", func(t *testing.T) {
			result := CIIBestPractices("CIIBestPractices", nil, nil)
			if result.Score != -1 {
				t.Errorf("CIIBestPractices() = %v, want %v", result.Score, -1)
			}
		})
	})
}
