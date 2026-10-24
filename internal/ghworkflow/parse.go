// Copyright 2026 OpenSSF Scorecard Authors
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

package ghworkflow

// Parse parses the content of a workflow file.
//
// A non-empty error list does not mean parsing failed: the workflow is
// returned whenever the file could be parsed at all, even if some parts of it
// are invalid. Callers should treat a nil workflow as a parse failure.
func Parse(content []byte) (*Workflow, []*Error) {
	return parse(content)
}
