// Copyright 2026 OpenSSF Scorecard Authors
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

package localdir

import (
	"testing"
	"time"
)

func TestGetMaintainerActivity(t *testing.T) {
	t.Parallel()

	client := &Client{}

	cutoff := time.Now().UTC().AddDate(0, -6, 0)
	result, err := client.GetMaintainerActivity(cutoff)
	// localdir returns an empty map (no maintainer info available)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}

	if result == nil {
		t.Error("expected non-nil result (empty map)")
	}

	if len(result) != 0 {
		t.Errorf("expected empty map, got %d entries", len(result))
	}
}
