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

//nolint:stylecheck
package hasOSVVulnerabilities

import (
	"maps"
	"slices"
	"strings"

	"github.com/ossf/scorecard/v5/clients"
)

func intersect(v1, v2 clients.Vulnerability) bool {
	// Check if any aliases intersect.
	for _, alias := range v1.Aliases {
		if slices.Contains(v2.Aliases, alias) {
			return true
		}
	}
	// Check if either IDs are in the others' aliases.
	return slices.Contains(v1.Aliases, v2.ID) || slices.Contains(v2.Aliases, v1.ID)
}

func group(vulns []clients.Vulnerability) []clients.Vulnerability {
	// Mapping of `vulns` index to a group ID. A group ID is just another index in the `vulns` slice.
	groups := make([]int, len(vulns))

	// Initially make every vulnerability its own group.
	for i := range vulns {
		groups[i] = i
	}

	// Do a pair-wise (n^2) comparison and merge all intersecting vulns.
	for i := range vulns {
		for j := i + 1; j < len(vulns); j++ {
			if intersect(vulns[i], vulns[j]) {
				// Merge the two groups. Use the smaller index as the representative ID.
				groups[i] = min(groups[i], groups[j])
				groups[j] = groups[i]
			}
		}
	}

	// Extract groups into the final result structure.
	extractedGroups := map[int][]string{}
	extractedAliases := map[int][]string{}
	for i, gid := range groups {
		extractedGroups[gid] = append(extractedGroups[gid], vulns[i].ID)
		extractedAliases[gid] = append(extractedAliases[gid], vulns[i].Aliases...)
	}

	// Sort by group ID to maintain stable order for tests.
	sortedKeys := slices.Sorted(maps.Keys(extractedGroups))

	result := make([]clients.Vulnerability, 0, len(sortedKeys))
	for _, key := range sortedKeys {
		slices.SortFunc(extractedGroups[key], idSort)

		// Add IDs to aliases
		extractedAliases[key] = append(extractedAliases[key], extractedGroups[key]...)

		// Dedup entries
		slices.Sort(extractedAliases[key])
		extractedAliases[key] = slices.Compact(extractedAliases[key])
		vuln := clients.Vulnerability{Aliases: extractedAliases[key]}
		if len(extractedGroups[key]) > 0 {
			vuln.ID = extractedGroups[key][0]
		}
		result = append(result, vuln)
	}

	return result
}

// match osv-scanner order, since a more specific ID is better
// https://github.com/google/osv-scanner/blob/main/internal/identifiers/identifiers.go#L7-L20
func idSort(a, b string) int {
	aPrefix, _, _ := strings.Cut(a, "-")
	bPrefix, _, _ := strings.Cut(b, "-")

	aOrd := prefixOrder(aPrefix)
	bOrd := prefixOrder(bPrefix)
	if aOrd > bOrd {
		return -1
	} else if aOrd < bOrd {
		return 1
	}

	return strings.Compare(a, b)
}

func prefixOrder(prefix string) int {
	switch prefix {
	case "DSA":
		return 3
	case "CVE":
		return 2
	case "GHSA":
		return 0
	default:
		return 1
	}
}
