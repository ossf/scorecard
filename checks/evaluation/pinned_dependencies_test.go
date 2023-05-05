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
					thirdParties: 1,
					gitHubOwned:  1,
				},
				dl: &scut.TestDetailLogger{},
			},
			want: 10,
		},
		{
			name: "github actions workflow pinned",
			args: args{
				r: worklowPinningResult{
					thirdParties: 2,
					gitHubOwned:  2,
				},
				dl: &scut.TestDetailLogger{},
			},
			want: 0,
		},
		{
			name: "error in github actions workflow pinned",
			args: args{
				r: worklowPinningResult{
					thirdParties: 2,
					gitHubOwned:  2,
				},
				dl: &scut.TestDetailLogger{},
			},
			want: 0,
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

func asPointer(s string) *string {
	return &s
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
					Location: &checker.File{},
					Msg:      asPointer("some message"),
					Type:     checker.DependencyUseTypeDownloadThenRun,
				},
			},
			expected: scut.TestReturn{
				Error:         nil,
				Score:         checker.MaxResultScore,
				NumberOfWarn:  0,
				NumberOfInfo:  7,
				NumberOfDebug: 1,
			},
		},
		{
			name: "download then run pinned debug and warn",
			dependencies: []checker.Dependency{
				{
					Location: &checker.File{},
					Msg:      asPointer("some message"),
					Type:     checker.DependencyUseTypeDownloadThenRun,
				},
				{
					Location: &checker.File{},
					Type:     checker.DependencyUseTypeDownloadThenRun,
				},
			},
			expected: scut.TestReturn{
				Error:         nil,
				Score:         6,
				NumberOfWarn:  1,
				NumberOfInfo:  5,
				NumberOfDebug: 1,
			},
		},
		{
			name: "various warnings",
			dependencies: []checker.Dependency{
				{
					Location: &checker.File{},
					Type:     checker.DependencyUseTypePipCommand,
				},
				{
					Location: &checker.File{},
					Type:     checker.DependencyUseTypeDownloadThenRun,
				},
				{
					Location: &checker.File{},
					Type:     checker.DependencyUseTypeDockerfileContainerImage,
				},
				{
					Location: &checker.File{},
					Msg:      asPointer("debug message"),
				},
			},
			expected: scut.TestReturn{
				Error:         nil,
				Score:         3,
				NumberOfWarn:  3,
				NumberOfInfo:  3,
				NumberOfDebug: 1,
			},
		},
		{
			name: "unpinned pip install",
			dependencies: []checker.Dependency{
				{
					Location: &checker.File{},
					Type:     checker.DependencyUseTypePipCommand,
				},
			},
			expected: scut.TestReturn{
				Error:         nil,
				Score:         8,
				NumberOfWarn:  1,
				NumberOfInfo:  6,
				NumberOfDebug: 0,
			},
		},
		{
			name: "undefined pip install",
			dependencies: []checker.Dependency{
				{
					Location: &checker.File{},
					Type:     checker.DependencyUseTypePipCommand,
					Msg:      asPointer("debug message"),
				},
			},
			expected: scut.TestReturn{
				Error:         nil,
				Score:         10,
				NumberOfWarn:  0,
				NumberOfInfo:  7,
				NumberOfDebug: 1,
			},
		},
		{
			name: "all dependencies pinned",
			expected: scut.TestReturn{
				Error:         nil,
				Score:         10,
				NumberOfWarn:  0,
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
				},
			},
			expected: scut.TestReturn{
				Error:         nil,
				Score:         8,
				NumberOfWarn:  1,
				NumberOfInfo:  6,
				NumberOfDebug: 0,
			},
		},
		{
			name: "undefined npm install",
			dependencies: []checker.Dependency{
				{
					Location: &checker.File{},
					Type:     checker.DependencyUseTypeNpmCommand,
					Msg:      asPointer("debug message"),
				},
			},
			expected: scut.TestReturn{
				Error:         nil,
				Score:         10,
				NumberOfWarn:  0,
				NumberOfInfo:  7,
				NumberOfDebug: 1,
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
				t:  checker.DependencyUseTypeDownloadThenRun,
				dl: &scut.TestDetailLogger{},
			},
			want: 10,
		},
		{
			name: "returns 10 if pinned undefined",
			args: args{
				t: checker.DependencyUseTypeDownloadThenRun,
				pr: map[checker.DependencyUseType]pinnedResult{
					checker.DependencyUseTypeDownloadThenRun: pinnedUndefined,
				},
				dl: &scut.TestDetailLogger{},
			},
			want: 10,
		},
		{
			name: "returns 10 if pinned",
			args: args{
				t: checker.DependencyUseTypeDownloadThenRun,
				pr: map[checker.DependencyUseType]pinnedResult{
					checker.DependencyUseTypeDownloadThenRun: pinned,
				},
				dl: &scut.TestDetailLogger{},
			},
			want: 10,
		},
		{
			name: "returns 0 if unpinned",
			args: args{
				t: checker.DependencyUseTypeDownloadThenRun,
				pr: map[checker.DependencyUseType]pinnedResult{
					checker.DependencyUseTypeDownloadThenRun: notPinned,
				},
				dl: &scut.TestDetailLogger{},
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := createReturnValues(tt.args.pr, tt.args.t, "some message", tt.args.dl)
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

func Test_maxScore(t *testing.T) {
	t.Parallel()
	type args struct {
		s1 int
		s2 int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "returns s1 if s1 is greater than s2",
			args: args{
				s1: 10,
				s2: 5,
			},
			want: 10,
		},
		{
			name: "returns s2 if s2 is greater than s1",
			args: args{
				s1: 5,
				s2: 10,
			},
			want: 10,
		},
		{
			name: "returns s1 if s1 is equal to s2",
			args: args{
				s1: 10,
				s2: 10,
			},
			want: 10,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := maxScore(tt.args.s1, tt.args.s2); got != tt.want {
				t.Errorf("maxScore() = %v, want %v", got, tt.want)
			}
		})
	}
}
