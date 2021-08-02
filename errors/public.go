// Copyright 2021 Security Scorecard Authors
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

package errors

import (
	"errors"
	"fmt"
)

//nolint
var (
	ErrScorecardInternal = errors.New("internal error")
	ErrRepoUnreachable   = errors.New("repo unreachable")
)

// Create a public error using any of the errors
// listed above. For examples, see errors/errors.md.
func Create(e error, msg string) error {
	// Note: Errorf automatically wraps the error when used with `%w`.
	if len(msg) > 0 {
		return fmt.Errorf("%w: %v", e, msg)
	}
	// We still need to use %w to prevent callers from using e == ErrInvalidDockerFile.
	return fmt.Errorf("%w", e)
}

// GetName returns the name of the error.
func GetName(err error) string {
	switch {
	case errors.Is(err, ErrScorecardInternal):
		return "ErrScorecardInternal"
	case errors.Is(err, ErrRepoUnreachable):
		return "ErrRepoUnreachable"
	default:
		return "ErrUnknown"
	}
}
