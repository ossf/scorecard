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

package registry

import (
	"testing"
)

func TestConstructMavenURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		groupID  string
		artifact string
		version  string
		want     string
	}{
		{
			name:     "simple coordinates",
			groupID:  "org.example",
			artifact: "myapp",
			version:  "1.0.0",
			want:     "https://repo1.maven.org/maven2/org/example/myapp/1.0.0/myapp-1.0.0.jar",
		},
		{
			name:     "nested group ID",
			groupID:  "org.apache.commons",
			artifact: "commons-lang3",
			version:  "3.12.0",
			want: "https://repo1.maven.org/maven2/" +
				"org/apache/commons/commons-lang3/3.12.0/" +
				"commons-lang3-3.12.0.jar",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := constructMavenURL(tt.groupID, tt.artifact, tt.version, ".jar")
			if got != tt.want {
				t.Errorf("constructMavenURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
