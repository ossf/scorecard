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

package clients

// Repo interface uniquely identifies a repo.
type Repo interface {
	// Path returns the specifier of the repository within its forge
	Path() string
	// URI returns the fully qualified address of the repository
	URI() string
	// Host returns the web domain of the repository
	Host() string
	// String returns a string representation of the repository URI
	String() string
	// IsValid returns whether the repository provided is a real URI
	IsValid() error
	Metadata() []string
	AppendMetadata(metadata ...string)
}
