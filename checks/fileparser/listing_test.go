// Copyright 2021 OpenSSF Scorecard Authors
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

	mockrepo "github.com/ossf/scorecard/v4/clients/mockclients"
)

var (
	errInvalidArgType   = errors.New("invalid arg type")
	errInvalidArgLength = errors.New("invalid arg length")
	errTest             = errors.New("test")
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
			got, err := isMatchingPath(tt.args.fullpath, PathMatcher{
				Pattern:       tt.args.pattern,
				CaseSensitive: tt.args.caseSensitive,
			})
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
		{
			name: "testdata file",
			args: args{
				fullpath: "archiva-modules/archiva-base/archiva-checksum/src/test/resources/examples/redback-authz-open.jar",
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

// TestOnMatchingFileContentDo tests the OnMatchingFileContent function.
func TestOnMatchingFileContent(t *testing.T) {
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
			x := func(path string, content []byte, args ...interface{}) (bool, error) {
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

			result := OnMatchingFileContentDo(mockRepo, PathMatcher{
				Pattern:       tt.shellPattern,
				CaseSensitive: tt.caseSensitive,
			}, x)

			if tt.wantErr && result == nil {
				t.Errorf("OnMatchingFileContentDo() = %v, want %v test name %v", result, tt.wantErr, tt.name)
			}
		})
	}
}

// TestOnAllFilesDo tests the OnAllFilesDo function.
//
//nolint:gocognit
func TestOnAllFilesDo(t *testing.T) {
	t.Parallel()

	type testArgsFn func(args ...interface{}) bool
	validateCountIs := func(count int) testArgsFn {
		return func(args ...interface{}) bool {
			if len(args) == 0 {
				return false
			}
			val, ok := args[0].(*int)
			if !ok {
				return false
			}
			return val != nil && *val == count
		}
	}

	incrementCount := func(path string, args ...interface{}) (bool, error) {
		if len(args) < 1 {
			return false, errInvalidArgLength
		}
		val, ok := args[0].(*int)
		if !ok || val == nil {
			return false, errInvalidArgType
		}
		(*val)++
		if len(args) > 1 {
			maxVal, ok := args[1].(int)
			if !ok {
				return false, errInvalidArgType
			}
			if *val >= maxVal {
				return false, nil
			}
		}
		return true, nil
	}
	alwaysFail := func(path string, args ...interface{}) (bool, error) {
		return false, errTest
	}
	//nolint
	tests := []struct {
		name         string
		onFile       DoWhileTrueOnFilename
		onFileArgs   []interface{}
		listFiles    []string
		errListFiles error
		err          error
		testArgs     testArgsFn
	}{
		{
			name:         "error during ListFiles",
			errListFiles: errTest,
			err:          errTest,
			onFile:       alwaysFail,
		},
		{
			name:       "empty ListFiles",
			listFiles:  []string{},
			onFile:     incrementCount,
			onFileArgs: []interface{}{new(int)},
			testArgs:   validateCountIs(0),
		},
		{
			name:       "onFile true and no error",
			listFiles:  []string{"foo", "bar"},
			onFile:     incrementCount,
			onFileArgs: []interface{}{new(int)},
			testArgs:   validateCountIs(2),
		},
		{
			name:       "onFile false and no error",
			listFiles:  []string{"foo", "bar"},
			onFile:     incrementCount,
			onFileArgs: []interface{}{new(int), 1 /*maxVal*/},
			testArgs:   validateCountIs(1),
		},
		{
			name:      "onFile has error",
			listFiles: []string{"foo", "bar"},
			onFile:    alwaysFail,
			err:       errTest,
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			mockRepoClient := mockrepo.NewMockRepoClient(ctrl)
			mockRepoClient.EXPECT().ListFiles(gomock.Any()).
				Return(tt.listFiles, tt.errListFiles).AnyTimes()

			err := OnAllFilesDo(mockRepoClient, tt.onFile, tt.onFileArgs...)
			if !errors.Is(err, tt.err) {
				t.Errorf("OnAllFilesDo() expected error = %v, got %v", tt.err, err)
			}
			if tt.testArgs != nil && !tt.testArgs(tt.onFileArgs...) {
				t.Error("testArgs validation failed")
			}
		})
	}
}
