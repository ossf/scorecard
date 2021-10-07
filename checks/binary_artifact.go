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

package checks

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/h2non/filetype"
	"github.com/h2non/filetype/types"

	"github.com/ossf/scorecard/v3/checker"
	sce "github.com/ossf/scorecard/v3/errors"
)

// CheckBinaryArtifacts is the exported name for Binary-Artifacts check.
const CheckBinaryArtifacts string = "Binary-Artifacts"

//nolint
func init() {
	registerCheck(CheckBinaryArtifacts, BinaryArtifacts)
}

// BinaryArtifacts  will check the repository if it contains binary artifacts.
func BinaryArtifacts(c *checker.CheckRequest) checker.CheckResult {
	var binFound bool
	err := CheckFilesContent("*", false, c, checkBinaryFileContent, &binFound)
	if err != nil {
		return checker.CreateRuntimeErrorResult(CheckBinaryArtifacts, err)
	}
	if binFound {
		return checker.CreateMinScoreResult(CheckBinaryArtifacts, "binaries present in source code")
	}

	return checker.CreateMaxScoreResult(CheckBinaryArtifacts, "no binaries found in the repo")
}

func checkBinaryFileContent(path string, content []byte,
	dl checker.DetailLogger, data FileCbData) (bool, error) {
	pfound := FileGetCbDataAsBoolPointer(data)
	binaryFileTypes := map[string]bool{
		"crx":     true,
		"deb":     true,
		"dex":     true,
		"dey":     true,
		"elf":     true,
		"bin":     true,
		"o":       true,
		"so":      true,
		"iso":     true,
		"class":   true,
		"jar":     true,
		"bundle":  true,
		"dylib":   true,
		"lib":     true,
		"msi":     true,
		"acm":     true,
		"ax":      true,
		"cpl":     true,
		"dll":     true,
		"drv":     true,
		"efi":     true,
		"exe":     true,
		"mui":     true,
		"ocx":     true,
		"scr":     true,
		"sys":     true,
		"tsp":     true,
		"pyc":     true,
		"pyo":     true,
		"par":     true,
		"rpm":     true,
		"swf":     true,
		"torrent": true,
		"cab":     true,
		"whl":     true,
	}
	var t types.Type
	var err error
	if t, err = filetype.Get(content); err != nil {
		return false, sce.WithMessage(sce.ErrScorecardInternal, fmt.Sprintf("filetype.Get:%v", err))
	}

	if _, ok := binaryFileTypes[t.Extension]; ok {
		dl.Warn3(&checker.LogMessage{
			Path: path, Type: checker.FileTypeBinary,
			Text: "binary detected",
		})
		*pfound = true
		return true, nil
	} else if _, ok := binaryFileTypes[strings.ReplaceAll(filepath.Ext(path), ".", "")]; ok {
		// Falling back to file based extension.
		dl.Warn3(&checker.LogMessage{
			Path: path, Type: checker.FileTypeBinary,
			Text: "binary detected",
		})
		*pfound = true
		return true, nil
	}

	return true, nil
}
