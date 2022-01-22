// Copyright 2022 Security Scorecard Authors
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

package raw

import (
	"reflect"
	"testing"

	"github.com/ossf/scorecard/v4/clients"
)

var branch = "master"

func Test_getBranchName(t *testing.T) {
	t.Parallel()
	type args struct {
		branch *clients.BranchRef
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "simple",
			args: args{
				branch: &clients.BranchRef{
					Name: &branch,
				},
			},
			want: master,
		},
		{
			name: "empty name",
			args: args{
				branch: &clients.BranchRef{},
			},
			want: "",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := getBranchName(tt.args.branch); got != tt.want {
				t.Errorf("getBranchName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getBranchMapFrom(t *testing.T) {
	t.Parallel()
	type args struct {
		branches []*clients.BranchRef
	}
	//nolint
	tests := []struct {
		name string
		args args
		want branchMap
	}{
		{
			name: "simple",
			args: args{
				branches: []*clients.BranchRef{
					{
						Name: &branch,
					},
				},
			},
			want: branchMap{
				master: &clients.BranchRef{
					Name: &branch,
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := getBranchMapFrom(tt.args.branches); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getBranchMapFrom() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_branchMap_getBranchByName(t *testing.T) {
	main := "main"
	t.Parallel()
	type args struct {
		name string
	}
	//nolint
	tests := []struct {
		name    string
		b       branchMap
		args    args
		want    *clients.BranchRef
		wantErr bool
	}{
		{
			name: "simple",
			b: branchMap{
				master: &clients.BranchRef{
					Name: &branch,
				},
			},
			args: args{
				name: master,
			},
			want: &clients.BranchRef{
				Name: &branch,
			},
		},
		{
			name: "main",
			b: branchMap{
				master: &clients.BranchRef{
					Name: &main,
				},
				main: &clients.BranchRef{
					Name: &main,
				},
			},
			args: args{
				name: "main",
			},
			want: &clients.BranchRef{
				Name: &main,
			},
		},
		{
			name: "not found",
			b: branchMap{
				master: &clients.BranchRef{
					Name: &branch,
				},
			},
			args: args{
				name: "not-found",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tt.b.getBranchByName(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("branchMap.getBranchByName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("branchMap.getBranchByName() = %v, want %v", got, tt.want)
			}
		})
	}
}
