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

package pkg

import (
	"fmt"
	"io"
	"math"
	"sort"
	"strings"

	"github.com/ossf/scorecard/v4/docs/checks"
)

const (
	// negInif is "negative infinity" used for dependencydiff results ranking.
	negInf = -math.MaxFloat64
)

type scoreAndDependencyName struct {
	dependencyName string
	aggregateScore float64
}

// DependencydiffResultsAsMarkdown exports the dependencydiff results as markdown.
func DependencydiffResultsAsMarkdown(depdiffResults []DependencyCheckResult, baseHead string,
	doc checks.Doc, writer io.Writer,
) error {
	added := map[string]DependencyCheckResult{}
	removed := map[string]DependencyCheckResult{}
	for _, d := range depdiffResults {
		if d.ChangeType != nil {
			switch *d.ChangeType {
			case Added:
				added[d.Name] = d
			case Removed:
				removed[d.Name] = d
			case Updated:
				// Do nothing, for now.
				// The current data source GitHub Dependency Review won't give the updated dependencies,
				// so we need to find them manually by checking the added/removed maps.
			}
		}
	}
	// Sort dependencies by their aggregate scores in descending orders.
	addedSortKeys, err := getDependencySortKeys(added, doc)
	if err != nil {
		return err
	}
	removedSortKeys, err := getDependencySortKeys(removed, doc)
	if err != nil {
		return err
	}
	sort.SliceStable(
		addedSortKeys,
		func(i, j int) bool { return addedSortKeys[i].aggregateScore > addedSortKeys[j].aggregateScore },
	)
	sort.SliceStable(
		removedSortKeys,
		func(i, j int) bool { return removedSortKeys[i].aggregateScore > removedSortKeys[j].aggregateScore },
	)
	results := ""
	for _, key := range addedSortKeys {
		dName := key.dependencyName
		if _, ok := added[dName]; !ok {
			continue
		}
		current := addedTag()
		if _, ok := removed[dName]; ok {
			// Dependency in the added map also found in the removed map, indicating an updated one.
			current += updatedTag()
		}
		current += scoreTag(key.aggregateScore)
		newResult := added[dName]
		current += fmt.Sprintf(
			"%s @ %s (new) ",
			newResult.Name, *newResult.Version,
		)
		if oldResult, ok := removed[dName]; ok {
			current += fmt.Sprintf(
				"~~%s @ %s (removed)~~ ",
				oldResult.Name, *oldResult.Version,
			)
		}
		results += current + "\n\n"
	}
	for _, key := range removedSortKeys {
		dName := key.dependencyName
		if _, ok := added[dName]; ok {
			// Skip updated ones.
			continue
		}
		if _, ok := removed[dName]; !ok {
			continue
		}
		current := removedTag()
		current += scoreTag(key.aggregateScore)
		oldResult := removed[dName]
		current += fmt.Sprintf(
			"~~%s @ %s~~ ",
			oldResult.Name, *oldResult.Version,
		)
		results += current + "\n\n"
	}
	commits := strings.Split(baseHead, "...")
	base, head := commits[0], commits[1]
	out := fmt.Sprintf(
		"Dependency-diffs (changes) between the BASE commit `%s` and the HEAD commit `%s`:\n\n",
		base, head,
	)
	if results == "" {
		out += fmt.Sprintln("No dependency changes found.")
	} else {
		out += fmt.Sprintln(results)
	}
	fmt.Fprint(writer, out)
	return nil
}

func getDependencySortKeys(dcMap map[string]DependencyCheckResult, doc checks.Doc) ([]scoreAndDependencyName, error) {
	sortKeys := []scoreAndDependencyName{}
	for k := range dcMap {
		scoreAndName := scoreAndDependencyName{
			dependencyName: dcMap[k].Name,
			aggregateScore: negInf,
			// Since this struct is for sorting, the dependency having a score of negative infinite
			//will be put to the very last, unless its agregate score is not empty.
		}
		scResults := dcMap[k].ScorecardResultsWithError.ScorecardResults
		if scResults != nil {
			score, err := scResults.GetAggregateScore(doc)
			if err != nil {
				return nil, err
			}
			scoreAndName.aggregateScore = score
		}
		sortKeys = append(sortKeys, scoreAndName)
	}
	return sortKeys, nil
}

func addedTag() string {
	return fmt.Sprintf(":sparkles: **`" + "added" + "`** ")
}

func updatedTag() string {
	return fmt.Sprintf("**`" + "updated" + "`** ")
}

func removedTag() string {
	return fmt.Sprintf("~~**`" + "removed" + "`**~~ ")
}

func scoreTag(score float64) string {
	switch score {
	case negInf:
		return ""
	default:
		return fmt.Sprintf("`Score: %.1f` ", score)
	}
}
