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

package git

import (
	"errors"
	"testing"
	"time"

	"github.com/ossf/scorecard/v5/clients"
)

func TestGetMaintainerActivity(t *testing.T) {
	t.Parallel()

	client := &Client{}

	cutoff := time.Now().UTC().AddDate(0, -6, 0)
	result, err := client.GetMaintainerActivity(cutoff)

	// Should return ErrUnsupportedFeature
	if !errors.Is(err, clients.ErrUnsupportedFeature) {
		t.Errorf("expected ErrUnsupportedFeature, got %v", err)
	}

	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}
