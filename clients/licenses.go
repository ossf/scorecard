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

// License represents a customized struct for licenses used by clients.
// from pkg.go.dev/github.com/google/go-github/github#RepositoryLicense.
type License struct {
	Key    string // RepositoryLicense.GetLicense().GetKey()
	Name   string // RepositoryLicense.GetLicense().GetName()
	Path   string // RepositoryLicense.GetName()
	SPDXId string // RepositoryLicense.GetLicense().GetSPDXID()
	Type   string // RepositoryLicense.GetType()
	Size   int    // RepositoryLicense.GetSize()
}
