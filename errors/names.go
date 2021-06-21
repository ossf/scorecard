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

package errors

import (
	"errors"
)

const (
	// RetryError occurs when checks fail after exhausting all retry attempts.
	RetryError = "RetryError"
	// ZeroConfidenceError shows an inconclusive result.
	ZeroConfidenceError = "ZeroConfidenceError"
	// UnknownError for all error types not handled.
	UnknownError = "UnknownError"
)

var (
	errRetry          *ErrRetry
	errZeroConfidence *ErrZeroConfidence
)

func GetErrorName(err error) string {
	switch {
	case errors.As(err, &errRetry):
		return RetryError
	case errors.As(err, &errZeroConfidence):
		return ZeroConfidenceError
	default:
		return UnknownError
	}
}
