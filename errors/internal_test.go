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
	"fmt"
	"testing"
)

func TestCreateInternal(t *testing.T) {
	type args struct {
		e   error
		msg string
	}
	test := struct { //nolint:govet
		name string
		args args
		want error
	}{
		name: "non-nil error and non-empty message",
		args: args{
			e:   errors.New("test error"), //nolint:goerr113
			msg: "test message",
		},
		want: fmt.Errorf("test error: test message"), //nolint:goerr113
	}

	t.Run(test.name, func(t *testing.T) {
		if got := CreateInternal(test.args.e, test.args.msg); got.Error() != test.want.Error() {
			t.Errorf("CreateInternal() = %v, want %v", got, test.want)
		}
	})
}
