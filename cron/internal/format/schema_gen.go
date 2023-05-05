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

package format

import (
	"encoding/json"
	"fmt"

	"cloud.google.com/go/bigquery"
	"github.com/mcuadros/go-jsonschema-generator"
)

// https://github.com/googleapis/google-cloud-go/blob/bigquery/v1.30.0/bigquery/schema.go#L544
type bigQueryJSONField struct {
	Description string              `json:"description"`
	Mode        string              `json:"mode"`
	Name        string              `json:"name"`
	Type        string              `json:"type"`
	Fields      []bigQueryJSONField `json:"fields,omitempty"`
}

func generateSchema(schema bigquery.Schema) []bigQueryJSONField {
	var bqs []bigQueryJSONField
	for _, fs := range schema {
		bq := bigQueryJSONField{
			Description: fs.Description,
			Name:        fs.Name,
			Type:        string(fs.Type),
			Fields:      generateSchema(fs.Schema),
		}
		// https://github.com/googleapis/google-cloud-go/blob/bigquery/v1.30.0/bigquery/schema.go#L125

		switch {
		// Make all fields optional to give us flexibility:
		// discard `fs.Required`.
		// An alternative would be to let the caller
		// use https://pkg.go.dev/cloud.google.com/go/bigquery#Schema.Relax.
		case fs.Repeated:
			bq.Mode = "REPEATED"
		default:
			bq.Mode = "NULLABLE"
		}

		bqs = append(bqs, bq)
	}

	return bqs
}

// GenerateBQSchema generates the BQ schema in JSON format.
// Can be used to generate a BQ table:
// `bq mk --table    --time_partitioning_type DAY \
// --require_partition_filter=TRUE \
// --time_partitioning_field date \
// openssf:scorecardcron.scorecard-rawdata-releasetest \
// cron/format/bq.raw.schema`.
// The structure `t` must be annotated using BQ fields:
// a string `bigquery:"name"`.
func GenerateBQSchema(t interface{}) (string, error) {
	schema, err := bigquery.InferSchema(t)
	if err != nil {
		return "", fmt.Errorf("bigquery.InferSchema: %w", err)
	}
	jsonFields := generateSchema(schema)

	jsonData, err := json.Marshal(jsonFields)
	if err != nil {
		return "", fmt.Errorf("json.Marshal: %w", err)
	}
	return string(jsonData), nil
}

// GenerateJSONSchema generates the schema for a JSON structure.
func GenerateJSONSchema(t interface{}) string {
	s := &jsonschema.Document{}
	s.Read(t)

	return s.String()
}
