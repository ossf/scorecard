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

package host

import "testing"

func TestIsGitHub(t *testing.T) {
	t.Setenv("GH_HOST", "ghe.example.com")
	tests := []struct {
		host string
		want bool
	}{
		{host: "", want: true},
		{host: "github.com", want: true},
		{host: "ghe.example.com", want: true},
		{host: "gitlab.com", want: false},
		{host: "dev.azure.com", want: false},
	}
	for _, tt := range tests {
		if got := IsGitHub(tt.host); got != tt.want {
			t.Errorf("IsGitHub(%q) = %v, want %v", tt.host, got, tt.want)
		}
	}
}
