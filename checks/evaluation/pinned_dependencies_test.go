// Copyright 2020 OpenSSF Scorecard Authors
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

package evaluation

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/ossf/scorecard/v4/checker"
	scut "github.com/ossf/scorecard/v4/utests"
)

func Test_createReturnForIsGitHubActionsWorkflowPinned(t *testing.T) {
	t.Parallel()
	//nolint
	type args struct {
		r  worklowPinningResult
		dl *scut.TestDetailLogger
	}
	type want struct {
		logs  []string
		score int
	}
	//nolint
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "GitHub-owned and Third-Party actions pinned",
			args: args{
				r: worklowPinningResult{
					thirdParties: pinnedResult{
						pinned: 1,
						total:  1,
					},
					gitHubOwned: pinnedResult{
						pinned: 1,
						total:  1,
					},
				},
				dl: &scut.TestDetailLogger{},
			},
			want: want{
				score: 10,
				logs:  []string{"GitHub-owned GitHubActions are pinned", "Third-party GitHubActions are pinned"},
			},
		},
		{
			name: "only GitHub-owned actions pinned",
			args: args{
				r: worklowPinningResult{
					thirdParties: pinnedResult{
						pinned: 0,
						total:  1,
					},
					gitHubOwned: pinnedResult{
						pinned: 1,
						total:  1,
					},
				},
				dl: &scut.TestDetailLogger{},
			},
			want: want{
				score: 2,
				logs:  []string{"GitHub-owned GitHubActions are pinned"},
			},
		},
		{
			name: "only Third-Party actions pinned",
			args: args{
				r: worklowPinningResult{
					thirdParties: pinnedResult{
						pinned: 1,
						total:  1,
					},
					gitHubOwned: pinnedResult{
						pinned: 0,
						total:  1,
					},
				},
				dl: &scut.TestDetailLogger{},
			},
			want: want{
				score: 8,
				logs:  []string{"Third-party GitHubActions are pinned"},
			},
		},
		{
			name: "no GitHub actions pinned",
			args: args{
				r: worklowPinningResult{
					thirdParties: pinnedResult{
						pinned: 0,
						total:  1,
					},
					gitHubOwned: pinnedResult{
						pinned: 0,
						total:  1,
					},
				},
				dl: &scut.TestDetailLogger{},
			},
			want: want{
				score: 0,
				logs:  []string{},
			},
		},
		{
			name: "no GitHub actions",
			args: args{
				r: worklowPinningResult{
					thirdParties: pinnedResult{
						pinned: 0,
						total:  0,
					},
					gitHubOwned: pinnedResult{
						pinned: 0,
						total:  0,
					},
				},
				dl: &scut.TestDetailLogger{},
			},
			want: want{
				score: -1,
				logs:  []string{"no GitHub-owned GitHubActions found", "no Third-party GitHubActions found"},
			},
		},
		{
			name: "no GitHub-owned actions",
			args: args{
				r: worklowPinningResult{
					thirdParties: pinnedResult{
						pinned: 0,
						total:  1,
					},
					gitHubOwned: pinnedResult{
						pinned: 0,
						total:  0,
					},
				},
				dl: &scut.TestDetailLogger{},
			},
			want: want{
				score: 2,
				logs:  []string{"no GitHub-owned GitHubActions found"},
			},
		},
		{
			name: "no Third-party actions",
			args: args{
				r: worklowPinningResult{
					thirdParties: pinnedResult{
						pinned: 0,
						total:  0,
					},
					gitHubOwned: pinnedResult{
						pinned: 0,
						total:  1,
					},
				},
				dl: &scut.TestDetailLogger{},
			},
			want: want{
				score: 8,
				logs:  []string{"no Third-party GitHubActions found"},
			},
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := createReturnForIsGitHubActionsWorkflowPinned(tt.args.r, tt.args.dl)
			if err != nil {
				t.Errorf("error during createReturnForIsGitHubActionsWorkflowPinned: %v", err)
			}
			if got != tt.want.score {
				t.Errorf("createReturnForIsGitHubActionsWorkflowPinned() = %v, want %v", got, tt.want.score)
			}
			for _, log := range tt.want.logs {
				isExpectedLog := func(logMessage checker.LogMessage, logType checker.DetailType) bool {
					return logMessage.Text == log && logType == checker.DetailInfo
				}
				if !scut.ValidateLogMessage(isExpectedLog, tt.args.dl) {
					t.Errorf("test failed: log message not present: %+v", log)
				}
			}
		})
	}
}

func asStringPointer(s string) *string {
	return &s
}

func asBoolPointer(b bool) *bool {
	return &b
}

func Test_PinningDependencies(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		dependencies []checker.Dependency
		expected     scut.TestReturn
	}{
		{
			name: "all dependencies pinned",
			dependencies: []checker.Dependency{
				{
					Location: &checker.File{},
					Type:     checker.DependencyUseTypePipCommand,
					Pinned:   asBoolPointer(true),
				},
			},
			expected: scut.TestReturn{
				Error:         nil,
				Score:         10,
				NumberOfWarn:  0,
				NumberOfInfo:  8,
				NumberOfDebug: 0,
			},
		},
		{
			name: "all dependencies unpinned",
			dependencies: []checker.Dependency{
				{
					Location: &checker.File{},
					Type:     checker.DependencyUseTypePipCommand,
					Pinned:   asBoolPointer(false),
				},
			},
			expected: scut.TestReturn{
				Error:         nil,
				Score:         0,
				NumberOfWarn:  1,
				NumberOfInfo:  7,
				NumberOfDebug: 0,
			},
		},
		{
			name: "1 ecossystem pinned and 1 ecossystem unpinned",
			dependencies: []checker.Dependency{
				{
					Location: &checker.File{},
					Type:     checker.DependencyUseTypePipCommand,
					Pinned:   asBoolPointer(false),
				},
				{
					Location: &checker.File{},
					Type:     checker.DependencyUseTypeGoCommand,
					Pinned:   asBoolPointer(true),
				},
			},
			expected: scut.TestReturn{
				Error:         nil,
				Score:         5,
				NumberOfWarn:  1,
				NumberOfInfo:  7,
				NumberOfDebug: 0,
			},
		},
		{
			name: "1 ecossystem partially pinned",
			dependencies: []checker.Dependency{
				{
					Location: &checker.File{},
					Type:     checker.DependencyUseTypePipCommand,
					Pinned:   asBoolPointer(false),
				},
				{
					Location: &checker.File{},
					Type:     checker.DependencyUseTypePipCommand,
					Pinned:   asBoolPointer(true),
				},
			},
			expected: scut.TestReturn{
				Error:         nil,
				Score:         0,
				NumberOfWarn:  1,
				NumberOfInfo:  7,
				NumberOfDebug: 0,
			},
		},
		{
			name:         "no dependencies found",
			dependencies: []checker.Dependency{},
			expected: scut.TestReturn{
				Error:         nil,
				Score:         -1,
				NumberOfWarn:  0,
				NumberOfInfo:  8,
				NumberOfDebug: 0,
			},
		},
		{
			name: "pinned dependency shows no warn message",
			dependencies: []checker.Dependency{
				{
					Location: &checker.File{},
					Type:     checker.DependencyUseTypePipCommand,
					Pinned:   asBoolPointer(true),
				},
			},
			expected: scut.TestReturn{
				Error:         nil,
				Score:         10,
				NumberOfWarn:  0,
				NumberOfInfo:  8,
				NumberOfDebug: 0,
			},
		},
		{
			name: "pinned dependency with debug message",
			dependencies: []checker.Dependency{
				{
					Location: &checker.File{},
					Msg:      asStringPointer("some message"),
					Type:     checker.DependencyUseTypePipCommand,
					Pinned:   asBoolPointer(true),
				},
			},
			expected: scut.TestReturn{
				Error:         nil,
				Score:         10,
				NumberOfWarn:  0,
				NumberOfInfo:  8,
				NumberOfDebug: 1,
			},
		},
		{
			name: "unpinned dependency shows warn message",
			dependencies: []checker.Dependency{
				{
					Location: &checker.File{},
					Type:     checker.DependencyUseTypePipCommand,
					Pinned:   asBoolPointer(false),
				},
			},
			expected: scut.TestReturn{
				Error:         nil,
				Score:         0,
				NumberOfWarn:  1,
				NumberOfInfo:  7,
				NumberOfDebug: 0,
			},
		},
		{
			name: "unpinned dependency with debug message shows no warn message",
			dependencies: []checker.Dependency{
				{
					Location: &checker.File{},
					Msg:      asStringPointer("some message"),
					Type:     checker.DependencyUseTypePipCommand,
					Pinned:   asBoolPointer(false),
				},
			},
			expected: scut.TestReturn{
				Error:         nil,
				Score:         0,
				NumberOfWarn:  0,
				NumberOfInfo:  7,
				NumberOfDebug: 1,
			},
		},
		{
			name: "unpinned dependency shows warn message",
			dependencies: []checker.Dependency{
				{
					Location: &checker.File{},
					Msg:      asStringPointer("some message"),
					Type:     checker.DependencyUseTypeDownloadThenRun,
					Pinned:   asBoolPointer(false),
				},
				{
					Location: &checker.File{},
					Type:     checker.DependencyUseTypeDownloadThenRun,
					Pinned:   asBoolPointer(false),
				},
			},
			expected: scut.TestReturn{
				Error:         nil,
				Score:         0,
				NumberOfWarn:  1,
				NumberOfInfo:  6,
				NumberOfDebug: 1,
			},
		},
		// TODO: choco installs should score for Pinned-Dependencies
		// {
		// 	name: "unpinned choco install",
		// 	dependencies: []checker.Dependency{
		// 		{
		// 			Location: &checker.File{},
		// 			Type:     checker.DependencyUseTypeChocoCommand,
		// 			Pinned:   asBoolPointer(false),
		// 		},
		// 	},
		// 	expected: scut.TestReturn{
		// 		Error:         nil,
		// 		Score:         0,
		// 		NumberOfWarn:  1,
		// 		NumberOfInfo:  7,
		// 		NumberOfDebug: 0,
		// 	},
		// },
		{
			name: "unpinned Dockerfile container image",
			dependencies: []checker.Dependency{
				{
					Location: &checker.File{},
					Type:     checker.DependencyUseTypeDockerfileContainerImage,
					Pinned:   asBoolPointer(false),
				},
			},
			expected: scut.TestReturn{
				Error:         nil,
				Score:         0,
				NumberOfWarn:  1,
				NumberOfInfo:  7,
				NumberOfDebug: 0,
			},
		},
		// TODO: Due to a bug download then run is score twice in shell scripts
		// and Dockerfile, the NumberOfInfo should be 7
		{
			name: "unpinned download then run in Dockerfile",
			dependencies: []checker.Dependency{
				{
					Location: &checker.File{
						Path: "Dockerfile",
					},
					Type:   checker.DependencyUseTypeDownloadThenRun,
					Pinned: asBoolPointer(false),
				},
			},
			expected: scut.TestReturn{
				Error:         nil,
				Score:         0,
				NumberOfWarn:  1,
				NumberOfInfo:  6,
				NumberOfDebug: 0,
			},
		},
		// TODO: Due to a bug download then run is score twice in shell scripts
		// and Dockerfile, the NumberOfInfo should be 7
		{
			name: "unpinned download then run in shell scripts",
			dependencies: []checker.Dependency{
				{
					Location: &checker.File{
						Path: "bash.sh",
					},
					Type:   checker.DependencyUseTypeDownloadThenRun,
					Pinned: asBoolPointer(false),
				},
			},
			expected: scut.TestReturn{
				Error:         nil,
				Score:         0,
				NumberOfWarn:  1,
				NumberOfInfo:  6,
				NumberOfDebug: 0,
			},
		},
		{
			name: "unpinned go install",
			dependencies: []checker.Dependency{
				{
					Location: &checker.File{},
					Type:     checker.DependencyUseTypeGoCommand,
					Pinned:   asBoolPointer(false),
				},
			},
			expected: scut.TestReturn{
				Error:         nil,
				Score:         0,
				NumberOfWarn:  1,
				NumberOfInfo:  7,
				NumberOfDebug: 0,
			},
		},
		{
			name: "unpinned npm install",
			dependencies: []checker.Dependency{
				{
					Location: &checker.File{},
					Type:     checker.DependencyUseTypeNpmCommand,
					Pinned:   asBoolPointer(false),
				},
			},
			expected: scut.TestReturn{
				Error:         nil,
				Score:         0,
				NumberOfWarn:  1,
				NumberOfInfo:  7,
				NumberOfDebug: 0,
			},
		},
		// TODO: nuget installs should score for Pinned-Dependencies
		// {
		// 	name: "unpinned nuget install",
		// 	dependencies: []checker.Dependency{
		// 		{
		// 			Location: &checker.File{},
		// 			Type:     checker.DependencyUseTypeNugetCommand,
		// 			Pinned:   asBoolPointer(false),
		// 		},
		// 	},
		// 	expected: scut.TestReturn{
		// 		Error:         nil,
		// 		Score:         0,
		// 		NumberOfWarn:  1,
		// 		NumberOfInfo:  7,
		// 		NumberOfDebug: 0,
		// 	},
		// },
		{
			name: "unpinned pip install",
			dependencies: []checker.Dependency{
				{
					Location: &checker.File{},
					Type:     checker.DependencyUseTypePipCommand,
					Pinned:   asBoolPointer(false),
				},
			},
			expected: scut.TestReturn{
				Error:         nil,
				Score:         0,
				NumberOfWarn:  1,
				NumberOfInfo:  7,
				NumberOfDebug: 0,
			},
		},
		{
			name: "2 unpinned dependencies for 1 ecossystem shows 2 warn messages",
			dependencies: []checker.Dependency{
				{
					Location: &checker.File{},
					Type:     checker.DependencyUseTypePipCommand,
					Pinned:   asBoolPointer(false),
				},
				{
					Location: &checker.File{},
					Type:     checker.DependencyUseTypePipCommand,
					Pinned:   asBoolPointer(false),
				},
			},
			expected: scut.TestReturn{
				Error:         nil,
				Score:         0,
				NumberOfWarn:  2,
				NumberOfInfo:  7,
				NumberOfDebug: 0,
			},
		},
		{
			name: "2 unpinned dependencies for 2 ecossystems shows 2 warn messages",
			dependencies: []checker.Dependency{
				{
					Location: &checker.File{},
					Type:     checker.DependencyUseTypePipCommand,
					Pinned:   asBoolPointer(false),
				},
				{
					Location: &checker.File{},
					Type:     checker.DependencyUseTypeGoCommand,
					Pinned:   asBoolPointer(false),
				},
			},
			expected: scut.TestReturn{
				Error:         nil,
				Score:         0,
				NumberOfWarn:  2,
				NumberOfInfo:  6,
				NumberOfDebug: 0,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dl := scut.TestDetailLogger{}
			c := checker.CheckRequest{Dlogger: &dl}
			actual := PinningDependencies("checkname", &c,
				&checker.PinningDependenciesData{
					Dependencies: tt.dependencies,
				})

			if !scut.ValidateTestReturn(t, tt.name, &tt.expected, &actual, &dl) {
				t.Fail()
			}
		})
	}
}

func Test_createReturnValues(t *testing.T) {
	t.Parallel()
	type args struct {
		pr map[checker.DependencyUseType]pinnedResult
		dl *scut.TestDetailLogger
		t  checker.DependencyUseType
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "no dependencies",
			args: args{
				t:  checker.DependencyUseTypePipCommand,
				dl: &scut.TestDetailLogger{},
			},
			want: -1,
		},
		{
			name: "all dependencies pinned",
			args: args{
				t:  checker.DependencyUseTypePipCommand,
				dl: &scut.TestDetailLogger{},
				pr: map[checker.DependencyUseType]pinnedResult{
					checker.DependencyUseTypePipCommand: {
						pinned: 1,
						total:  1,
					},
				},
			},
			want: 10,
		},
		{
			name: "all dependencies unpinned",
			args: args{
				t:  checker.DependencyUseTypePipCommand,
				dl: &scut.TestDetailLogger{},
				pr: map[checker.DependencyUseType]pinnedResult{
					checker.DependencyUseTypePipCommand: {
						pinned: 0,
						total:  1,
					},
				},
			},
			want: 0,
		},
		{
			name: "1 or more dependencies unpinned",
			args: args{
				t:  checker.DependencyUseTypePipCommand,
				dl: &scut.TestDetailLogger{},
				pr: map[checker.DependencyUseType]pinnedResult{
					checker.DependencyUseTypePipCommand: {
						pinned: 1,
						total:  2,
					},
				},
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := createReturnValues(tt.args.pr, tt.args.t, "all dependencies are pinned",
				"no dependencies found", tt.args.dl)
			if err != nil {
				t.Errorf("error during createReturnValues: %v", err)
			}
			if got != tt.want {
				t.Errorf("createReturnValues() = %v, want %v", got, tt.want)
			}

			switch tt.want {
			case -1:
				isExpectedLog := func(logMessage checker.LogMessage, logType checker.DetailType) bool {
					return logMessage.Text == "no dependencies found" && logType == checker.DetailInfo
				}
				if !scut.ValidateLogMessage(isExpectedLog, tt.args.dl) {
					t.Errorf("test failed: log message not present: %+v", "no dependencies found")
				}
			case 0:
				return
			case 10:
				isExpectedLog := func(logMessage checker.LogMessage, logType checker.DetailType) bool {
					return logMessage.Text == "all dependencies are pinned" && logType == checker.DetailInfo
				}
				if !scut.ValidateLogMessage(isExpectedLog, tt.args.dl) {
					t.Errorf("test failed: log message not present: %+v", "all dependencies are pinned")
				}
			}
		})
	}
}

func Test_generateOwnerToDisplay(t *testing.T) {
	t.Parallel()
	tests := []struct { //nolint:govet
		name        string
		gitHubOwned bool
		want        string
	}{
		{
			name:        "returns GitHub if gitHubOwned is true",
			gitHubOwned: true,
			want:        "GitHub-owned",
		},
		{
			name:        "returns GitHub if gitHubOwned is false",
			gitHubOwned: false,
			want:        "third-party",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := generateOwnerToDisplay(tt.gitHubOwned); got != tt.want {
				t.Errorf("generateOwnerToDisplay() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_addWorkflowPinnedResult(t *testing.T) {
	t.Parallel()
	type args struct {
		dependency *checker.Dependency
		w          *worklowPinningResult
		isGitHub   bool
	}
	tests := []struct { //nolint:govet
		name string
		want *worklowPinningResult
		args args
	}{
		{
			name: "add pinned GitHub-owned action dependency",
			args: args{
				dependency: &checker.Dependency{
					Pinned: asBoolPointer(true),
				},
				w:        &worklowPinningResult{},
				isGitHub: true,
			},
			want: &worklowPinningResult{
				thirdParties: pinnedResult{
					pinned: 0,
					total:  0,
				},
				gitHubOwned: pinnedResult{
					pinned: 1,
					total:  1,
				},
			},
		},
		{
			name: "add unpinned GitHub-owned action dependency",
			args: args{
				dependency: &checker.Dependency{
					Pinned: asBoolPointer(false),
				},
				w:        &worklowPinningResult{},
				isGitHub: true,
			},
			want: &worklowPinningResult{
				thirdParties: pinnedResult{
					pinned: 0,
					total:  0,
				},
				gitHubOwned: pinnedResult{
					pinned: 0,
					total:  1,
				},
			},
		},
		{
			name: "add pinned Third-Party action dependency",
			args: args{
				dependency: &checker.Dependency{
					Pinned: asBoolPointer(true),
				},
				w:        &worklowPinningResult{},
				isGitHub: false,
			},
			want: &worklowPinningResult{
				thirdParties: pinnedResult{
					pinned: 1,
					total:  1,
				},
				gitHubOwned: pinnedResult{
					pinned: 0,
					total:  0,
				},
			},
		},
		{
			name: "add unpinned Third-Party action dependency",
			args: args{
				dependency: &checker.Dependency{
					Pinned: asBoolPointer(false),
				},
				w:        &worklowPinningResult{},
				isGitHub: false,
			},
			want: &worklowPinningResult{
				thirdParties: pinnedResult{
					pinned: 0,
					total:  1,
				},
				gitHubOwned: pinnedResult{
					pinned: 0,
					total:  0,
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			addWorkflowPinnedResult(tt.args.dependency, tt.args.w, tt.args.isGitHub)
			if tt.want.thirdParties != tt.args.w.thirdParties {
				t.Errorf("addWorkflowPinnedResult Third-party GitHub actions mismatch (-want +got):"+
					"\nThird-party pinned: %s\nThird-party total: %s",
					cmp.Diff(tt.want.thirdParties.pinned, tt.args.w.thirdParties.pinned),
					cmp.Diff(tt.want.thirdParties.total, tt.args.w.thirdParties.total))
			}
			if tt.want.gitHubOwned != tt.args.w.gitHubOwned {
				t.Errorf("addWorkflowPinnedResult GitHub-owned GitHub actions mismatch (-want +got):"+
					"\nGitHub-owned pinned: %s\nGitHub-owned total: %s",
					cmp.Diff(tt.want.gitHubOwned.pinned, tt.args.w.gitHubOwned.pinned),
					cmp.Diff(tt.want.gitHubOwned.total, tt.args.w.gitHubOwned.total))
			}
		})
	}
}

func TestGenerateText(t *testing.T) {
	tests := []struct {
		name         string
		dependency   *checker.Dependency
		expectedText string
	}{
		{
			name: "GitHub action not pinned by hash",
			dependency: &checker.Dependency{
				Type: checker.DependencyUseTypeGHAction,
				Location: &checker.File{
					Snippet: "actions/checkout@v2",
				},
			},
			expectedText: "GitHub-owned GitHubAction not pinned by hash",
		},
		{
			name: "Third-party action not pinned by hash",
			dependency: &checker.Dependency{
				Type: checker.DependencyUseTypeGHAction,
				Location: &checker.File{
					Snippet: "third-party/action@v1",
				},
			},
			expectedText: "third-party GitHubAction not pinned by hash",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := generateText(tc.dependency)
			if !cmp.Equal(tc.expectedText, result) {
				t.Errorf("generateText mismatch (-want +got):\n%s", cmp.Diff(tc.expectedText, result))
			}
		})
	}
}

func TestUpdatePinningResults(t *testing.T) {
	t.Parallel()
	type args struct {
		dependency *checker.Dependency
		w          *worklowPinningResult
		pr         map[checker.DependencyUseType]pinnedResult
	}
	type want struct {
		w  *worklowPinningResult
		pr map[checker.DependencyUseType]pinnedResult
	}
	tests := []struct { //nolint:govet
		name string
		args args
		want want
	}{
		{
			name: "add pinned GitHub-owned action",
			args: args{
				dependency: &checker.Dependency{
					Type: checker.DependencyUseTypeGHAction,
					Location: &checker.File{
						Snippet: "actions/checkout@a81bbbf8298c0fa03ea29cdc473d45769f953675",
					},
					Pinned: asBoolPointer(true),
				},
				w:  &worklowPinningResult{},
				pr: make(map[checker.DependencyUseType]pinnedResult),
			},
			want: want{
				w: &worklowPinningResult{
					thirdParties: pinnedResult{
						pinned: 0,
						total:  0,
					},
					gitHubOwned: pinnedResult{
						pinned: 1,
						total:  1,
					},
				},
				pr: make(map[checker.DependencyUseType]pinnedResult),
			},
		},
		{
			name: "add unpinned GitHub-owned action",
			args: args{
				dependency: &checker.Dependency{
					Type: checker.DependencyUseTypeGHAction,
					Location: &checker.File{
						Snippet: "actions/checkout@v2",
					},
					Pinned: asBoolPointer(false),
				},
				w:  &worklowPinningResult{},
				pr: make(map[checker.DependencyUseType]pinnedResult),
			},
			want: want{
				w: &worklowPinningResult{
					thirdParties: pinnedResult{
						pinned: 0,
						total:  0,
					},
					gitHubOwned: pinnedResult{
						pinned: 0,
						total:  1,
					},
				},
				pr: make(map[checker.DependencyUseType]pinnedResult),
			},
		},
		{
			name: "add pinned Third-party action",
			args: args{
				dependency: &checker.Dependency{
					Type: checker.DependencyUseTypeGHAction,
					Location: &checker.File{
						Snippet: "other/checkout@ffa6706ff2127a749973072756f83c532e43ed02",
					},
					Pinned: asBoolPointer(true),
				},
				w:  &worklowPinningResult{},
				pr: make(map[checker.DependencyUseType]pinnedResult),
			},
			want: want{
				w: &worklowPinningResult{
					thirdParties: pinnedResult{
						pinned: 1,
						total:  1,
					},
					gitHubOwned: pinnedResult{
						pinned: 0,
						total:  0,
					},
				},
				pr: make(map[checker.DependencyUseType]pinnedResult),
			},
		},
		{
			name: "add unpinned Third-party action",
			args: args{
				dependency: &checker.Dependency{
					Type: checker.DependencyUseTypeGHAction,
					Location: &checker.File{
						Snippet: "other/checkout@v2",
					},
					Pinned: asBoolPointer(false),
				},
				w:  &worklowPinningResult{},
				pr: make(map[checker.DependencyUseType]pinnedResult),
			},
			want: want{
				w: &worklowPinningResult{
					thirdParties: pinnedResult{
						pinned: 0,
						total:  1,
					},
					gitHubOwned: pinnedResult{
						pinned: 0,
						total:  0,
					},
				},
				pr: make(map[checker.DependencyUseType]pinnedResult),
			},
		},
		{
			name: "add pinned pip install",
			args: args{
				dependency: &checker.Dependency{
					Type:   checker.DependencyUseTypePipCommand,
					Pinned: asBoolPointer(true),
				},
				w:  &worklowPinningResult{},
				pr: make(map[checker.DependencyUseType]pinnedResult),
			},
			want: want{
				w: &worklowPinningResult{},
				pr: map[checker.DependencyUseType]pinnedResult{
					checker.DependencyUseTypePipCommand: {
						pinned: 1,
						total:  1,
					},
				},
			},
		},
		{
			name: "add unpinned pip install",
			args: args{
				dependency: &checker.Dependency{
					Type:   checker.DependencyUseTypePipCommand,
					Pinned: asBoolPointer(false),
				},
				w:  &worklowPinningResult{},
				pr: make(map[checker.DependencyUseType]pinnedResult),
			},
			want: want{
				w: &worklowPinningResult{},
				pr: map[checker.DependencyUseType]pinnedResult{
					checker.DependencyUseTypePipCommand: {
						pinned: 0,
						total:  1,
					},
				},
			},
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			updatePinningResults(tc.args.dependency, tc.args.w, tc.args.pr)
			if tc.want.w.thirdParties != tc.args.w.thirdParties {
				t.Errorf("updatePinningResults Third-party GitHub actions mismatch (-want +got):"+
					"\nThird-party pinned: %s\nThird-party total: %s",
					cmp.Diff(tc.want.w.thirdParties.pinned, tc.args.w.thirdParties.pinned),
					cmp.Diff(tc.want.w.thirdParties.total, tc.args.w.thirdParties.total))
			}
			if tc.want.w.gitHubOwned != tc.args.w.gitHubOwned {
				t.Errorf("updatePinningResults GitHub-owned GitHub actions mismatch (-want +got):"+
					"\nGitHub-owned pinned: %s\nGitHub-owned total: %s",
					cmp.Diff(tc.want.w.gitHubOwned.pinned, tc.args.w.gitHubOwned.pinned),
					cmp.Diff(tc.want.w.gitHubOwned.total, tc.args.w.gitHubOwned.total))
			}
			for dependencyUseType := range tc.want.pr {
				if tc.want.pr[dependencyUseType] != tc.args.pr[dependencyUseType] {
					t.Errorf("updatePinningResults %s mismatch (-want +got):\npinned: %s\ntotal: %s",
						dependencyUseType,
						cmp.Diff(tc.want.pr[dependencyUseType].pinned, tc.args.pr[dependencyUseType].pinned),
						cmp.Diff(tc.want.pr[dependencyUseType].total, tc.args.pr[dependencyUseType].total))
				}
			}
		})
	}
}
