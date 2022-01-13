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

package raw

import (
	"fmt"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/checks/fileparser"
	"github.com/ossf/scorecard/v4/clients"
	"github.com/ossf/scorecard/v4/errors"
)

// BinaryArtifacts retrieves the raw data for the Binary-Artifacts check.
func BinaryArtifacts(c clients.RepoClient, bac clients.BinaryArtifactsClient) (checker.BinaryArtifactData, error) {
	files := []checker.File{}
	err := fileparser.CheckFilesContentV6("*", false, c, checkBinaryFileContent, &files, bac)
	if err != nil {
		return checker.BinaryArtifactData{}, fmt.Errorf("%w", err)
	}

	// No error, return the files.
	return checker.BinaryArtifactData{Files: files}, nil
}

func checkBinaryFileContent(path string, content []byte,
	data fileparser.FileCbData, bac clients.BinaryArtifactsClient) (bool, error) {
	pfiles, ok := data.(*[]checker.File)
	if !ok {
		// This never happens.
		panic("invalid type")
	}
	result, err := bac.CheckBinaryFileContent(path, &content)
	if err != nil {
		return false, errors.WithMessage(err, "error checking binary file content")
	}
	if result {
		*pfiles = append(*pfiles, checker.File{
			Path:   path,
			Type:   checker.FileTypeBinary,
			Offset: checker.OffsetDefault,
		})
	}
	return true, nil
}
