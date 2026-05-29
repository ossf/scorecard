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

func TestPackagingIncludesWorkflowPathOnParseError(t *testing.T) {
	t.Parallel()

	const workflowPath = ".github/workflows/bad.yaml"
	ctrl := gomock.NewController(t)
	mockRepo := mockrepo.NewMockRepoClient(ctrl)
	mockRepo.EXPECT().ListFiles(gomock.Any()).Return([]string{workflowPath}, nil)
	mockRepo.EXPECT().GetFileReader(workflowPath).Return(
		io.NopCloser(strings.NewReader("name: bad\non: [push\n")),
		nil,
	)

	_, err := Packaging(&checker.CheckRequest{RepoClient: mockRepo})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), workflowPath) {
		t.Fatalf("expected error to include %q, got %q", workflowPath, err)
	}
}
