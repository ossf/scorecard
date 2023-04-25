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

// SearchRequest queries a repo for `Query`.
// If `Filename` is provided, only matching filenames are queried.
// If `Path` is provided, only files with matching paths are queried.
type SearchRequest struct {
	Query    string
	Filename string
	Path     string
}

// SearchResponse returns the results from a search request on a repo.
type SearchResponse struct {
	Results []SearchResult
	Hits    int
}

// SearchResult represents a matching result from the search query.
type SearchResult struct {
	Path string
}

// SearchCommitsOptions represents the parameters in the search commit query.
type SearchCommitsOptions struct {
	Author string
}
