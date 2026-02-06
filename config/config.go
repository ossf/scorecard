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

package config

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"go.yaml.in/yaml/v3"

	sce "github.com/ossf/scorecard/v5/errors"
	"github.com/ossf/scorecard/v5/internal/checknames"
)

var (
	errInvalidCheck  = errors.New("check is not valid")
	errInvalidReason = errors.New("reason is not valid")
)

// Config contains configurations defined by maintainers.
type Config struct {
	Annotations []Annotation `yaml:"annotations"`
	GPGKeys     GPGKeyConfig `yaml:"gpg-keys"`
}

// GPGKeyConfig contains GPG key URLs for signature verification.
type GPGKeyConfig struct {
	// URLs are global fallback keys used for all releases
	URLs []string `yaml:"urls"`
	// Releases maps release tag patterns to specific key URLs
	Releases []ReleaseKeyConfig `yaml:"releases"`
}

// ReleaseKeyConfig specifies keys for specific release tags.
type ReleaseKeyConfig struct {
	// Tag is a glob pattern matching release tags (e.g., "v1.*", "v2.0.0")
	Tag string `yaml:"tag"`
	// URLs are the GPG key URLs for releases matching this tag pattern
	URLs []string `yaml:"urls"`
}

// GetGPGKeys returns the GPG key configuration.
func (c Config) GetGPGKeys() interface{} {
	return c.GPGKeys
}

// GetURLs returns the list of global GPG key URLs.
func (g GPGKeyConfig) GetURLs() []string {
	return g.URLs
}

// GetReleases returns the per-release key configurations.
func (g GPGKeyConfig) GetReleases() interface{} {
	// Convert to []interface{} for type assertion compatibility
	result := make([]interface{}, len(g.Releases))
	for i := range g.Releases {
		result[i] = g.Releases[i]
	}
	return result
}

// GetTag returns the tag pattern for this release key config.
func (r ReleaseKeyConfig) GetTag() string {
	return r.Tag
}

// GetURLs returns the GPG key URLs for this release.
func (r ReleaseKeyConfig) GetURLs() []string {
	return r.URLs
}

// parseFile takes the scorecard.yml file content and returns a `Config`.
func parseFile(c *Config, content []byte) error {
	unmarshalErr := yaml.Unmarshal(content, c)
	if unmarshalErr != nil {
		return sce.WithMessage(sce.ErrScorecardInternal, unmarshalErr.Error())
	}

	return nil
}

func isValidCheck(check string) bool {
	for _, c := range checknames.AllValidChecks {
		if strings.EqualFold(c, check) {
			return true
		}
	}
	return false
}

func validate(c Config) error {
	for _, annotation := range c.Annotations {
		for _, check := range annotation.Checks {
			if !isValidCheck(check) {
				return fmt.Errorf("%w: %s", errInvalidCheck, check)
			}
		}
		for _, reasonGroup := range annotation.Reasons {
			if !isValidReason(reasonGroup.Reason) {
				return fmt.Errorf("%w: %s", errInvalidReason, reasonGroup.Reason)
			}
		}
	}
	return nil
}

// Parse reads the configuration file from the repo, stored in scorecard.yml, and returns a `Config`.
func Parse(r io.Reader) (Config, error) {
	c := Config{}
	// Find scorecard.yml file in the repository's root
	content, err := io.ReadAll(r)
	if err != nil {
		return Config{}, fmt.Errorf("fail to read configuration file: %w", err)
	}

	err = parseFile(&c, content)
	if err != nil {
		return Config{}, fmt.Errorf("fail to parse configuration file: %w", err)
	}

	err = validate(c)
	if err != nil {
		return Config{}, fmt.Errorf("configuration file is not valid: %w", err)
	}

	// Return configuration
	return c, nil
}
