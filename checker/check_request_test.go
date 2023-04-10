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
package checker

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestListUnsupported(t *testing.T) {
	t.Parallel()
	type args struct {
		required  []RequestType
		supported []RequestType
	}
	tests := []struct {
		name string
		args args
		want []RequestType
	}{
		{
			name: "empty",
			args: args{
				required:  []RequestType{},
				supported: []RequestType{},
			},
			want: []RequestType{FileBased},
		},
		{
			name: "empty required",
			args: args{
				required:  []RequestType{},
				supported: []RequestType{FileBased},
			},
			want: []RequestType{},
		},
		{
			name: "supported",
			args: args{
				required:  []RequestType{FileBased},
				supported: []RequestType{FileBased},
			},
			want: []RequestType{FileBased},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := ListUnsupported(tt.args.required, tt.args.supported); cmp.Equal(got, tt.want) {
				t.Errorf("ListUnsupported() = %v, want %v", got, cmp.Diff(got, tt.want))
			}
		})
	}
}

func Test_contains(t *testing.T) {
	t.Parallel()
	type args struct {
		in     []RequestType
		exists RequestType
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "empty",
			args: args{
				in:     []RequestType{},
				exists: FileBased,
			},
			want: false,
		},
		{
			name: "empty exists",
			args: args{
				in:     []RequestType{FileBased},
				exists: FileBased,
			},
			want: true,
		},
		{
			name: "empty exists",
			args: args{
				in:     []RequestType{FileBased},
				exists: CommitBased,
			},
			want: false,
		},
		{
			name: "empty exists",
			args: args{
				in:     []RequestType{FileBased, CommitBased},
				exists: CommitBased,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := contains(tt.args.in, tt.args.exists); got != tt.want {
				t.Errorf("contains() = %v, want %v", got, tt.want)
			}
		})
	}
}
