// Copyright 2024 OpenSSF Scorecard Authors
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

package azuredevopsrepo

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7"

	sce "github.com/ossf/scorecard/v5/errors"
)

const (
	repoDir      = "repo*"
	repoFilename = "azuredevopsrepo*.zip"
	maxSize      = 100 * 1024 * 1024 // 100MB limit
)

var (
	errUnexpectedStatusCode = errors.New("unexpected status code")
	errZipNotFound          = errors.New("zip not found")
	errInvalidFilePath      = errors.New("invalid zip file: contains file path outside of target directory")
	errFileTooLarge         = errors.New("file too large, possible zip bomb")
)

type zipHandler struct {
	client      *azuredevops.Client
	errSetup    error
	once        *sync.Once
	ctx         context.Context
	repourl     *Repo
	tempDir     string
	tempZipFile string
	files       []string
}

func (z *zipHandler) init(ctx context.Context, repourl *Repo) {
	z.errSetup = nil
	z.once = new(sync.Once)
	z.ctx = ctx
	z.repourl = repourl
}

func (z *zipHandler) setup() error {
	z.once.Do(func() {
		if err := z.cleanup(); err != nil {
			z.errSetup = sce.WithMessage(sce.ErrScorecardInternal, err.Error())
			return
		}

		if err := z.getZipfile(); err != nil {
			z.errSetup = sce.WithMessage(sce.ErrScorecardInternal, err.Error())
			return
		}

		if err := z.extractZip(); err != nil {
			z.errSetup = sce.WithMessage(sce.ErrScorecardInternal, err.Error())
			return
		}
	})

	return z.errSetup
}

func (z *zipHandler) getZipfile() error {
	tempDir, err := os.MkdirTemp("", repoDir)
	if err != nil {
		return fmt.Errorf("os.MkdirTemp: %w", err)
	}
	repoFile, err := os.CreateTemp(tempDir, repoFilename)
	if err != nil {
		return fmt.Errorf("%w io.Copy: %w", errZipNotFound, err)
	}
	defer repoFile.Close()

	// The zip download API is not exposed in the Azure DevOps Go SDK, so we need to construct the request manually.
	baseURL := fmt.Sprintf(
		"https://%s/%s/%s/_apis/git/repositories/%s/items",
		z.repourl.host,
		z.repourl.organization,
		z.repourl.project,
		z.repourl.id)

	queryParams := url.Values{}
	queryParams.Add("path", "/")
	queryParams.Add("download", "true")
	queryParams.Add("api-version", "7.1-preview.1")
	queryParams.Add("resolveLfs", "true")
	queryParams.Add("$format", "zip")

	if z.repourl.commitSHA == "HEAD" {
		queryParams.Add("versionDescriptor.versionType", "branch")
		queryParams.Add("versionDescriptor.version", z.repourl.defaultBranch)
	} else {
		queryParams.Add("versionDescriptor.versionType", "commit")
		queryParams.Add("versionDescriptor.version", z.repourl.commitSHA)
	}

	parsedURL, err := url.Parse(baseURL + "?" + queryParams.Encode())
	if err != nil {
		return fmt.Errorf("url.Parse: %w", err)
	}

	req := &http.Request{
		Method: http.MethodGet,
		URL:    parsedURL,
	}
	res, err := z.client.SendRequest(req)
	if err != nil {
		return fmt.Errorf("client.SendRequest: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: status code %d", errUnexpectedStatusCode, res.StatusCode)
	}

	if _, err := io.Copy(repoFile, res.Body); err != nil {
		return fmt.Errorf("io.Copy: %w", err)
	}

	z.tempDir = tempDir
	z.tempZipFile = repoFile.Name()

	return nil
}

func (z *zipHandler) getLocalPath() (string, error) {
	if err := z.setup(); err != nil {
		return "", fmt.Errorf("error during zipHandler.setup: %w", err)
	}
	absTempDir, err := filepath.Abs(z.tempDir)
	if err != nil {
		return "", fmt.Errorf("error during filepath.Abs: %w", err)
	}
	return absTempDir, nil
}

func (z *zipHandler) extractZip() error {
	zipReader, err := zip.OpenReader(z.tempZipFile)
	if err != nil {
		return fmt.Errorf("zip.OpenReader: %w", err)
	}
	defer zipReader.Close()

	destinationPrefix := filepath.Clean(z.tempDir) + string(os.PathSeparator)
	z.files = make([]string, 0, len(zipReader.File))
	for _, file := range zipReader.File {
		//nolint:gosec // G305: Handling of file paths is done below
		filenamepath := filepath.Join(z.tempDir, file.Name)
		if !strings.HasPrefix(filepath.Clean(filenamepath), destinationPrefix) {
			return errInvalidFilePath
		}

		if err := os.MkdirAll(filepath.Dir(filenamepath), 0o755); err != nil {
			return fmt.Errorf("error during os.MkdirAll: %w", err)
		}

		if file.FileInfo().IsDir() {
			continue
		}

		outFile, err := os.OpenFile(filenamepath, os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return fmt.Errorf("os.OpenFile: %w", err)
		}

		rc, err := file.Open()
		if err != nil {
			return fmt.Errorf("file.Open: %w", err)
		}

		written, err := io.CopyN(outFile, rc, maxSize)
		if err != nil && !errors.Is(err, io.EOF) {
			return fmt.Errorf("%w io.Copy: %w", errZipNotFound, err)
		}
		if written > maxSize {
			return errFileTooLarge
		}
		outFile.Close()

		filename := strings.TrimPrefix(filenamepath, destinationPrefix)
		z.files = append(z.files, filename)
	}
	return nil
}

func (z *zipHandler) listFiles(predicate func(string) (bool, error)) ([]string, error) {
	if err := z.setup(); err != nil {
		return nil, fmt.Errorf("error during zipHandler.setup: %w", err)
	}
	ret := make([]string, 0)
	for _, file := range z.files {
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

func (z *zipHandler) getFile(filename string) (*os.File, error) {
	if err := z.setup(); err != nil {
		return nil, fmt.Errorf("error during zipHandler.setup: %w", err)
	}
	f, err := os.Open(filepath.Join(z.tempDir, filename))
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	return f, nil
}

func (z *zipHandler) cleanup() error {
	if err := os.RemoveAll(z.tempDir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("os.Remove: %w", err)
	}

	z.files = nil
	return nil
}
