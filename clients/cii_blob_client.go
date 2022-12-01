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
	"fmt"

	"gocloud.dev/blob"
	// Needed to link GCP drivers.
	_ "gocloud.dev/blob/gcsblob"
)

// blobClientCIIBestPractices implements the CIIBestPracticesClient interface.
// A gocloud blob client is used to communicate with the CII Best Practices data.
type blobClientCIIBestPractices struct {
	bucketURL string
}

// GetBadgeLevel implements CIIBestPracticesClient.GetBadgeLevel.
func (client *blobClientCIIBestPractices) GetBadgeLevel(ctx context.Context, uri string) (BadgeLevel, error) {
	bucket, err := blob.OpenBucket(ctx, client.bucketURL)
	if err != nil {
		return Unknown, fmt.Errorf("error during blob.OpenBucket: %w", err)
	}
	defer bucket.Close()

	objectName := fmt.Sprintf("%s/result.json", uri)

	exists, err := bucket.Exists(ctx, objectName)
	if err != nil {
		return Unknown, fmt.Errorf("error during bucket.Exists: %w", err)
	}
	if !exists {
		return NotFound, nil
	}

	jsonData, err := bucket.ReadAll(ctx, objectName)
	if err != nil {
		return Unknown, fmt.Errorf("error during bucket.ReadAll: %w", err)
	}

	parsedResponse, err := ParseBadgeResponseFromJSON(jsonData)
	if err != nil {
		return Unknown, fmt.Errorf("error parsing data: %w", err)
	}
	if len(parsedResponse) < 1 {
		return NotFound, nil
	}
	return parsedResponse[0].getBadgeLevel()
}
