// Copyright 2022 OpenSSF Scorecard Authors
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
package clients

import (
	"reflect"
	"testing"
)

func TestRemoveDuplicate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		keyExtract func(string) string
		list       []string
		want       []string
	}{
		{
			name: "Basic list with dup items",
			list: []string{"A", "B", "C", "B"},
			want: []string{"A", "B", "C"},
			keyExtract: func(in string) string {
				return in
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := removeDuplicate(tt.list, tt.keyExtract)
			if !reflect.DeepEqual(tt.want, got) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
