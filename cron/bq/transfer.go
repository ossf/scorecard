// Copyright 2021 Security Scorecard Authors
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

	"github.com/ossf/scorecard/cron/config"
)

const partitionDateFormat = "20060102"

func getBQConfig() (projectID, datasetName, tableName string, err error) {
	projectID, err = config.GetProjectID()
	if err != nil {
		return projectID, datasetName, tableName, fmt.Errorf("error getting ProjectId: %w", err)
	}
	datasetName, err = config.GetBigQueryDataset()
	if err != nil {
		return projectID, datasetName, tableName, fmt.Errorf("error getting BigQuery dataset: %w", err)
	}
	tableName, err = config.GetBigQueryTable()
	if err != nil {
		return projectID, datasetName, tableName, fmt.Errorf("error getting BigQuery table: %w", err)
	}
	return
}

func createGCSRef(bucketURL, fileURI string) *bigquery.GCSReference {
	gcsRef := bigquery.NewGCSReference(fmt.Sprintf("%s/%s", bucketURL, fileURI))
	gcsRef.SourceFormat = bigquery.JSON
	return gcsRef
}

func createBQLoader(ctx context.Context, projectID, datasetName, tableName string,
	partitionDate time.Time, gcsRef *bigquery.GCSReference) (*bigquery.Client, *bigquery.Loader, error) {
	bqClient, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create bigquery client: %w", err)
	}

	partitionedTable := fmt.Sprintf("%s$%s", tableName, partitionDate.Format(partitionDateFormat))
	loader := bqClient.Dataset(datasetName).Table(partitionedTable).LoaderFrom(gcsRef)
	loader.WriteDisposition = bigquery.WriteTruncate
	return bqClient, loader, nil
}

func StartDataTransferJob(ctx context.Context, bucketURL, fileURI string, partitionDate time.Time) error {
	projectID, datasetName, tableName, err := getBQConfig()
	if err != nil {
		return fmt.Errorf("error getting BQ config: %w", err)
	}
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
