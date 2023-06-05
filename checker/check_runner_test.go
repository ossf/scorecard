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
	"context"
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestNewRunner(t *testing.T) {
	t.Parallel()
	type args struct { //nolint:govet
		checkName string
		repo      string
		checkReq  *CheckRequest
	}
	tests := []struct { //nolint:govet
		name string
		args args
		want *Runner
	}{
		{
			name: "TestNewRunner",
			args: args{
				checkName: "TestNewRunner",
				repo:      "TestNewRunner",
				checkReq:  &CheckRequest{},
			},
			want: &Runner{
				CheckName:    "TestNewRunner",
				Repo:         "TestNewRunner",
				CheckRequest: CheckRequest{},
			},
		},
		{
			name: "TestNewRunner2",
			args: args{
				checkName: "TestNewRunner2",
				repo:      "TestNewRunner2",
				checkReq:  &CheckRequest{},
			},
			want: &Runner{
				CheckName:    "TestNewRunner2",
				Repo:         "TestNewRunner2",
				CheckRequest: CheckRequest{},
			},
		},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := NewRunner(tt.args.checkName, tt.args.repo, tt.args.checkReq); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewRunner() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRunner_SetCheckName(t *testing.T) {
	t.Parallel()
	type fields struct {
		CheckName    string
		Repo         string
		CheckRequest CheckRequest
	}
	type args struct {
		check string
	}
	tests := []struct { //nolint:govet
		name   string
		fields fields
		args   args
	}{
		{
			name: "TestRunner_SetCheckName",
			fields: fields{
				CheckName:    "TestRunner_SetCheckName",
				Repo:         "TestRunner_SetCheckName",
				CheckRequest: CheckRequest{},
			},
			args: args{
				check: "TestRunner_SetCheckName",
			},
		},
		{
			name: "TestRunner_SetCheckName2",
			fields: fields{
				CheckName:    "TestRunner_SetCheckName2",
				Repo:         "TestRunner_SetCheckName2",
				CheckRequest: CheckRequest{},
			},
			args: args{
				check: "TestRunner_SetCheckName2",
			},
		},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := &Runner{
				CheckName:    tt.fields.CheckName,
				Repo:         tt.fields.Repo,
				CheckRequest: tt.fields.CheckRequest,
			}
			r.SetCheckName(tt.args.check)
			if r.CheckName != tt.args.check {
				t.Errorf("SetCheckName() = %v, want %v", r.CheckName, tt.args.check)
			}
		})
	}
}

func TestRunner_SetRepo(t *testing.T) {
	t.Parallel()
	type fields struct {
		CheckName    string
		Repo         string
		CheckRequest CheckRequest
	}
	type args struct {
		repo string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "TestRunner_SetCheckName",
			fields: fields{
				CheckName:    "TestRunner_SetCheckName",
				Repo:         "TestRunner_SetCheckName",
				CheckRequest: CheckRequest{},
			},
			args: args{
				repo: "TestRunner_SetCheckName",
			},
		},
		{
			name: "TestRunner_SetCheckName2",
			fields: fields{
				CheckName:    "TestRunner_SetCheckName2",
				Repo:         "TestRunner_SetCheckName2",
				CheckRequest: CheckRequest{},
			},
			args: args{
				repo: "TestRunner_SetCheckName2",
			},
		},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := &Runner{
				CheckName:    tt.fields.CheckName,
				Repo:         tt.fields.Repo,
				CheckRequest: tt.fields.CheckRequest,
			}
			r.SetRepo(tt.args.repo)
			if r.Repo != tt.args.repo {
				t.Errorf("SetCheckName() = %v, want %v", r.CheckName, tt.args.repo)
			}
		})
	}
}

func Test_logStats(t *testing.T) {
	t.Parallel()
	type args struct {
		ctx       context.Context
		startTime time.Time
		result    *CheckResult
	}
	tests := []struct { //nolint:govet
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Test_logStats",
			args: args{
				ctx:       context.Background(),
				startTime: time.Now(),
				result:    &CheckResult{},
			},
		},
		{
			name: "Test_logStats2",
			args: args{
				ctx:       context.Background(),
				startTime: time.Now(),
				result:    &CheckResult{Error: errors.New("Test_logStats2")}, //nolint:goerr113
			},
		},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := logStats(tt.args.ctx, tt.args.startTime, tt.args.result); (err != nil) != tt.wantErr {
				t.Errorf("logStats() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
