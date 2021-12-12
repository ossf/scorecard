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
	"path/filepath"
	"strings"

	"github.com/h2non/filetype"
	"github.com/h2non/filetype/types"

	"github.com/ossf/scorecard/v3/checker"
	"github.com/ossf/scorecard/v3/checks/fileparser"
	sce "github.com/ossf/scorecard/v3/errors"
)

// File represents a file.
type File struct {
	Path string
	// TODO: add hash if needed.
}

// BinaryArtifactData contains the raw results.
type BinaryArtifactData struct {
	// Files contains a list of files.
	Files []File
}

// BinaryArtifacts retrieves the raw data for the Binary-Artifacts check.
func BinaryArtifacts(c *checker.CheckRequest) (BinaryArtifactData, error) {
	var files []File
	err := fileparser.CheckFilesContentV6("*", false, c.RepoClient, checkBinaryFileContent, &files)
	if err != nil {
		return BinaryArtifactData{}, err
	}

	// No error, return the files.
	return BinaryArtifactData{Files: files}, nil
}

func checkBinaryFileContent(path string, content []byte,
	data fileparser.FileCbData) (bool, error) {
	pfiles, ok := data.(*[]File)
	if !ok {
		// This never happens.
		panic("invalid type")
	}

	binaryFileTypes := map[string]bool{
		"crx":    true,
		"deb":    true,
		"dex":    true,
		"dey":    true,
		"elf":    true,
		"o":      true,
		"so":     true,
		"iso":    true,
		"class":  true,
		"jar":    true,
		"bundle": true,
		"dylib":  true,
		"lib":    true,
		"msi":    true,
		"dll":    true,
		"drv":    true,
		"efi":    true,
		"exe":    true,
		"ocx":    true,
		"pyc":    true,
		"pyo":    true,
		"par":    true,
		"rpm":    true,
		"whl":    true,
	}
	var t types.Type
	var err error
	if len(content) == 0 {
		return true, nil
	}
	if t, err = filetype.Get(content); err != nil {
		return false, sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("filetype.Get:%v", err))
	}

	exists1 := binaryFileTypes[t.Extension]
	exists2 := binaryFileTypes[strings.ReplaceAll(filepath.Ext(path), ".", "")]
	if exists1 || exists2 {
		*pfiles = append(*pfiles, File{
			Path: path,
		})
	}

	return true, nil
}
