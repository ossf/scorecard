// Copyright 2024 OpenSSF Scorecard Authors
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

package csproj

import (
	"encoding/xml"
	"errors"
)

var errInvalidCsProjFile = errors.New("error parsing csproj file")

type PropertyGroup struct {
	XMLName           xml.Name `xml:"PropertyGroup"`
	RestoreLockedMode bool     `xml:"RestoreLockedMode"`
}

type Project struct {
	XMLName        xml.Name        `xml:"Project"`
	PropertyGroups []PropertyGroup `xml:"PropertyGroup"`
}

func IsRestoreLockedModeEnabled(content []byte) (bool, error) {
	var project Project

	err := xml.Unmarshal(content, &project)
	if err != nil {
		return false, errInvalidCsProjFile
	}

	for _, propertyGroup := range project.PropertyGroups {
		if propertyGroup.RestoreLockedMode {
			return true, nil
		}
	}

	return false, nil
}
