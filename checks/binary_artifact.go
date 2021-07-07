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

	"github.com/h2non/filetype"
	"github.com/h2non/filetype/types"

	"github.com/ossf/scorecard/checker"
)

//nolint
func init() {
	registerCheck(checkBinaryArtifacts, binaryArtifacts)
}

// TODO: read the check code from file?
const checkBinaryArtifacts string = "Binary-Artifacts"

// BinaryArtifacts  will check the repository if it contains binary artifacts.
func binaryArtifacts(c *checker.CheckRequest) checker.CheckResult {
	r, err := CheckFilesContent2("*", false, c, checkBinaryFileContent)
	if err != nil {
		// TODO: check for the repo retry error, which should be a common
		// scorecard error independent of the underlying implementation.
		return checker.MakeInternalErrorResult(checkBinaryArtifacts, err)
	}
	if !r {
		// We need not provid a reason because it's already done
		// in checkBinaryFileContent via `Pass` call.
		return checker.MakeFailResultWithHighConfidence(checkBinaryArtifacts)
	}

	// High confidence result.
	// We provide a reason to help the user.
	return checker.MakePassResultWithHighConfidenceAndReason(checkBinaryArtifacts, c, "no binary files found in the repo")
}

func checkBinaryFileContent(path string, content []byte,
	cl checker.CheckLogger) (bool, error) {
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
		return false, fmt.Errorf("failed in getting the content type %w", err)
	}

	if _, ok := binaryFileTypes[t.Extension]; ok {
		cl.Fail("binary-artifact found: %s", path)
		return false, nil
	} else if _, ok := binaryFileTypes[filepath.Ext(path)]; ok {
		// Falling back to file based extension.
		cl.Fail("binary-artifact found: %s", path)
		return false, nil
	}

	return true, nil
}
