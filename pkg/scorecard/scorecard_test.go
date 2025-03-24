// Copyright 2020 OpenSSF Scorecard Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package scorecard

import (
	"context"
	"fmt"
	"io"
	"slices"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"sigs.k8s.io/release-utils/version"

	"github.com/ossf/scorecard/v5/checker"
	"github.com/ossf/scorecard/v5/clients"
	"github.com/ossf/scorecard/v5/clients/localdir"
	mockrepo "github.com/ossf/scorecard/v5/clients/mockclients"
	"github.com/ossf/scorecard/v5/finding"
	"github.com/ossf/scorecard/v5/log"
	"github.com/ossf/scorecard/v5/probes/fuzzed"
)

func Test_getRepoCommitHash(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		want    string
		wantErr bool
	}{
		{
			name:    "success",
			want:    "abcdef",
			wantErr: false,
		},
		{
			name:    "empty commit",
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			mockRepoClient := mockrepo.NewMockRepoClient(ctrl)
			defer ctrl.Finish()
			mockRepoClient.EXPECT().ListCommits().DoAndReturn(func() ([]clients.Commit, error) {
				if tt.want == "" {
					return []clients.Commit{}, nil
				}
				return []clients.Commit{
					{
						SHA: tt.want,
					},
				}, nil
			})

			got, err := getRepoCommitHash(mockRepoClient)
			if (err != nil) != tt.wantErr {
				t.Errorf("getRepoCommitHash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getRepoCommitHash() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getRepoCommitHashLocal(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		path    string
		want    string
		wantErr bool
	}{
		{
			name:    "local directory",
			path:    "testdata",
			want:    "unknown",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			logger := log.NewLogger(log.DebugLevel)
			localDirClient := localdir.CreateLocalDirClient(context.Background(), logger)
			localRepo, err := localdir.MakeLocalDirRepo("testdata")
			if err != nil {
				t.Errorf("MakeLocalDirRepo: %v", err)
				return
			}
			if err := localDirClient.InitRepo(localRepo, clients.HeadSHA, 0); err != nil {
				t.Errorf("InitRepo: %v", err)
				return
			}

			got, err := getRepoCommitHash(localDirClient)
			if (err != nil) != tt.wantErr {
				t.Errorf("getRepoCommitHash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getRepoCommitHash() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRun(t *testing.T) {
	t.Parallel()
	type args struct {
		uri       string
		commitSHA string
	}
	// These values depend on the environment,
	// so don't encode particular expectations
	// in the test:
	versionInfo := version.GetVersionInfo()
	tests := []struct {
		name    string
		args    args
		want    Result
		wantErr bool
	}{
		{
			name: "empty commits repos should return repo details but no checks",
			args: args{
				uri:       "github.com/ossf/scorecard",
				commitSHA: "",
			},
			want: Result{
				Repo: RepoInfo{
					Name: "github.com/ossf/scorecard",
				},
				Scorecard: ScorecardInfo{
					Version:   versionInfo.GitVersion,
					CommitSHA: versionInfo.GitCommit,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			mockRepoClient := mockrepo.NewMockRepoClient(ctrl)
			repo := mockrepo.NewMockRepo(ctrl)

			repo.EXPECT().URI().Return(tt.args.uri).AnyTimes()

			mockRepoClient.EXPECT().InitRepo(repo, tt.args.commitSHA, 0).Return(nil)

			mockRepoClient.EXPECT().Close().DoAndReturn(func() error {
				return nil
			})

			mockRepoClient.EXPECT().ListCommits().DoAndReturn(func() ([]clients.Commit, error) {
				if tt.args.commitSHA == "" {
					return []clients.Commit{}, nil
				}
				return []clients.Commit{
					{
						SHA: tt.args.commitSHA,
					},
				}, nil
			})
			got, err := Run(context.Background(), repo,
				WithCommitSHA(tt.args.commitSHA),
				WithRepoClient(mockRepoClient),
			)
			if (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			ignoreDate := cmpopts.IgnoreFields(Result{}, "Date")
			if !cmp.Equal(got, tt.want, ignoreDate) {
				t.Errorf("expected %v, got %v", got, cmp.Diff(tt.want, got, ignoreDate))
			}
		})
	}
}

func TestRun_WithProbes(t *testing.T) {
	t.Parallel()
	// These values depend on the environment,
	// so don't encode particular expectations
	// in the test:
	versionInfo := version.GetVersionInfo()
	type args struct {
		uri       string
		commitSHA string
		probes    []string
	}
	tests := []struct {
		files   []string
		name    string
		args    args
		want    Result
		wantErr bool
	}{
		{
			name: "empty commits repos should return repo details but no checks",
			args: args{
				uri:       "github.com/ossf/scorecard",
				commitSHA: "1a17bb812fb2ac23e9d09e86e122f8b67563aed7",
				probes:    []string{fuzzed.Probe},
			},
			want: Result{
				Repo: RepoInfo{
					Name:      "github.com/ossf/scorecard",
					CommitSHA: "1a17bb812fb2ac23e9d09e86e122f8b67563aed7",
				},
				RawResults: checker.RawResults{
					Metadata: checker.MetadataData{
						Metadata: map[string]string{
							"repository.defaultBranch": "main",
							"repository.host":          "github.com",
							"repository.name":          "ossf/scorecard",
							"repository.sha1":          "1a17bb812fb2ac23e9d09e86e122f8b67563aed7",
							"repository.uri":           "github.com/ossf/scorecard",
							"localPath":                "test_path",
						},
					},
				},
				Scorecard: ScorecardInfo{
					Version:   versionInfo.GitVersion,
					CommitSHA: versionInfo.GitCommit,
				},
				Findings: []finding.Finding{
					{
						Probe:   fuzzed.Probe,
						Outcome: finding.OutcomeFalse,
						Message: "no fuzzer integrations found",
						Remediation: &finding.Remediation{
							Effort: finding.RemediationEffortHigh,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Wrong probe",
			args: args{
				uri:       "github.com/ossf/scorecard",
				commitSHA: "1a17bb812fb2ac23e9d09e86e122f8b67563aed7",
				probes:    []string{"nonExistentProbe"},
			},
			want:    Result{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			mockRepoClient := mockrepo.NewMockRepoClient(ctrl)
			mockRepoClient.EXPECT().LocalPath().DoAndReturn(func() (string, error) {
				return "test_path", nil
			}).AnyTimes()
			repo := mockrepo.NewMockRepo(ctrl)

			repo.EXPECT().URI().Return(tt.args.uri).AnyTimes()
			repo.EXPECT().Host().Return("github.com").AnyTimes()

			mockRepoClient.EXPECT().InitRepo(repo, tt.args.commitSHA, 0).Return(nil)
			mockRepoClient.EXPECT().URI().Return(repo.URI()).AnyTimes()
			mockRepoClient.EXPECT().Close().DoAndReturn(func() error {
				return nil
			})

			mockRepoClient.EXPECT().ListCommits().DoAndReturn(func() ([]clients.Commit, error) {
				if tt.args.commitSHA == "" {
					return []clients.Commit{}, nil
				}
				return []clients.Commit{
					{
						SHA: tt.args.commitSHA,
					},
				}, nil
			})
			mockRepoClient.EXPECT().ListFiles(gomock.Any()).Return(tt.files, nil).AnyTimes()
			progLanguages := []clients.Language{
				{
					Name:     clients.Go,
					NumLines: 100,
				},
				{
					Name:     clients.Java,
					NumLines: 70,
				},
				{
					Name:     clients.Cpp,
					NumLines: 100,
				},
				{
					Name:     clients.Ruby,
					NumLines: 70,
				},
			}
			mockRepoClient.EXPECT().ListProgrammingLanguages().Return(progLanguages, nil).AnyTimes()

			mockRepoClient.EXPECT().GetDefaultBranchName().Return("main", nil).AnyTimes()
			mockOSSFuzzClient := mockrepo.NewMockRepoClient(ctrl)
			mockOSSFuzzClient.EXPECT().Search(gomock.Any()).Return(clients.SearchResponse{}, nil).AnyTimes()
			got, err := Run(context.Background(), repo,
				WithRepoClient(mockRepoClient),
				WithOSSFuzzClient(mockOSSFuzzClient),
				WithCommitSHA(tt.args.commitSHA),
				WithProbes(tt.args.probes),
			)
			if (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			ignoreRemediationText := cmpopts.IgnoreFields(finding.Remediation{}, "Text", "Markdown")
			ignoreDate := cmpopts.IgnoreFields(Result{}, "Date")
			ignoreUnexported := cmpopts.IgnoreUnexported(finding.Finding{})
			if !cmp.Equal(got, tt.want, ignoreDate, ignoreRemediationText, ignoreUnexported) {
				t.Errorf("expected %v, got %v", got, cmp.Diff(tt.want, got, ignoreDate,
					ignoreRemediationText, ignoreUnexported))
			}
		})
	}
}

func Test_findConfigFile(t *testing.T) {
	t.Parallel()

	//nolint:govet
	tests := []struct {
		locs      []string
		desc      string
		found     string
		wantFound bool
	}{
		{
			desc:      "scorecard.yml exists",
			locs:      []string{"scorecard.yml"},
			found:     "scorecard.yml",
			wantFound: true,
		},
		{
			desc:      "scorecard.yaml exists",
			locs:      []string{"scorecard.yaml"},
			found:     "scorecard.yaml",
			wantFound: true,
		},
		{
			desc:      ".scorecard.yml exists",
			locs:      []string{".scorecard.yml"},
			found:     ".scorecard.yml",
			wantFound: true,
		},
		{
			desc:      ".scorecard.yaml exists",
			locs:      []string{".scorecard.yaml"},
			found:     ".scorecard.yaml",
			wantFound: true,
		},
		{
			desc:      ".github/scorecard.yml exists",
			locs:      []string{".github/scorecard.yml"},
			found:     ".github/scorecard.yml",
			wantFound: true,
		},
		{
			desc:      ".github/scorecard.yaml exists",
			locs:      []string{".github/scorecard.yaml"},
			found:     ".github/scorecard.yaml",
			wantFound: true,
		},
		{
			desc:      "multiple configs exist",
			locs:      []string{"scorecard.yml", ".github/scorecard.yaml"},
			found:     "scorecard.yml",
			wantFound: true,
		},
		{
			desc:      "no config exists",
			locs:      []string{},
			found:     "",
			wantFound: false,
		},
	}

	// Define the order of locations to check
	checkOrder := []string{
		"scorecard.yml",
		"scorecard.yaml",
		".scorecard.yml",
		".scorecard.yaml",
		".github/scorecard.yml",
		".github/scorecard.yaml",
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.desc, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			mockRepoClient := mockrepo.NewMockRepoClient(ctrl)
			defer ctrl.Finish()

			// Set up expectations for each location in the correct order
			for _, loc := range checkOrder {
				if slices.Contains(tt.locs, loc) {
					mockRepoClient.EXPECT().
						GetFileReader(loc).
						Return(io.NopCloser(strings.NewReader("test config")), nil)
					// If this is the first matching file, we should stop here
					if loc == tt.found {
						break
					}
				} else {
					mockRepoClient.EXPECT().
						GetFileReader(loc).
						Return(nil, fmt.Errorf("file not found"))
				}
			}

			gotReader, gotLoc := findConfigFile(mockRepoClient)
			if tt.wantFound {
				if gotReader == nil {
					t.Error("findConfigFile() returned nil reader when file should exist")
				}
				if gotLoc != tt.found {
					t.Errorf("findConfigFile() location = %v, want %v", gotLoc, tt.found)
				}
			} else {
				if gotReader != nil {
					t.Error("findConfigFile() returned non-nil reader when no file should exist")
				}
				if gotLoc != "" {
					t.Errorf("findConfigFile() location = %v, want empty string", gotLoc)
				}
			}
		})
	}
}
