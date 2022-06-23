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
	"unicode"

	"github.com/h2non/filetype"
	"github.com/h2non/filetype/types"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/checks/fileparser"
	"github.com/ossf/scorecard/v4/clients"
	sce "github.com/ossf/scorecard/v4/errors"
)

// BinaryArtifacts retrieves the raw data for the Binary-Artifacts check.
func BinaryArtifacts(c clients.RepoClient) (checker.BinaryArtifactData, error) {
	files := []checker.File{}
	err := fileparser.OnMatchingFileContentDo(c, fileparser.PathMatcher{
		Pattern:       "*",
		CaseSensitive: false,
	}, checkBinaryFileContent, &files)
	if err != nil {
		return checker.BinaryArtifactData{}, fmt.Errorf("%w", err)
	}

	// No error, return the files.
	return checker.BinaryArtifactData{Files: files}, nil
}

var checkBinaryFileContent fileparser.DoWhileTrueOnFileContent = func(path string, content []byte,
	args ...interface{},
) (bool, error) {
	if len(args) != 1 {
		return false, fmt.Errorf(
			"checkBinaryFileContent requires exactly one argument: %w", errInvalidArgLength)
	}
	pfiles, ok := args[0].(*[]checker.File)
	if !ok {
		return false, fmt.Errorf(
			"checkBinaryFileContent requires argument of type *[]checker.File: %w", errInvalidArgType)
	}

	binaryFileTypes := map[string]bool{
		"crx":    true,
		"deb":    true,
		"dex":    true,
		"dey":    true,
		"elf":    true,
		"o":      true,
		"so":     true,
		"macho":  true,
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
	if exists1 {
		*pfiles = append(*pfiles, checker.File{
			Path:   path,
			Type:   checker.FileTypeBinary,
			Offset: checker.OffsetDefault,
		})
		return true, nil
	}

	exists2 := binaryFileTypes[strings.ReplaceAll(filepath.Ext(path), ".", "")]
	if !isText(content) && exists2 {
		*pfiles = append(*pfiles, checker.File{
			Path:   path,
			Type:   checker.FileTypeBinary,
			Offset: checker.OffsetDefault,
		})
	}

	return true, nil
}

// TODO: refine this function.
func isText(content []byte) bool {
	for _, c := range string(content) {
		if c == '\t' || c == '\n' || c == '\r' {
			continue
		}
		if !unicode.IsPrint(c) {
			return false
		}
	}
	return true
}
