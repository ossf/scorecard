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

// Package options implements Scorecard options.
package options

import (
	"errors"
	"fmt"
)

// DependencydiffOptions define common options for configuring scorecard dependency-diff.
type DependencydiffOptions struct {
	// Base is the base branch name reference or the base commitSHA.
	Base string

	// Head is the head branch name reference or the head commitSHA.
	Head string

	// ChangeTypes is an array of dependency change types for specifying what change types the dependency-diff
	// will surface the scorecard results for.
	// This is not a required option and can be nullable. If null, we will surface the scorecard results
	// for all types of dependencies.
	ChangeTypes []string
}

var (
	errBaseIsEmpty       = errors.New("base should be non-empty")
	errHeadIsEmpty       = errors.New("head should be non-empty")
	errInvalidChangeType = errors.New("invalid change type")
)

// NewDepdiff creates a new instance of `DependencydiffOptions`.
func NewDepdiff() *DependencydiffOptions {
	depdiffOpts := &DependencydiffOptions{}
	// No need to do the env.Parse() for a dependency-diff option since there
	// are no struct env tags for now.
	return depdiffOpts
}

// Validate validates scorecard dependency-diff configuration options.
func (depOptions *DependencydiffOptions) Validate() error {
	var errs []error
	// Validate `base` is non-empty.
	if depOptions.Base == "" {
		errs = append(
			errs,
			errBaseIsEmpty,
		)
	}
	// Validate `head` is non-empty.
	if depOptions.Head == "" {
		errs = append(
			errs,
			errHeadIsEmpty,
		)
	}
	// ChangeTypes can be null, but users must give valid types if this param is specified.
	if len(depOptions.ChangeTypes) != 0 {
		for _, ct := range depOptions.ChangeTypes {
			if !isChangeTypeValid(ct) {
				errs = append(
					errs,
					errInvalidChangeType,
				)
			}
		}
	}
	if len(errs) != 0 {
		return fmt.Errorf(
			"%w: %+v",
			errValidate,
			errs,
		)
	}
	return nil
}

func isChangeTypeValid(ct string) bool {
	return ct == "added" || ct == "removed"
}
