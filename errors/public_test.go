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
package errors

import (
	"errors"
	"strings"
	"testing"
)

func TestWithMessage(t *testing.T) {
	t.Parallel()
	type args struct {
		e   error
		msg string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "with error and message",
			args: args{
				e:   ErrScorecardInternal,
				msg: "additional context",
			},
			wantErr: true,
		},
		{
			name: "with error and no message",
			args: args{
				e:   ErrScorecardInternal,
				msg: "",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := WithMessage(tt.args.e, tt.args.msg); (err != nil) != tt.wantErr {
				t.Errorf("WithMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetName(t *testing.T) {
	t.Parallel()
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "ErrScorecardInternal",
			args: args{
				err: ErrScorecardInternal,
			},
			want: "ErrScorecardInternal",
		},
		{
			name: "ErrRepoUnreachable",
			args: args{
				err: ErrRepoUnreachable,
			},
			want: "ErrRepoUnreachable",
		},
		{
			name: "ErrorShellParsing",
			args: args{
				err: ErrorShellParsing,
			},
			want: "ErrorShellParsing",
		},
		{
			name: "unknown error",
			args: args{
				err: errors.New("unknown error"), //nolint:goerr113
			},
			want: "ErrUnknown",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := GetName(tt.args.err); !strings.EqualFold(got, tt.want) {
				t.Errorf("GetName() = %v, want %v", got, tt.want)
			}
		})
	}
}
