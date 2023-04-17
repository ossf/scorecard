// Copyright 2021 Security Scorecard Authors
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

package gitlabrepo

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/xanzy/go-gitlab"

	sce "github.com/ossf/scorecard/v4/errors"
)

const (
	repoDir      = "project*"
	repoFilename = "gitlabproject*.tar.gz"
)

var (
	errTarballNotFound  = errors.New("tarball not found")
	errTarballCorrupted = errors.New("corrupted tarball")
	errZipSlip          = errors.New("ZipSlip path detected")
)

func extractAndValidateArchivePath(path, dest string) (string, error) {
	const splitLength = 2
	// The tarball will have a top-level directory which contains all the repository files.
	// Discard the directory and only keep the actual files.
	names := strings.SplitN(path, "/", splitLength)
	if len(names) < splitLength {
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
	errSetup    error
	once        *sync.Once
	ctx         context.Context
	repo        *gitlab.Project
	repourl     *repoURL
	commitSHA   string
	tempDir     string
	tempTarFile string
	files       []string
}

func (handler *tarballHandler) init(ctx context.Context, repourl *repoURL, repo *gitlab.Project, commitSHA string) {
	handler.errSetup = nil
	handler.once = new(sync.Once)
	handler.ctx = ctx
	handler.repo = repo
	handler.repourl = repourl
	handler.commitSHA = commitSHA
}

func (handler *tarballHandler) setup() error {
	handler.once.Do(func() {
		// cleanup any previous state.
		if err := handler.cleanup(); err != nil {
			handler.errSetup = sce.WithMessage(sce.ErrScorecardInternal, err.Error())
			return
		}

		// setup tem dir/files and download repo tarball.
		if err := handler.getTarball(); errors.Is(err, errTarballNotFound) {
			log.Printf("unable to get tarball %v. Skipping...", err)
			return
		} else if err != nil {
			handler.errSetup = sce.WithMessage(sce.ErrScorecardInternal, err.Error())
			return
		}

		// extract file names and content from tarball.
		if err := handler.extractTarball(); errors.Is(err, errTarballCorrupted) {
			log.Printf("unable to extract tarball %v. Skipping...", err)
		} else if err != nil {
			handler.errSetup = sce.WithMessage(sce.ErrScorecardInternal, err.Error())
		}
	})
	return handler.errSetup
}

func (handler *tarballHandler) getTarball() error {
	url := fmt.Sprintf("%s/api/v4/projects/%d/repository/archive.tar.gz?sha=%s",
		handler.repourl.Host(), handler.repo.ID, handler.commitSHA)
	req, err := http.NewRequestWithContext(handler.ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("http.NewRequestWithContext: %w", err)
	}
	req.Header.Set("PRIVATE-TOKEN", os.Getenv("GITLAB_AUTH_TOKEN"))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("http.DefaultClient.Do: %w", err)
	}
	defer resp.Body.Close()

	// Handler 400/404 errors.
	switch resp.StatusCode {
	case http.StatusNotFound, http.StatusBadRequest:
		return fmt.Errorf("%w: %s", errTarballNotFound, url)
	}

	// Create a temp file.  This automatically appends a random number to the name.
	tempDir, err := os.MkdirTemp("", repoDir)
	if err != nil {
		return fmt.Errorf("os.MkdirTemp: %w", err)
	}
	repoFile, err := os.CreateTemp(tempDir, repoFilename)
	if err != nil {
		return fmt.Errorf("os.CreateTemp: %w", err)
	}
	defer repoFile.Close()
	if _, err := io.Copy(repoFile, resp.Body); err != nil {
		// If the incomming tarball is corrupted or the server times out.
		return fmt.Errorf("%w io.Copy: %v", errTarballNotFound, err)
	}

	handler.tempDir = tempDir
	handler.tempTarFile = repoFile.Name()
	return nil
}

// nolint: gocognit
func (handler *tarballHandler) extractTarball() error {
	in, err := os.OpenFile(handler.tempTarFile, os.O_RDONLY, 0o644)
	if err != nil {
		return fmt.Errorf("os.OpenFile: %w", err)
	}
	gz, err := gzip.NewReader(in)
	if err != nil {
		return fmt.Errorf("%w: gzip.NewReader %v %v", errTarballCorrupted, handler.tempTarFile, err)
	}
	tr := tar.NewReader(gz)
	for {
		header, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("%w tarReader.Next: %v", errTarballCorrupted, err)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			dirpath, err := extractAndValidateArchivePath(header.Name, handler.tempDir)
			if err != nil {
				return err
			}
			if dirpath == filepath.Clean(handler.tempDir) {
				continue
			}

			if err := os.Mkdir(dirpath, 0o755); err != nil {
				return fmt.Errorf("error during os.Mkdir: %w", err)
			}
		case tar.TypeReg:
			if header.Size <= 0 {
				continue
			}
			filenamepath, err := extractAndValidateArchivePath(header.Name, handler.tempDir)
			if err != nil {
				return err
			}

			if _, err := os.Stat(filepath.Dir(filenamepath)); os.IsNotExist(err) {
				if err := os.Mkdir(filepath.Dir(filenamepath), 0o755); err != nil {
					return fmt.Errorf("os.Mkdir: %w", err)
				}
			}
			outFile, err := os.Create(filenamepath)
			if err != nil {
				return fmt.Errorf("os.Create: %w", err)
			}

			//nolint: gosec
			// Potential for DoS vulnerability via decompression bomb.
			// Since such an attack will only impact a single shard, ignoring this for now.
			if _, err := io.Copy(outFile, tr); err != nil {
				return fmt.Errorf("%w io.Copy: %v", errTarballCorrupted, err)
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

func (handler *tarballHandler) listFiles(predicate func(string) (bool, error)) ([]string, error) {
	if err := handler.setup(); err != nil {
		return nil, fmt.Errorf("error during tarballHandler.setup: %w", err)
	}
	ret := make([]string, 0)
	for _, file := range handler.files {
		matches, err := predicate(file)
		if err != nil {
			return nil, err
		}
		if matches {
			ret = append(ret, file)
		}
	}
	return ret, nil
}

func (handler *tarballHandler) getFileContent(filename string) ([]byte, error) {
	if err := handler.setup(); err != nil {
		fmt.Printf("err: %v\n", err)
		return nil, fmt.Errorf("error during tarballHandler.setup: %w", err)
	}
	fmt.Printf("handler.tempDir: %v\n", handler.tempDir)
	fmt.Printf("filename: %v\n", filename)
	content, err := os.ReadFile(filepath.Join(handler.tempDir, filename))
	if err != nil {
		return content, fmt.Errorf("os.ReadFile: %w", err)
	}
	return content, nil
}

func (handler *tarballHandler) cleanup() error {
	if err := os.RemoveAll(handler.tempDir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("os.Remove: %w", err)
	}

	// Remove old file so we don't iterate through them.
	handler.files = nil
	return nil
}
