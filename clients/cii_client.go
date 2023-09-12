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

import (
	"context"
)

const (
	// Unknown or non-parsable CII Best Practices badge.
	Unknown BadgeLevel = iota
	// NotFound represents when CII Best Practices returns an empty response for a project.
	NotFound
	// InProgress state of CII Best Practices badge.
	InProgress
	// Passing level for CII Best Practices badge.
	Passing
	// Silver level for CII Best Practices badge.
	Silver
	// Gold level for CII Best Practices badge.
	Gold
)

// BadgeLevel corresponds to CII-Best-Practices badge levels.
// https://www.bestpractices.dev/en
type BadgeLevel uint

// String returns a string value for BadgeLevel enum.
func (badge BadgeLevel) String() string {
	switch badge {
	case Unknown:
		return "Unknown"
	case NotFound:
		return "not_found"
	case InProgress:
		return "in_progress"
	case Passing:
		return "passing"
	case Silver:
		return "silver"
	case Gold:
		return "gold"
	default:
		return ""
	}
}

// CIIBestPracticesClient interface returns the BadgeLevel for a repo URL.
type CIIBestPracticesClient interface {
	GetBadgeLevel(ctx context.Context, uri string) (BadgeLevel, error)
}

// DefaultCIIBestPracticesClient returns http-based implementation of the interface.
func DefaultCIIBestPracticesClient() CIIBestPracticesClient {
	return &httpClientCIIBestPractices{}
}

// BlobCIIBestPracticesClient returns a blob-based implementation of the interface.
func BlobCIIBestPracticesClient(bucketURL string) CIIBestPracticesClient {
	return &blobClientCIIBestPractices{
		bucketURL: bucketURL,
	}
}
