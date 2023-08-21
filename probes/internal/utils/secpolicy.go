// Copyright 2023 OpenSSF Scorecard Authors
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

package utils

import (
	"github.com/ossf/scorecard/v4/checker"
)

func CountSecInfo(secInfo []checker.SecurityPolicyInformation,
	infoType checker.SecurityPolicyInformationType,
	unique bool,
) int {
	keys := make(map[string]bool)
	count := 0
	for _, entry := range secInfo {
		if _, present := keys[entry.InformationValue.Match]; !present && entry.InformationType == infoType {
			keys[entry.InformationValue.Match] = true
			count += 1
		} else if !unique && entry.InformationType == infoType {
			count += 1
		}
	}
	return count
}

func FindSecInfo(secInfo []checker.SecurityPolicyInformation,
	infoType checker.SecurityPolicyInformationType,
	unique bool,
) []checker.SecurityPolicyInformation {
	keys := make(map[string]bool)
	var secList []checker.SecurityPolicyInformation
	for _, entry := range secInfo {
		if _, present := keys[entry.InformationValue.Match]; !present && entry.InformationType == infoType {
			keys[entry.InformationValue.Match] = true
			secList = append(secList, entry)
		} else if !unique && entry.InformationType == infoType {
			secList = append(secList, entry)
		}
	}
	return secList
}
