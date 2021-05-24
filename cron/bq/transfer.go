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

package bq

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/bigquery"

	"github.com/ossf/scorecard/cron/config"
)

func StartDataTransferJob(ctx context.Context, bucketURL, filename string) error {
	projectID, err := config.GetProjectID()
	if err != nil {
		return fmt.Errorf("error getting ProjectId: %w", err)
	}
	bq, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to create bigquery client: %w", err)
	}

	datasetName, err := config.GetBigQueryDataset()
	if err != nil {
		return fmt.Errorf("error getting BigQuery dataset: %w", err)
	}
	tableName, err := config.GetBigQueryTable()
	if err != nil {
		return fmt.Errorf("error getting BigQuery table: %w", err)
	}
	gcsRef := bigquery.NewGCSReference(fmt.Sprintf("%s/%s", bucketURL, filename))
	gcsRef.AutoDetect = true
	gcsRef.SourceFormat = bigquery.JSON
	dataset := bq.Dataset(datasetName)
	loader := dataset.Table(tableName).LoaderFrom(gcsRef)
	loader.WriteDisposition = bigquery.WriteTruncate

	job, err := loader.Run(ctx)
	if err != nil {
		return fmt.Errorf("failed to create load job: %w", err)
	}
	log.Printf("Job created: %s", job.ID())
	return nil
}
