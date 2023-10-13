// Copyright 2022 OpenSSF Scorecard Authors
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

// TestBinaryArtifacts tests the binary artifacts check.
func TestBinaryArtifacts(t *testing.T) {
	t.Parallel()
	//nolint
	type args struct {
		dl   checker.DetailLogger
		r    *checker.BinaryArtifactData
		name string
	}
	tests := []struct {
		name    string
		args    args
		want    checker.CheckResult
		wantErr bool
	}{
		{
			name: "r nil",
			args: args{
				name: "test_binary_artifacts_check_pass",
				dl:   &scut.TestDetailLogger{},
			},
			wantErr: true,
		},
		{
			name: "no binary artifacts",
			args: args{
				name: "no binary artifacts",
				dl:   &scut.TestDetailLogger{},
				r:    &checker.BinaryArtifactData{},
			},
			want: checker.CheckResult{
				Score: checker.MaxResultScore,
			},
		},
		{
			name: "1 binary artifact",
			args: args{
				name: "no binary artifacts",
				dl:   &scut.TestDetailLogger{},
				r: &checker.BinaryArtifactData{
					Files: []checker.File{
						{
							Path: "test_binary_artifacts_check_pass",
							Snippet: `
								package main
								import "fmt"
								func main() {
									fmt.Println("Hello, world!")
								}i`,
							Offset: 0,
							Type:   0,
						},
					},
				},
			},
			want: checker.CheckResult{
				Score: 9,
			},
		},
		{
			name: "many binary artifact",
			args: args{
				name: "no binary artifacts",
				dl:   &scut.TestDetailLogger{},
				r: &checker.BinaryArtifactData{
					Files: []checker.File{
						{
							Path: "test_binary_artifacts_check_pass",
							Snippet: `
								package main
								import "fmt"
								func main() {
									fmt.Println("Hello, world!")
								}i`,
							Offset: 0,
							Type:   0,
						},
						{
							Path: "test_binary_artifacts_check_pass",
							Snippet: `
								package main
								import "fmt"
								func main() {
									fmt.Println("Hello, world!")
								}i`,
							Offset: 0,
							Type:   0,
						},
						{
							Path: "test_binary_artifacts_check_pass",
							Snippet: `
								package main
								import "fmt"
								func main() {
									fmt.Println("Hello, world!")
								}i`,
							Offset: 0,
							Type:   0,
						},
						{
							Path: "test_binary_artifacts_check_pass",
							Snippet: `
								package main
								import "fmt"
								func main() {
									fmt.Println("Hello, world!")
								}i`,
							Offset: 0,
							Type:   0,
						},
						{
							Path: "test_binary_artifacts_check_pass",
							Snippet: `
								package main
								import "fmt"
								func main() {
									fmt.Println("Hello, world!")
								}i`,
							Offset: 0,
							Type:   0,
						},
						{
							Path: "test_binary_artifacts_check_pass",
							Snippet: `
								package main
								import "fmt"
								func main() {
									fmt.Println("Hello, world!")
								}i`,
							Offset: 0,
							Type:   0,
						},
						{
							Path: "test_binary_artifacts_check_pass",
							Snippet: `
								package main
								import "fmt"
								func main() {
									fmt.Println("Hello, world!")
								}i`,
							Offset: 0,
							Type:   0,
						},
						{
							Path: "test_binary_artifacts_check_pass",
							Snippet: `
								package main
								import "fmt"
								func main() {
									fmt.Println("Hello, world!")
								}i`,
							Offset: 0,
							Type:   0,
						},
						{
							Path: "test_binary_artifacts_check_pass",
							Snippet: `
								package main
								import "fmt"
								func main() {
									fmt.Println("Hello, world!")
								}i`,
							Offset: 0,
							Type:   0,
						},
						{
							Path: "test_binary_artifacts_check_pass",
							Snippet: `
								package main
								import "fmt"
								func main() {
									fmt.Println("Hello, world!")
								}i`,
							Offset: 0,
							Type:   0,
						},
						{
							Path: "test_binary_artifacts_check_pass",
							Snippet: `
								package main
								import "fmt"
								func main() {
									fmt.Println("Hello, world!")
								}i`,
							Offset: 0,
							Type:   0,
						},
						{
							Path: "test_binary_artifacts_check_pass",
							Snippet: `
								package main
								import "fmt"
								func main() {
									fmt.Println("Hello, world!")
								}i`,
							Offset: 0,
							Type:   0,
						},
						{
							Path: "test_binary_artifacts_check_pass",
							Snippet: `
								package main
								import "fmt"
								func main() {
									fmt.Println("Hello, world!")
								}i`,
							Offset: 0,
							Type:   0,
						},
						{
							Path: "test_binary_artifacts_check_pass",
							Snippet: `
								package main
								import "fmt"
								func main() {
									fmt.Println("Hello, world!")
								}i`,
							Offset: 0,
							Type:   0,
						},
						{
							Path: "test_binary_artifacts_check_pass",
							Snippet: `
								package main
								import "fmt"
								func main() {
									fmt.Println("Hello, world!")
								}i`,
							Offset: 0,
							Type:   0,
						},
						{
							Path: "test_binary_artifacts_check_pass",
							Snippet: `
								package main
								import "fmt"
								func main() {
									fmt.Println("Hello, world!")
								}i`,
							Offset: 0,
							Type:   0,
						},
					},
				},
			},
			want: checker.CheckResult{
				Score: 0,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := BinaryArtifacts(tt.args.name, tt.args.dl, tt.args.r)
			if tt.wantErr {
				if got.Error == nil {
					t.Errorf("BinaryArtifacts() error = %v, wantErr %v", got.Error, tt.wantErr)
				}
			} else {
				if got.Score != tt.want.Score {
					t.Errorf("BinaryArtifacts() = %v, want %v", got.Score, tt.want.Score)
				}
			}
		})
	}
}
