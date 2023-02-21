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

	"github.com/ossf/scorecard/v4/checker"
	scut "github.com/ossf/scorecard/v4/utests"
)

func Test_createReturnValuesForGitHubActionsWorkflowPinned(t *testing.T) {
	t.Parallel()
	//nolint
	type args struct {
		r       worklowPinningResult
		infoMsg string
		dl      checker.DetailLogger
	}
	//nolint
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "both actions workflow pinned",
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
			want: 10,
		},
		{
			name: "github actions workflow pinned",
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
			want: 2,
		},
		{
			name: "third-parties actions workflow pinned",
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
			want: 8,
		},
		{
			name: "partial actions workflow pinned",
			args: args{
				r: worklowPinningResult{
					thirdParties: pinnedResult{
						pinned: 1,
						total:  2,
					},
					gitHubOwned: pinnedResult{
						pinned: 1,
						total:  2,
					},
				},
				dl: &scut.TestDetailLogger{},
			},
			want: 5,
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := createReturnValuesForGitHubActionsWorkflowPinned(tt.args.r, tt.args.infoMsg, tt.args.dl)
			if err != nil {
				t.Errorf("error during createReturnValuesForGitHubActionsWorkflowPinned: %v", err)
			}
			if got != tt.want {
				t.Errorf("createReturnValuesForGitHubActionsWorkflowPinned() = %v, want %v", got, tt.want)
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
			name: "download then run pinned debug",
			dependencies: []checker.Dependency{
				{
					Location: &checker.File{
						Path: "Dockerfile",
					},
					Msg:    asStringPointer("some message"),
					Type:   checker.DependencyUseTypeDownloadThenRun,
					Pinned: asBoolPointer(true),
				},
			},
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  9,
				NumberOfDebug: 1,
			},
		},
		{
			name: "download then run pinned debug and warn",
			dependencies: []checker.Dependency{
				{
					Location: &checker.File{
						Path: "Dockerfile",
					},
					Msg:    asStringPointer("some message"),
					Type:   checker.DependencyUseTypeDownloadThenRun,
					Pinned: asBoolPointer(false),
				},
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
				Score:         8,
				NumberOfWarn:  1,
				NumberOfInfo:  8,
				NumberOfDebug: 1,
			},
		},
		{
			name: "various warnings",
			dependencies: []checker.Dependency{
				{
					Location: &checker.File{
						Path: "Dockerfile",
					},
					Type:   checker.DependencyUseTypePipCommand,
					Pinned: asBoolPointer(false),
				},
				{
					Location: &checker.File{
						Path: "Dockerfile",
					},
					Type:   checker.DependencyUseTypeDownloadThenRun,
					Pinned: asBoolPointer(false),
				},
				{
					Location: &checker.File{
						Path: "Dockerfile",
					},
					Type:   checker.DependencyUseTypeDockerfileContainerImage,
					Pinned: asBoolPointer(false),
				},
				{
					Location: &checker.File{
						Path: "Dockerfile",
					},
					Msg:    asStringPointer("debug message"),
					Pinned: asBoolPointer(false),
				},
			},
			expected: scut.TestReturn{
				Error:         nil,
				Score:         6,
				NumberOfWarn:  3,
				NumberOfInfo:  6,
				NumberOfDebug: 1,
			},
		},
		{
			name: "download then run dockerfile and shell script",
			dependencies: []checker.Dependency{
				{
					Location: &checker.File{
						Path: "Dockerfile",
					},
					Type:   checker.DependencyUseTypeDownloadThenRun,
					Pinned: asBoolPointer(false),
				},
				{
					Location: &checker.File{
						Path: "script.sh",
					},
					Type:   checker.DependencyUseTypeDownloadThenRun,
					Pinned: asBoolPointer(false),
				},
			},
			expected: scut.TestReturn{
				Error:         nil,
				Score:         7,
				NumberOfWarn:  2,
				NumberOfInfo:  7,
				NumberOfDebug: 0,
			},
		},
		{
			name: "pip, npm, choco and go installs",
			dependencies: []checker.Dependency{
				{
					Location: &checker.File{},
					Type:     checker.DependencyUseTypePipCommand,
					Pinned:   asBoolPointer(false),
				},
				{
					Location: &checker.File{},
					Type:     checker.DependencyUseTypeNpmCommand,
					Pinned:   asBoolPointer(false),
				},
				{
					Location: &checker.File{},
					Type:     checker.DependencyUseTypeChocoCommand,
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
				Score:         5,
				NumberOfWarn:  4,
				NumberOfInfo:  5,
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
			name: "returns 10 if no error and no pinnedResult",
			args: args{
				t:  checker.DependencyUseTypePipCommand,
				dl: &scut.TestDetailLogger{},
			},
			want: 10,
		},
		{
			name: "returns 10 if no pinnedResult",
			args: args{
				t: checker.DependencyUseTypePipCommand,
				pr: map[checker.DependencyUseType]pinnedResult{
					checker.DependencyUseTypePipCommand: {
						pinned: 0,
						total:  0,
					},
				},
				dl: &scut.TestDetailLogger{},
			},
			want: 10,
		},
		{
			name: "returns 10 if pinned",
			args: args{
				t: checker.DependencyUseTypePipCommand,
				pr: map[checker.DependencyUseType]pinnedResult{
					checker.DependencyUseTypePipCommand: {
						pinned: 1,
						total:  1,
					},
				},
				dl: &scut.TestDetailLogger{},
			},
			want: 10,
		},
		{
			name: "returns 0 if unpinned",
			args: args{
				t: checker.DependencyUseTypePipCommand,
				pr: map[checker.DependencyUseType]pinnedResult{
					checker.DependencyUseTypePipCommand: {
						pinned: 0,
						total:  2,
					},
				},
				dl: &scut.TestDetailLogger{},
			},
			want: 0,
		},
		{
			name: "returns partial pinned",
			args: args{
				t: checker.DependencyUseTypeDownloadThenRun,
				pr: map[checker.DependencyUseType]pinnedResult{
					checker.DependencyUseTypeDownloadThenRun: {
						pinned: 1,
						total:  2,
					},
				},
				dl: &scut.TestDetailLogger{},
			},
			want: 5,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := createReturnValues(tt.args.pr[tt.args.t], "some message", tt.args.dl)
			if err != nil {
				t.Errorf("error during createReturnValues: %v", err)
			}
			if got != tt.want {
				t.Errorf("createReturnValues() = %v, want %v", got, tt.want)
			}

			if tt.want < 10 {
				return
			}

			isExpectedLog := func(logMessage checker.LogMessage, logType checker.DetailType) bool {
				return logMessage.Text == "some message" && logType == checker.DetailInfo
			}
			if !scut.ValidateLogMessage(isExpectedLog, tt.args.dl) {
				t.Errorf("test failed: log message not present: %+v", "some message")
			}
		})
	}
}
