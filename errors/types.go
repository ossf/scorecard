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
	"fmt"
)

type (
	// ErrRetry is returned when a check failed after maximum num_retries.
	ErrRetry struct{ wrappedError }
	// ErrLowConfidence is returned when check result is inconclusive.
	ErrLowConfidence struct{ wrappedError }
)

// MakeRetryError returns a wrapped error of type ErrRetry.
func MakeRetryError(err error) error {
	return &ErrRetry{
		wrappedError{
			msg:        "unable to run check, retry",
			innerError: err,
		},
	}
}

// MakeLowConfidenceError returns a wrapped error of type ErrLowConfidence.
func MakeLowConfidenceError(err error) error {
	return &ErrLowConfidence{
		wrappedError{
			msg:        "low confidence check result",
			innerError: err,
		},
	}
}

type wrappedError struct {
	innerError error
	msg        string
}

func (err *wrappedError) Error() string {
	return fmt.Sprintf("%s: %v", err.msg, err.innerError)
}

func (err *wrappedError) Unwrap() error {
	return err.innerError
}
