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

package fileparser

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/ossf/scorecard/v4/checker"
	mockrepo "github.com/ossf/scorecard/v4/clients/mockclients"
)

func TestIsTemplateFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		filename   string
		isTemplate bool
	}{
		{
			filename:   "Dockerfile.template",
			isTemplate: true,
		},
		{
			filename:   "Dockerfile.tmpl",
			isTemplate: true,
		},
		{
			filename:   "Dockerfile.template-debian",
			isTemplate: true,
		},
		{
			filename:   "Dockerfile.tmpl.",
			isTemplate: true,
		},
		{
			filename:   "Dockerfile-template",
			isTemplate: true,
		},
		{
			filename:   "tmpl.Dockerfile",
			isTemplate: true,
		},
		{
			filename:   "template.Dockerfile",
			isTemplate: true,
		},
		{
			filename:   "Dockerfile_template",
			isTemplate: true,
		},
		{
			filename:   "Dockerfile.tmpl.prod",
			isTemplate: true,
		},
		{
			filename:   "Dockerfile.Template",
			isTemplate: true,
		},
		{
			filename:   "dockerfile.tpl",
			isTemplate: true,
		},
		{
			filename:   "build/Dockerfile.tpl",
			isTemplate: true,
		},
		{
			filename:   "build/tpl.Dockerfile",
			isTemplate: true,
		},
		{
			filename:   "DockerfileTemplate",
			isTemplate: false,
		},
		{
			filename:   "Dockerfile.linux",
			isTemplate: false,
		},
		{
			filename:   "tmp.Dockerfile",
			isTemplate: false,
		},
		{
			filename:   "Dockerfile",
			isTemplate: false,
		},
		{
			filename:   "Dockerfile.temp.late",
			isTemplate: false,
		},
		{
			filename:   "Dockerfile.temp",
			isTemplate: false,
		},
		{
			filename:   "template/Dockerfile",
			isTemplate: false,
		},
		{
			filename:   "linux.Dockerfile",
			isTemplate: false,
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.filename, func(t *testing.T) {
			t.Parallel()
			if got := IsTemplateFile(tt.filename); got != tt.isTemplate {
				t.Errorf("%v: Got (%v) expected (%v)", tt.filename, got, tt.isTemplate)
			}
		})
	}
}

// TestCheckFileContainsCommands tests if the content starts with a comment.
func TestCheckFileContainsCommands(t *testing.T) {
	t.Parallel()
	//nolint
	type args struct {
		content []byte
		comment string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Dockerfile.template",
			args: args{
				content: []byte(`FROM golang:1.12.4`),
			},
			want: false,
		},
		{
			name: "Dockerfile.template with a comment",
			args: args{
				content: []byte(`# This is a comment
				FROM golang:1.12.4`),
				// start with a comment
				comment: "#",
			},
			want: true,
		},
		{
			name: "empty file",
			want: false,
		},
	}

	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := CheckFileContainsCommands(tt.args.content, tt.args.comment); got != tt.want {
				t.Errorf("CheckFileContainsCommands() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isMatchingPath(t *testing.T) {
	t.Parallel()
	type args struct {
		pattern       string
		fullpath      string
		caseSensitive bool
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "matching path",
			args: args{
				pattern:       "Dockerfile",
				fullpath:      "Dockerfile",
				caseSensitive: true,
			},
			want: true,
		},
		{
			name: "matching path with case insensitive",
			args: args{
				pattern:       "Dockerfile",
				fullpath:      "dockerfile",
				caseSensitive: false,
			},
			want: true,
		},
		{
			name: "matching path with case insensitive",
			args: args{
				pattern:       "Dockerfile",
				fullpath:      "dockerfile",
				caseSensitive: true,
			},
			want: false,
		},
		{
			name: "matching path with case insensitive",
			args: args{
				pattern:       "Dockerfile",
				fullpath:      "Dockerfile.template",
				caseSensitive: false,
			},
			want: false,
		},
		{
			name: "matching path with case insensitive",
			args: args{
				pattern:       "Dockerfile",
				fullpath:      "Dockerfile.template",
				caseSensitive: true,
			},
			want: false,
		},
		{
			name: "matching path with case insensitive",
			args: args{
				pattern:       "Dockerfile",
				fullpath:      "Dockerfile.template",
				caseSensitive: false,
			},
			want: false,
		},
		{
			name: "matching path with case insensitive",
			args: args{
				pattern:       "Dockerfile",
				fullpath:      "Dockerfile.template",
				caseSensitive: true,
			},
		},
		{
			name: "matching path with case insensitive",
			args: args{
				pattern:       "Dockerfile",
				fullpath:      "Dockerfile.template",
				caseSensitive: false,
			},
		},
		{
			name: "matching path with case insensitive",
			args: args{
				pattern:       "Dockerfile",
				fullpath:      "Dockerfile.template",
				caseSensitive: true,
			},
		},
		{
			name: "matching path with case insensitive",
			args: args{
				pattern:       "Dockerfile",
				fullpath:      "Dockerfile.template",
				caseSensitive: false,
			},
		},
		{
			name: "matching path with case insensitive",
			args: args{
				pattern:       "Dockerfile",
				fullpath:      "Dockerfile.template",
				caseSensitive: true,
			},
		},
		{
			name: "matching path with case insensitive",
			args: args{
				pattern:       "Dockerfile",
				fullpath:      "Dockerfile.template",
				caseSensitive: false,
			},
		},
		{
			name: "matching path with case insensitive",
			args: args{
				pattern:       "Dockerfile",
				fullpath:      "Dockerfile.template",
				caseSensitive: true,
			},
		},
		{
			name: "matching path with case insensitive",
			args: args{
				pattern:       "Dockerfile",
				fullpath:      "Dockerfile.template",
				caseSensitive: false,
			},
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := isMatchingPath(tt.args.pattern, tt.args.fullpath, tt.args.caseSensitive)
			if (err != nil) != tt.wantErr {
				t.Errorf("isMatchingPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("isMatchingPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isTestdataFile(t *testing.T) {
	t.Parallel()
	type args struct {
		fullpath string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "testdata file",
			args: args{
				fullpath: "testdata/Dockerfile",
			},
			want: true,
		},
		{
			name: "testdata file",
			args: args{
				fullpath: "testdata/Dockerfile.template",
			},
			want: true,
		},
		{
			name: "testdata file",
			args: args{
				fullpath: "testdata/Dockerfile.template.template",
			},
			want: true,
		},
		{
			name: "testdata file",
			args: args{
				fullpath: "testdata/Dockerfile.template.template.template",
			},
			want: true,
		},
		{
			name: "testdata file",
			args: args{
				fullpath: "testdata/Dockerfile.template.template.template.template",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := isTestdataFile(tt.args.fullpath); got != tt.want {
				t.Errorf("isTestdataFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestFileGetCbDataAsBoolPointer tests the FileGetCbDataAsBoolPointer function.
func TestFileGetCbDataAsBoolPointer(t *testing.T) {
	t.Parallel()
	type args struct {
		data FileCbData
	}
	b := true
	//nolint
	tests := []struct {
		name      string
		args      args
		want      *bool
		wantPanic bool
	}{
		{
			name: "true",
			args: args{
				data: &b,
			},
			want: &b,
		},
		{
			name:      "nil",
			args:      args{},
			want:      &b,
			wantPanic: true,
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.wantPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("FileGetCbDataAsBoolPointer() did not panic")
					}
				}()
				FileGetCbDataAsBoolPointer(tt.args.data)
				return
			}
			if got := FileGetCbDataAsBoolPointer(tt.args.data); got != tt.want {
				t.Errorf("FileGetCbDataAsBoolPointer() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestCheckFilesContentV6 tests the CheckFilesContentV6 function.
func TestCheckFilesContentV6(t *testing.T) {
	t.Parallel()
	//nolint
	tests := []struct {
		name                   string
		wantErr                bool
		shellPattern           string
		caseSensitive          bool
		shouldFuncFail         bool
		shouldGetPredicateFail bool
		files                  []string
	}{
		{
			name:          "no files",
			shellPattern:  "Dockerfile",
			caseSensitive: true,
			files: []string{
				"Dockerfile",
				"Dockerfile.template",
				"Dockerfile.template.template",
				"Dockerfile.template.template.template",
				"Dockerfile.template.template.template.template",
			},
		},
		{
			name:          "no files",
			shellPattern:  "Dockerfile",
			caseSensitive: false,
			files:         []string{},
		},
		{
			name:          "no files",
			shellPattern:  "Dockerfile",
			caseSensitive: true,
			files:         []string{},
		},
		{
			name:          "no files",
			shellPattern:  "Dockerfile",
			caseSensitive: false,
			files: []string{
				"Dockerfile",
				"Dockerfile.template",
				"Dockerfile.template.template",
				"Dockerfile.template.template.template",
				"Dockerfile.template.template.template.template",
			},
			shouldFuncFail: true,
		},
		{
			name:                   "no files",
			shellPattern:           "Dockerfile",
			caseSensitive:          false,
			shouldGetPredicateFail: true,
			files: []string{
				"Dockerfile",
				"Dockerfile.template",
				"Dockerfile.template.template",
				"Dockerfile.template.template.template",
				"Dockerfile.template.template.template.template",
			},
		},
	}

	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			x := func(path string, content []byte, data FileCbData) (bool, error) {
				if tt.shouldFuncFail {
					//nolint
					return false, errors.New("test error")
				}
				if tt.shouldGetPredicateFail {
					return false, nil
				}
				return true, nil
			}

			ctrl := gomock.NewController(t)
			mockRepo := mockrepo.NewMockRepoClient(ctrl)
			mockRepo.EXPECT().ListFiles(gomock.Any()).Return(tt.files, nil).AnyTimes()
			mockRepo.EXPECT().GetFileContent(gomock.Any()).Return(nil, nil).AnyTimes()

			result := CheckFilesContentV6(tt.shellPattern, tt.caseSensitive, mockRepo, x, x)

			if tt.wantErr && result == nil {
				t.Errorf("CheckFilesContentV6() = %v, want %v test name %v", result, tt.wantErr, tt.name)
			}
		})
	}
}

// TestCheckFilesContent tests the CheckFilesContent function.
func TestCheckFilesContent(t *testing.T) {
	t.Parallel()
	//nolint
	tests := []struct {
		name                   string
		wantErr                bool
		shellPattern           string
		caseSensitive          bool
		shouldFuncFail         bool
		shouldGetPredicateFail bool
		files                  []string
	}{
		{
			name:          "no files",
			shellPattern:  "Dockerfile",
			caseSensitive: true,
			files: []string{
				"Dockerfile",
				"Dockerfile.template",
				"Dockerfile.template.template",
				"Dockerfile.template.template.template",
				"Dockerfile.template.template.template.template",
			},
		},
		{
			name:          "no files",
			shellPattern:  "Dockerfile",
			caseSensitive: false,
			files:         []string{},
		},
		{
			name:          "no files",
			shellPattern:  "Dockerfile",
			caseSensitive: true,
			files:         []string{},
		},
		{
			name:          "no files",
			shellPattern:  "Dockerfile",
			caseSensitive: false,
			files: []string{
				"Dockerfile",
				"Dockerfile.template",
				"Dockerfile.template.template",
				"Dockerfile.template.template.template",
				"Dockerfile.template.template.template.template",
			},
			shouldFuncFail: true,
		},
		{
			name:                   "no files",
			shellPattern:           "Dockerfile",
			caseSensitive:          false,
			shouldGetPredicateFail: true,
			files: []string{
				"Dockerfile",
				"Dockerfile.template",
				"Dockerfile.template.template",
				"Dockerfile.template.template.template",
				"Dockerfile.template.template.template.template",
			},
		},
	}

	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			x := func(path string, content []byte,
				dl checker.DetailLogger, data FileCbData) (bool, error) {
				if tt.shouldFuncFail {
					//nolint
					return false, errors.New("test error")
				}
				if tt.shouldGetPredicateFail {
					return false, nil
				}
				return true, nil
			}

			ctrl := gomock.NewController(t)
			mockRepo := mockrepo.NewMockRepoClient(ctrl)
			mockRepo.EXPECT().ListFiles(gomock.Any()).Return(tt.files, nil).AnyTimes()
			mockRepo.EXPECT().GetFileContent(gomock.Any()).Return(nil, nil).AnyTimes()

			c := checker.CheckRequest{
				RepoClient: mockRepo,
			}

			result := CheckFilesContent(tt.shellPattern, tt.caseSensitive, &c, x, x)

			if tt.wantErr && result == nil {
				t.Errorf("CheckFilesContentV6() = %v, want %v test name %v", result, tt.wantErr, tt.name)
			}
		})
	}
}

// TestCheckFilesContentV6 tests the CheckFilesContentV6 function.
func TestCheckIfFileExistsV6(t *testing.T) {
	t.Parallel()
	//nolint
	type args struct {
		cbReturn             bool
		cbwantErr            bool
		listFilesReturnError error
	}
	//nolint
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "cb true and no error",
			args: args{
				cbReturn:             true,
				cbwantErr:            false,
				listFilesReturnError: nil,
			},
			wantErr: false,
		},
		{
			name: "cb false and no error",
			args: args{
				cbReturn:             false,
				cbwantErr:            false,
				listFilesReturnError: nil,
			},
			wantErr: false,
		},
		{
			name: "cb wantErr and error",
			args: args{
				cbReturn:             true,
				cbwantErr:            true,
				listFilesReturnError: nil,
			},
			wantErr: true,
		},
		{
			name: "listFilesReturnError and error",
			args: args{
				cbReturn:  true,
				cbwantErr: true,
				//nolint
				listFilesReturnError: errors.New("test error"),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			x := func(path string, data FileCbData) (bool, error) {
				if tt.args.cbwantErr {
					//nolint
					return false, errors.New("test error")
				}
				return tt.args.cbReturn, nil
			}

			ctrl := gomock.NewController(t)
			mockRepo := mockrepo.NewMockRepoClient(ctrl)
			mockRepo.EXPECT().ListFiles(gomock.Any()).Return([]string{"foo"}, nil).AnyTimes()

			err := CheckIfFileExistsV6(mockRepo, x, x)

			if (err != nil) != tt.wantErr {
				t.Errorf("CheckIfFileExistsV6() error = %v, wantErr %v for %v", err, tt.wantErr, tt.name)
				return
			}
		})
	}
}

// TestCheckIfFileExists tests the CheckIfFileExists function.
func TestCheckIfFileExists(t *testing.T) {
	t.Parallel()
	//nolint
	type args struct {
		cbReturn bool
		cbErr    error
	}
	//nolint
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "cb true and no error",
			args: args{
				cbReturn: true,
				cbErr:    nil,
			},
			wantErr: false,
		},
		{
			name: "cb false and no error",
			args: args{
				cbReturn: false,
				cbErr:    nil,
			},
			wantErr: false,
		},
		{
			name: "cb error",
			args: args{
				cbReturn: true,
				cbErr:    errors.New("test error"),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			mockRepo := mockrepo.NewMockRepoClient(ctrl)
			mockRepo.EXPECT().ListFiles(gomock.Any()).Return([]string{"foo"}, nil).AnyTimes()
			c := checker.CheckRequest{
				RepoClient: mockRepo,
			}
			x := func(path string,
				dl checker.DetailLogger, data FileCbData) (bool, error) {
				return tt.args.cbReturn, tt.args.cbErr
			}

			err := CheckIfFileExists(&c, x, x)

			if (err != nil) != tt.wantErr {
				t.Errorf("CheckIfFileExists() error = %v, wantErr %v for %v", err, tt.wantErr, tt.name)
				return
			}
		})
	}
}
