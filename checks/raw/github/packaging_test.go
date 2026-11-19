// Copyright 2026 OpenSSF Scorecard Authors
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

package github

import (
	"io"
	"strings"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/ossf/scorecard/v5/checker"
	mockrepo "github.com/ossf/scorecard/v5/clients/mockclients"
)

func TestPackagingParseErrorIncludesFilePath(t *testing.T) {
	t.Parallel()

	const path = ".github/workflows/broken-release.yaml"
	ctrl := gomock.NewController(t)
	repo := mockrepo.NewMockRepoClient(ctrl)
	repo.EXPECT().ListFiles(gomock.Any()).Return([]string{path}, nil)
	repo.EXPECT().GetFileReader(path).Return(
		io.NopCloser(strings.NewReader("jobs:\n  release:\n    runs-on: [\n")), nil)

	data, err := Packaging(&checker.CheckRequest{RepoClient: repo})
	if err == nil {
		t.Fatal("Packaging() succeeded with an unparsable workflow")
	}
	if !strings.Contains(err.Error(), "("+path+"):") {
		t.Errorf("Packaging() error = %q, want workflow path %q", err, path)
	}
	if !strings.Contains(err.Error(), "[syntax-check]") {
		t.Errorf("Packaging() error = %q, want the parser diagnostic", err)
	}
	if len(data.Packages) != 0 {
		t.Errorf("Packaging() returned packages for an unparsable workflow: %+v", data.Packages)
	}
}
