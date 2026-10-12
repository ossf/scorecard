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

package raw

import (
	"bufio"
	"errors"
	"io"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/google/go-cmp/cmp"
	"go.uber.org/mock/gomock"

	"github.com/ossf/scorecard/v5/checker"
	mockrepo "github.com/ossf/scorecard/v5/clients/mockclients"
)

func TestCollectShellScriptInsecureDownloadsReader(t *testing.T) {
	t.Parallel()
	const download = "curl https://example.com/install.sh | bash\n"
	const dataSize = 1 << 20
	readErr := errors.New("fixture read failed")
	paddedScript := "#!/bin/bash\n" + strings.Repeat("# padding\n", 7000) + download
	tests := []struct {
		reader           io.Reader
		wantErr          error
		name             string
		filename         string
		content          string
		maxRead          int
		wantDependencies int
	}{
		{
			name: "large non-shell file", filename: "observations.dat",
			reader:  io.MultiReader(strings.NewReader("ordinary data\n"), io.LimitReader(repeatedDataReader{}, dataSize)),
			maxRead: bufio.MaxScanTokenSize,
		},
		{
			name: "long non-shell first line", filename: "observations.dat",
			reader: io.LimitReader(repeatedDataReader{}, dataSize), maxRead: bufio.MaxScanTokenSize,
		},
		{
			name: "unsupported shebang overrides extension", filename: "analysis.sh",
			reader:  io.MultiReader(strings.NewReader("#!/usr/bin/python3\n"), io.LimitReader(repeatedDataReader{}, dataSize)),
			maxRead: bufio.MaxScanTokenSize,
		},
		{
			name: "extensionless supported script", filename: "install",
			content: "#!/usr/bin/env bash\n" + download, wantDependencies: 1,
		},
		{
			name: "shell extension without shebang", filename: "install.sh",
			content: download, wantDependencies: 1,
		},
		{
			name: "download after inspection buffer", filename: "install",
			content: paddedScript, wantDependencies: 1,
		},
		{
			name: "long first line preserves extension fallback", filename: "install.sh",
			content:          "#" + strings.Repeat("x", bufio.MaxScanTokenSize) + "\n" + download,
			wantDependencies: 1,
		},
		{name: "empty shell file", filename: "empty.sh", content: ""},
		{name: "empty data file", filename: "empty.dat", content: ""},
		{
			name: "initial read error", filename: "file.dat",
			reader: iotest.ErrReader(readErr), wantErr: readErr,
		},
		{
			name: "error after script prefix", filename: "install.sh",
			reader:  io.MultiReader(strings.NewReader(strings.Repeat("# padding\n", 7000)), iotest.ErrReader(readErr)),
			wantErr: readErr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			source := tt.reader
			if source == nil {
				source = strings.NewReader(tt.content)
			}
			reader := &countingFileReader{Reader: source}
			repo := mockrepo.NewMockRepoClient(gomock.NewController(t))
			repo.EXPECT().ListFiles(gomock.Any()).Return([]string{tt.filename}, nil)
			repo.EXPECT().GetFileReader(tt.filename).Return(reader, nil)
			var got checker.PinningDependenciesData
			err := collectShellScriptInsecureDownloads(&checker.CheckRequest{RepoClient: repo}, &got)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
			if !reader.closed {
				t.Error("reader was not closed")
			}
			if tt.maxRead != 0 && reader.consumed > tt.maxRead {
				t.Errorf("read %d bytes of a non-shell file, want at most %d", reader.consumed, tt.maxRead)
			}
			t.Logf("bytes read: %d", reader.consumed)
			if err != nil {
				return
			}
			if len(got.Dependencies) != tt.wantDependencies {
				t.Fatalf("dependencies = %d, want %d", len(got.Dependencies), tt.wantDependencies)
			}
			if len(got.ProcessingErrors) != 0 {
				t.Fatalf("unexpected processing errors: %v", got.ProcessingErrors)
			}
			if tt.reader == nil {
				checkCollectedShellScript(t, tt.filename, tt.content, &got, reader.consumed)
			}
		})
	}
}

// Check complete findings, including offsets, against the existing byte-based
// validator to detect dropped or duplicated buffered data.
func checkCollectedShellScript(t *testing.T, filename, content string, got *checker.PinningDependenciesData, consumed int) {
	t.Helper()
	var want checker.PinningDependenciesData
	if _, err := validateShellScriptIsFreeOfInsecureDownloads(filename, []byte(content), &want); err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(&want, got); diff != "" {
		t.Errorf("findings changed (-want +got):\n%s", diff)
	}
	if consumed != len(content) {
		t.Errorf("script read %d bytes, want %d", consumed, len(content))
	}
}

type repeatedDataReader struct{}

func (repeatedDataReader) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = 'x'
	}
	return len(p), nil
}

type countingFileReader struct {
	io.Reader
	consumed int
	closed   bool
}

func (r *countingFileReader) Read(p []byte) (int, error) {
	n, err := r.Reader.Read(p)
	r.consumed += n
	return n, err
}
func (r *countingFileReader) Close() error { r.closed = true; return nil }
