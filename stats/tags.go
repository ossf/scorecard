// Copyright 2020 OpenSSF Scorecard Authors
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

package stats

import "go.opentelemetry.io/otel/attribute"

const (
	// CheckName is the attribute key for the check name.
	CheckName = attribute.Key("checkName")
	// ErrorName is the attribute key for errors.
	ErrorName = attribute.Key("errorName")
	// RequestTag is the attribute key for the request type.
	RequestTag = attribute.Key("requestTag")
	// RepoHost is the attribute key for the host of the repository being scanned.
	RepoHost = attribute.Key("repoHost")
)
