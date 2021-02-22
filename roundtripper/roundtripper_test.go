// Copyright 2020 Security Scorecard Authors
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

package roundtripper

import (
	"os"
	"testing"
)

func thelperHandleError(t *testing.T, e error) {
	if e != nil {
		t.Errorf(e.Error())
	}
}

func Test_shouldUseDiskCache(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		diskCachePath string
		useDiskCache  bool
	}{
		{
			name:          "Want to use Disk Cache",
			diskCachePath: "foo",
			useDiskCache:  true,
		},
		{
			name:          "Want to use Disk Cache",
			diskCachePath: "",
			useDiskCache:  false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.useDiskCache {
				if tt.diskCachePath != "" {
					e := os.Setenv(UseDiskCache, "1")
					thelperHandleError(t, e)
					e = os.Setenv(DiskCachePath, tt.diskCachePath)
					thelperHandleError(t, e)
				}
			} else {
				os.Unsetenv(UseDiskCache)
			}
			got, got1 := shouldUseDiskCache()
			if got != tt.diskCachePath {
				t.Errorf("shouldUseDiskCache() got = %v, want %v", got, tt.diskCachePath)
			}
			if got1 != tt.useDiskCache {
				t.Errorf("shouldUseDiskCache() got1 = %v, want %v", got1, tt.useDiskCache)
			}
		})
	}
}
