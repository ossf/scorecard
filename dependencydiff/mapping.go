// Copyright 2022 Security Scorecard Authors
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

package dependencydiff

import (
	"fmt"
)

// EcosystemOSV is a package ecosystem supported by OSV.
type ecosystemOSV string

// Data source: https://ossf.github.io/osv-schema/#affectedpackage-field
const (
	// The Go ecosystem.
	goOSV ecosystemOSV = "Go"

	// The NPM ecosystem.
	npmOSV = "npm"

	// The Android ecosystem
	androidOSV = "Android" // nolint:unused

	// The crates.io ecosystem for RUST.
	cratesOSV = "crates.io"

	// For reports from the OSS-Fuzz project that have no more appropriate ecosystem.
	ossFuzzOSV = "OSS-Fuzz" // nolint:unused

	// The Python PyPI ecosystem. PyPI is the main package source of pip.
	pyPIOSV = "PyPI"

	// The RubyGems ecosystem.
	rubyGemsOSV = "RubyGems"

	// The PHP package manager ecosystem. Packagist is the main Composer repository.
	packagistOSV = "Packagist"

	// The Maven Java package ecosystem.
	mavenOSV = "Maven"

	// The NuGet package ecosystem.
	nuGetOSV = "Nuget"

	// The Linux kernel.
	linuxOSV = "Linux" // nolint:unused

	// The Debian package ecosystem.
	debianOSV = "Debian" // nolint:unused

	// Hex is the package manager of Erlang.
	// TODO: GitHub doesn't support hex as the ecosystem for Erlang yet. Add this mapping in the future.
	hexOSV = "Hex" // nolint:unused
)

// ecosystemGitHub is a package ecosystem supported by GitHub.
type ecosystemGitHub string

// Data source: https://docs.github.com/en/code-security/supply-chain-security/understanding-
// your-software-supply-chain/about-the-dependency-graph#supported-package-ecosystems
// nolint
const (
	// The Go ecosystem on GitHub includes go.mod and go.sum, but both use gomod as the ecosystem name.
	goGitHub ecosystemGitHub = "gomod"

	// Npm is the package manager of JavaScript.
	// Yarn is another package manager of JavaScript, GitHub also uses "npm" as its ecosys name.
	npmGitHub ecosystemGitHub = "npm"

	// RubyGems is the package manager of Ruby.
	rubyGemsGitHub ecosystemGitHub = "rubygems"

	// Pip is the package manager of Python.
	// Poetry is another package manager of Python, GitHub also uses "pip" as its ecosys name.
	pipGitHub ecosystemGitHub = "pip"

	// Action is the GitHub Action.
	actionGitHub ecosystemGitHub = "actions" // nolint:unused

	// Cargo is the package manager of RUST.
	cargoGitHub ecosystemGitHub = "cargo"

	// Composer is the package manager of PHP, there is currently no mapping to the OSV.
	composerGitHub ecosystemGitHub = "composer"

	// NuGet is the package manager of .NET languages (C#, F#, VB), C++.
	nuGetGitHub ecosystemGitHub = "nuget"

	// Maven is the package manager of 	Java and Scala.
	mavenGitHub ecosystemGitHub = "maven"
)

func (e ecosystemGitHub) isValid() bool {
	switch e {
	case goGitHub, npmGitHub, rubyGemsGitHub, pipGitHub, actionGitHub,
		cargoGitHub, composerGitHub, nuGetGitHub, mavenGitHub:
		return true
	default:
		return false
	}
}

var (
	gitHubToOSV = map[ecosystemGitHub]ecosystemOSV{
		goGitHub:       goOSV, /* go.mod and go.sum */
		cargoGitHub:    cratesOSV,
		pipGitHub:      pyPIOSV, /* pip and poetry */
		npmGitHub:      npmOSV,  /* npm and yarn */
		mavenGitHub:    mavenOSV,
		composerGitHub: packagistOSV,
		rubyGemsGitHub: rubyGemsOSV,
		nuGetGitHub:    nuGetOSV,
	}
)

func (e ecosystemGitHub) toOSV() (ecosystemOSV, error) {
	if ecosystemOSV, found := gitHubToOSV[e]; found {
		return ecosystemOSV, nil
	}
	return "", fmt.Errorf("%w for github entry %s", errMappingNotFound, e)
}
