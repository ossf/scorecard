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

// Ecosystem is a package ecosystem supported by OSV, GitHub, etc.
type ecosystem string

// OSV ecosystem naming data source: https://ossf.github.io/osv-schema/#affectedpackage-field
// nolint
const (
	// The Go ecosystem.
	ecosystemGo ecosystem = "Go"

	// The NPM ecosystem.
	ecosystemNpm ecosystem = "npm"

	// The Android ecosystem
	ecosystemAndroid ecosystem = "Android" // nolint:unused

	// The crates.io ecosystem for RUST.
	ecosystemCrates ecosystem = "crates.io"

	// For reports from the OSS-Fuzz project that have no more appropriate ecosystem.
	ecosystemOssFuzz ecosystem = "OSS-Fuzz" // nolint:unused

	// The Python PyPI ecosystem. PyPI is the main package source of pip.
	ecosystemPyPI ecosystem = "PyPI"

	// The RubyGems ecosystem.
	ecosystemRubyGems ecosystem = "RubyGems"

	// The PHP package manager ecosystem. Packagist is the main Composer repository.
	ecosystemPackagist ecosystem = "Packagist"

	// The Maven Java package ecosystem.
	ecosystemMaven ecosystem = "Maven"

	// The NuGet package ecosystem.
	ecosystemNuGet ecosystem = "Nuget"

	// The Linux kernel.
	ecosystemLinux ecosystem = "Linux" // nolint:unused

	// The Debian package ecosystem.
	ecosystemDebian ecosystem = "Debian" // nolint:unused

	// Hex is the package manager of Erlang.
	// TODO: GitHub doesn't support hex as the ecosystem for Erlang yet. Add this to the map in the future.
	ecosystemHex ecosystem = "Hex" // nolint:unused
)

var (
	//gitHubToOSV defines the ecosystem naming mapping relationship between GitHub and others.
	gitHubToOSV = map[string]ecosystem{
		// GitHub ecosystem naming data source: https://docs.github.com/en/code-security/supply-chain-security/
		// understanding-your-software-supply-chain/about-the-dependency-graph#supported-package-ecosystems
		"gomod":    ecosystemGo, /* go.mod and go.sum */
		"cargo":    ecosystemCrates,
		"pip":      ecosystemPyPI, /* pip and poetry */
		"npm":      ecosystemNpm,  /* npm and yarn */
		"maven":    ecosystemMaven,
		"composer": ecosystemPackagist,
		"rubygems": ecosystemRubyGems,
		"nuget":    ecosystemNuGet,
	}
)

func toEcosystem(e string) (ecosystem, error) {
	if ecosystemOSV, found := gitHubToOSV[e]; found {
		return ecosystemOSV, nil
	}
	return "", fmt.Errorf("%w for github entry %s", errMappingNotFound, e)
}
