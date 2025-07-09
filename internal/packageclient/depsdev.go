// Copyright 2024 OpenSSF Scorecard Authors
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

package packageclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// This interface lets Scorecard look up package manager metadata for a project.
type ProjectPackageClient interface {
	GetProjectPackageVersions(ctx context.Context, host, project string) (*ProjectPackageVersions, error)
	GetPackage(ctx context.Context, systemName, packageName string) (*Package, error)
}

type depsDevClient struct {
	client *http.Client
}

type ProjectPackageVersions struct {
	Versions []Version `json:"versions"`
}

type Package struct {
	PackageKey PackageKey `json:"packageKey"`
	Purl       string     `json:"purl"`
	Versions   []Version  `json:"versions"`
}

type PackageKey struct {
	PackageSystem string `json:"system"`
	PackageName   string `json:"name"`
}

type VersionKey struct {
	System          string           `json:"system"`
	Name            string           `json:"name"`
	Version         string           `json:"version"`
	SLSAProvenances []SLSAProvenance `json:"slsaProvenances"`
}

type SLSAProvenance struct {
	SourceRepository string `json:"sourceRepository"`
	Commit           string `json:"commit"`
	Verified         bool   `json:"verified"`
}

type Version struct {
	RelationType       string           `json:"relationType"`
	RelationProvenance string           `json:"relationProvenance"`
	Purl               string           `json:"purl"`
	PublishedAt        string           `json:"publishedAt"`
	VersionKey         VersionKey       `json:"versionKey"`
	SLSAProvenances    []SLSAProvenance `json:"slsaProvenances"`
	IsDefault          bool             `json:"isDefault"`
	IsDeprecated       bool             `json:"isDeprecated"`
}

func CreateDepsDevClient() ProjectPackageClient {
	return depsDevClient{
		client: &http.Client{},
	}
}

var (
	ErrDepsDevAPI            = errors.New("deps.dev")
	ErrProjNotFoundInDepsDev = errors.New("project not found in deps.dev")
)

func (d depsDevClient) GetProjectPackageVersions(
	ctx context.Context, host, project string,
) (*ProjectPackageVersions, error) {
	path := fmt.Sprintf("%s/%s", host, project)
	query := fmt.Sprintf("https://api.deps.dev/v3/projects/%s:packageversions", url.QueryEscape(path))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, query, nil)
	if err != nil {
		return nil, fmt.Errorf("http.NewRequestWithContext: %w", err)
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("deps.dev GetProjectPackageVersions: %w", err)
	}
	defer resp.Body.Close()

	var res ProjectPackageVersions
	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrProjNotFoundInDepsDev
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %s", ErrDepsDevAPI, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("resp.Body.Read: %w", err)
	}

	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, fmt.Errorf("deps.dev json.Unmarshal: %w", err)
	}

	return &res, nil
}

func (d depsDevClient) GetPackage(ctx context.Context, systemName, packageName string) (*Package, error) {
	query := fmt.Sprintf("https://api.deps.dev/v3alpha/systems/%s/packages/%s", systemName, packageName)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, query, nil)
	if err != nil {
		return nil, fmt.Errorf("http.NewRequestWithContext: %w", err)
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("deps.dev GetPackage: %w", err)
	}
	defer resp.Body.Close()

	var res Package
	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrProjNotFoundInDepsDev
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %s", ErrDepsDevAPI, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("resp.Body.Read: %w", err)
	}

	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, fmt.Errorf("deps.dev json.Unmarshal: %w", err)
	}

	return &res, nil
}
