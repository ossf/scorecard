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

package clients

import (
	"path/filepath"
	"strings"

	"github.com/h2non/filetype"
	"github.com/h2non/filetype/types"

	"github.com/ossf/scorecard/v4/errors"
)

// BinaryArtifactsClient is a new BinaryArtifacts client.
type BinaryArtifactsClient interface {
	CheckBinaryFileContent(path string, content *[]byte) (bool, error)
}

type binaryArtifactsClient struct{}

// DefaultBinaryArtifactsClient is the default BinaryArtifacts client.
func DefaultBinaryArtifactsClient() BinaryArtifactsClient {
	return binaryArtifactsClient{}
}

// CheckBinaryFileContent checks if the file is a binary file.
func (v binaryArtifactsClient) CheckBinaryFileContent(path string, content *[]byte) (bool, error) {
	if path == "" {
		return false, errInternalPathEmpty
	}
	if content == nil {
		return false, errContentNil
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
	if len(*content) == 0 {
		return true, nil
	}
	if t, err = filetype.Get(*content); err != nil {
		return false, errors.WithMessage(err, "could not get file type")
	}

	return binaryFileTypes[t.Extension] || binaryFileTypes[strings.ReplaceAll(filepath.Ext(path), ".", "")], nil
}
