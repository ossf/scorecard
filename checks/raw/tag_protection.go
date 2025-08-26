// Copyright 2025 OpenSSF Scorecard Authors
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
	"errors"
	"fmt"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/clients"
)

type tagSet struct {
	exists map[string]bool
	set    []clients.RepoRef
}

func (set *tagSet) add(tag *clients.RepoRef) bool {
	if tag != nil &&
		tag.Name != nil &&
		*tag.Name != "" &&
		!set.exists[*tag.Name] {
		set.set = append(set.set, *tag)
		set.exists[*tag.Name] = true
		return true
	}
	return false
}

// TagProtection retrieves the raw data for the Tag-Protection check.
func TagProtection(cr *checker.CheckRequest) (checker.TagProtectionsData, error) {
	c := cr.RepoClient
	tags := tagSet{
		exists: make(map[string]bool),
	}
	repoTags, err := c.ListTags()
	if err != nil && !errors.Is(err, clients.ErrUnsupportedFeature) {
		return checker.TagProtectionsData{}, fmt.Errorf("%w", err)
	}
	for _, tag := range repoTags {
		tagRef, err := c.GetTag(*tag.Name)
		if err != nil {
			return checker.TagProtectionsData{},
				fmt.Errorf("error during GetTag(): %w", err)
		}
		tags.add(tagRef)
	}

	codeownersFiles := []string{}
	if err := collectCodeownersFiles(c, &codeownersFiles); err != nil {
		return checker.TagProtectionsData{}, err
	}

	// No error, return the data.
	return checker.TagProtectionsData{
		Tags:            tags.set,
		CodeownersFiles: codeownersFiles,
	}, nil
}
