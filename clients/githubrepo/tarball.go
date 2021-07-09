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

package githubrepo

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-github/v32/github"
)

const (
	repoDir      = "repo*"
	repoFilename = "githubrepo*.tar.gz"
)

var (
	errTarballNotFound = errors.New("tarball not found")
	errZipSlip         = errors.New("ZipSlip path detected")
)

func extractAndValidateArchivePath(path, dest string) (string, error) {
	const splitLength = 2
	// The tarball will have a top-level directory which contains all the repository files.
	// Discard the directory and only keep the actual files.
	names := strings.SplitN(path, "/", splitLength)
	if len(names) < splitLength {
		log.Printf("Unable to split path: %s", path)
		return dest, nil
	}
	if names[1] == "" {
		return dest, nil
	}
	// Check for ZipSlip: https://snyk.io/research/zip-slip-vulnerability
	cleanpath := filepath.Join(dest, names[1])
	if !strings.HasPrefix(cleanpath, filepath.Clean(dest)+string(os.PathSeparator)) {
		return "", fmt.Errorf("%w: %s", errZipSlip, names[1])
	}
	return cleanpath, nil
}

type tarballHandler struct {
	tempDir     string
	tempTarFile string
	files       []string
}

func (handler *tarballHandler) init(ctx context.Context, repo *github.Repository) error {
	// Cleanup any previous state.
	if err := handler.cleanup(); err != nil {
		return fmt.Errorf("error during githubrepo cleanup: %w", err)
	}

	// Setup temp dir/files and download repo tarball.
	if err := handler.getTarball(ctx, repo); errors.Is(err, errTarballNotFound) {
		log.Printf("%v", err)
		return nil
	} else if err != nil {
		return fmt.Errorf("error getting githurepo tarball: %w", err)
	}

	// Extract file names and content from tarball.
	if err := handler.extractTarball(); err != nil {
		return fmt.Errorf("error extracting githubrepo tarball: %w", err)
	}

	return nil
}

func (handler *tarballHandler) getTarball(ctx context.Context, repo *github.Repository) error {
	url := repo.GetArchiveURL()
	url = strings.Replace(url, "{archive_format}", "tarball/", 1)
	url = strings.Replace(url, "{/ref}", "", 1)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("http.NewRequestWithContext: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("http.DefaultClient.Do: %w", err)
	}
	defer resp.Body.Close()

	// Handle 400/404 errors
	switch resp.StatusCode {
	case http.StatusNotFound, http.StatusBadRequest:
		return fmt.Errorf("%w: %s", errTarballNotFound, *repo.URL)
	}

	// Create a temp file. This automatically appends a random number to the name.
	tempDir, err := ioutil.TempDir("", repoDir)
	if err != nil {
		return fmt.Errorf("error creating TempDir in githubrepo: %w", err)
	}
	repoFile, err := ioutil.TempFile(tempDir, repoFilename)
	if err != nil {
		return fmt.Errorf("error during ioutil.TempFile in githubrepo: %w", err)
	}
	defer repoFile.Close()
	if _, err := io.Copy(repoFile, resp.Body); err != nil {
		return fmt.Errorf("error during io.Copy in githubrepo tarball: %w", err)
	}

	handler.tempDir = tempDir
	handler.tempTarFile = repoFile.Name()
	return nil
}

// nolint: gocognit
func (handler *tarballHandler) extractTarball() error {
	// nolint: gomnd
	in, err := os.OpenFile(handler.tempTarFile, os.O_RDONLY, 0o644)
	if err != nil {
		return fmt.Errorf("error opening %s: %w", handler.tempTarFile, err)
	}
	gz, err := gzip.NewReader(in)
	if err != nil {
		return fmt.Errorf("error reading %s: %w", handler.tempTarFile, err)
	}
	tr := tar.NewReader(gz)
	for {
		header, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("error in tarReader.Next(): %w", err)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			dirpath, err := extractAndValidateArchivePath(header.Name, handler.tempDir)
			if err != nil {
				return fmt.Errorf("error extracting dirpath: %w", err)
			}
			if dirpath == filepath.Clean(handler.tempDir) {
				continue
			}
			// nolint: gomnd
			if err := os.Mkdir(dirpath, 0o755); err != nil {
				return fmt.Errorf("error during os.Mkdir: %w", err)
			}
		case tar.TypeReg:
			if header.Size <= 0 {
				continue
			}
			filenamepath, err := extractAndValidateArchivePath(header.Name, handler.tempDir)
			if err != nil {
				return fmt.Errorf("error extracting file path: %w", err)
			}

			if _, err := os.Stat(filepath.Dir(filenamepath)); os.IsNotExist(err) {
				// nolint: gomnd
				if err := os.Mkdir(filepath.Dir(filenamepath), 0o755); err != nil {
					return fmt.Errorf("error during os.Mkdir: %w", err)
				}
			}
			outFile, err := os.Create(filenamepath)
			if err != nil {
				return fmt.Errorf("error during os.Create: %w", err)
			}

			// nolint: gosec
			// Potential for DoS vulnerability via decompression bomb.
			// Since such an attack will only impact a single shard, ignoring this for now.
			if _, err := io.Copy(outFile, tr); err != nil {
				return fmt.Errorf("error during io.Copy: %w", err)
			}
			outFile.Close()
			handler.files = append(handler.files,
				strings.TrimPrefix(filenamepath, filepath.Clean(handler.tempDir)+string(os.PathSeparator)))
		case tar.TypeXGlobalHeader, tar.TypeSymlink:
			continue
		default:
			log.Printf("Unknown file type %s: '%s'", header.Name, string(header.Typeflag))
			continue
		}
	}
	return nil
}

func (handler *tarballHandler) listFiles(predicate func(string) bool) []string {
	ret := make([]string, 0)
	for _, file := range handler.files {
		if predicate(file) {
			ret = append(ret, file)
		}
	}
	return ret
}

func (handler *tarballHandler) getFileContent(filename string) ([]byte, error) {
	content, err := ioutil.ReadFile(filepath.Join(handler.tempDir, filename))
	if err != nil {
		return content, fmt.Errorf("error trying to ReadFile: %w", err)
	}
	return content, nil
}

func (handler *tarballHandler) cleanup() error {
	if err := os.RemoveAll(handler.tempDir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("os.Remove: %w", err)
	}
	// Remove old files so we don't iterate through them.
	handler.files = nil
	return nil
}
