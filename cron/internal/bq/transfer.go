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

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"cloud.google.com/go/bigquery"
)

const partitionDateFormat = "20060102"

func createGCSRef(bucketURL, fileURI string) *bigquery.GCSReference {
	gcsRef := bigquery.NewGCSReference(fmt.Sprintf("%s/%s", bucketURL, fileURI))
	gcsRef.SourceFormat = bigquery.JSON
	return gcsRef
}

func createBQLoader(ctx context.Context, projectID, datasetName, tableName string,
	partitionDate time.Time, gcsRef *bigquery.GCSReference,
) (*bigquery.Client, *bigquery.Loader, error) {
	bqClient, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create bigquery client: %w", err)
	}

	partitionedTable := fmt.Sprintf("%s$%s", tableName, partitionDate.Format(partitionDateFormat))
	loader := bqClient.Dataset(datasetName).Table(partitionedTable).LoaderFrom(gcsRef)
	loader.WriteDisposition = bigquery.WriteTruncate
	return bqClient, loader, nil
}

func startDataTransferJob(ctx context.Context,
	bucketURL, fileURI, projectID, datasetName, tableName string,
	partitionDate time.Time,
) error {
	gcsRef := createGCSRef(bucketURL, fileURI)
	bqClient, loader, err := createBQLoader(ctx, projectID, datasetName, tableName, partitionDate, gcsRef)
	if err != nil {
		return fmt.Errorf("error creating BQ loader: %w", err)
	}
	defer bqClient.Close()

	job, err := loader.Run(ctx)
	if err != nil {
		return fmt.Errorf("failed to create load job: %w", err)
	}

	log.Printf("Job created: %s", job.ID())
	status, err := job.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error during job.Wait: %w", err)
	}
	if status.Err() != nil {
		return fmt.Errorf("job returned error status: %w", status.Err())
	}
	return nil
}
