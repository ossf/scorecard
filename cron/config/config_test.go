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

package config

import (
	"errors"
	"log"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

const (
	testEnvVar                  string = "TEST_ENV_VAR"
	prodProjectID                      = "openssf"
	prodBucket                         = "gs://ossf-scorecard-data2"
	prodTopic                          = "gcppubsub://projects/openssf/topics/scorecard-batch-requests"
	prodSubscription                   = "gcppubsub://projects/openssf/subscriptions/scorecard-batch-worker"
	prodBigQueryDataset                = "scorecardcron"
	prodBigQueryTable                  = "scorecard-v2"
	prodCompletionThreshold            = 0.99
	prodWebhookURL                     = ""
	prodCIIDataBucket                  = "gs://ossf-scorecard-cii-data"
	prodBlacklistedChecks              = "CI-Tests,Contributors,Dependency-Update-Tool,Fuzzing,SAST,Vulnerabilities"
	prodShardSize               int    = 10
	prodMetricExporter          string = "stackdriver"
	prodMetricStackdriverPrefix string = "scorecard-cron"
	// Raw results.
	prodRawBucket             = "gs://ossf-scorecard-rawdata"
	prodRawBigQueryTable      = "scorecard-rawdata"
	prodAPIBucketURL          = "gs://ossf-scorecard-cron-results"
	prodInputBucketURL        = "gs://ossf-scorecard-input-projects"
	prodInputBucketPrefix     = ""
	prodInputBucketPrefixFile = ""
)

var (
	prodInputBucketParams = map[string]string{
		"url":         prodInputBucketURL,
		"prefix":      prodInputBucketPrefix,
		"prefix-file": prodInputBucketPrefixFile,
	}
	prodScorecardParams = map[string]string{
		"api-results-bucket-url":     prodAPIBucketURL,
		"blacklisted-checks":         prodBlacklistedChecks,
		"cii-data-bucket-url":        prodCIIDataBucket,
		"raw-bigquery-table":         prodRawBigQueryTable,
		"raw-result-data-bucket-url": prodRawBucket,
	}
	prodAdditionalParams = map[string]map[string]string{
		"input-bucket": prodInputBucketParams,
		"scorecard":    prodScorecardParams,
	}
)

func getByteValueFromFile(filename string) ([]byte, error) {
	if filename == "" {
		return nil, nil
	}
	//nolint
	return os.ReadFile(filename)
}

// runs once before all tests, to initialize the config file for testing purposes.
func TestMain(m *testing.M) {
	// TODO change to config.yaml when removing built-in embedding
	*configFilename = ""
	if err := ReadConfig(); err != nil {
		log.Fatalf("failed to read config: %v", err)
	}
	os.Exit(m.Run())
}

func TestYAMLParsing(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		expectedConfig config
		name           string
		filename       string
	}{
		{
			name:     "validate",
			filename: "config.yaml",
			expectedConfig: config{
				ProjectID:               prodProjectID,
				ResultDataBucketURL:     prodBucket,
				RequestTopicURL:         prodTopic,
				RequestSubscriptionURL:  prodSubscription,
				BigQueryDataset:         prodBigQueryDataset,
				BigQueryTable:           prodBigQueryTable,
				CompletionThreshold:     prodCompletionThreshold,
				WebhookURL:              prodWebhookURL,
				ShardSize:               prodShardSize,
				MetricExporter:          prodMetricExporter,
				MetricStackdriverPrefix: prodMetricStackdriverPrefix,
				InputBucketURL:          prodInputBucketURL,
				InputBucketPrefix:       prodInputBucketPrefix,
				AdditionalParams:        prodAdditionalParams,
			},
		},
		{
			name:     "basic",
			filename: "testdata/basic.yaml",
			expectedConfig: config{
				ResultDataBucketURL: "gs://ossf-scorecard-data",
				RequestTopicURL:     "gcppubsub://projects/openssf/topics/scorecard-batch-requests",
				ShardSize:           250,
			},
		},
		{
			name:     "missingField",
			filename: "testdata/missing_field.yaml",
			expectedConfig: config{
				ResultDataBucketURL: "gs://ossf-scorecard-data",
				RequestTopicURL:     "gcppubsub://projects/openssf/topics/scorecard-batch-requests",
				ShardSize:           250,
			},
		},
		{
			name:     "optional map values",
			filename: "testdata/optional_maps.yaml",
			expectedConfig: config{
				AdditionalParams: map[string]map[string]string{
					"criticality": {
						"empty-test": "",
					},
					"scorecard": {
						"cii-data-bucket-url":    "gs://ossf-scorecard-cii-data",
						"result-data-bucket-url": "gs://ossf-scorecard-data2",
					},
				},
			},
		},
	}
	for _, testcase := range testcases {
		testcase := testcase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()
			byteValue, err := getByteValueFromFile(testcase.filename)
			if err != nil {
				t.Errorf("test failed to parse input file: %v", err)
			}
			parsedConfig, err := getParsedConfigFromFile(byteValue)
			if err != nil {
				t.Errorf("failed to parse test file: %v", err)
			}
			if !cmp.Equal(parsedConfig, testcase.expectedConfig) {
				diff := cmp.Diff(parsedConfig, testcase.expectedConfig)
				t.Errorf("test failed: expected - %v, got - %v. \n%s", testcase.expectedConfig, parsedConfig, diff)
			}
		})
	}
}

//nolint:paralleltest // Since os.Setenv is used.
func TestGetStringConfigValue(t *testing.T) {
	testcases := []struct {
		expectedErr error
		name        string
		envVal      string
		filename    string
		fieldName   string
		expected    string
		hasError    bool
		setEnv      bool
	}{
		{
			name:     "Basic",
			envVal:   "test",
			setEnv:   true,
			expected: "test",
		},
		{
			name:      "GetParsedValue",
			setEnv:    false,
			filename:  "testdata/basic.yaml",
			fieldName: "ResultDataBucketURL",
			expected:  "gs://ossf-scorecard-data",
		},
		{
			name:        "EmptyValue",
			envVal:      "",
			setEnv:      true,
			hasError:    true,
			expectedErr: ErrorEmptyConfigValue,
		},
	}
	for _, testcase := range testcases {
		testcase := testcase
		t.Run(testcase.name, func(t *testing.T) {
			os.Unsetenv(testEnvVar)
			if testcase.setEnv {
				os.Setenv(testEnvVar, testcase.envVal)
			}

			byteValue, err := getByteValueFromFile(testcase.filename)
			if err != nil {
				t.Errorf("test failed during input parsing: %v", err)
			}
			actual, err := getStringConfigValue(testEnvVar, byteValue, testcase.fieldName, "test-config" /*configName*/)
			if testcase.hasError {
				if err == nil || !errors.Is(err, testcase.expectedErr) {
					t.Errorf("test failed: expectedErr - %v, got - %v", testcase.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("test failed: expected - no error, got - %v", err)
				}
				if actual != testcase.expected {
					t.Errorf("testcase failed: expected - %s, got - %s", testcase.expected, actual)
				}
			}
		})
	}
}

//nolint:paralleltest // Since os.Setenv is used.
func TestGetIntConfigValue(t *testing.T) {
	testcases := []struct {
		expectedErr error
		name        string
		envVal      string
		filename    string
		fieldName   string
		expected    int
		hasError    bool
		setEnv      bool
	}{
		{
			name:     "Basic",
			envVal:   "11",
			setEnv:   true,
			expected: 11,
		},
		{
			name:      "GetParsedValue",
			setEnv:    false,
			filename:  "testdata/basic.yaml",
			fieldName: "ShardSize",
			expected:  250,
		},
	}
	for _, testcase := range testcases {
		testcase := testcase
		t.Run(testcase.name, func(t *testing.T) {
			os.Unsetenv(testEnvVar)
			if testcase.setEnv {
				os.Setenv(testEnvVar, testcase.envVal)
			}

			byteValue, err := getByteValueFromFile(testcase.filename)
			if err != nil {
				t.Errorf("test failed during input parsing: %v", err)
			}
			actual, err := getIntConfigValue(testEnvVar, byteValue, testcase.fieldName, "test-config" /*configName*/)
			if testcase.hasError {
				if err == nil || !errors.Is(err, testcase.expectedErr) {
					t.Errorf("test failed: expectedErr - %v, got - %v", testcase.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("test failed: expected - no error, got - %v", err)
				}
				if actual != testcase.expected {
					t.Errorf("testcase failed: expected - %d, got - %d", testcase.expected, actual)
				}
			}
		})
	}
}

//nolint:paralleltest // Since os.Setenv is used.
func TestGetProjectID(t *testing.T) {
	t.Run("GetProjectID", func(t *testing.T) {
		os.Unsetenv(projectID)
		project, err := GetProjectID()
		if err != nil {
			t.Errorf("failed to get production ProjectID from config: %v", err)
		}
		if project != prodProjectID {
			t.Errorf("test failed: expected - %s, got = %s", prodProjectID, project)
		}
	})
}

//nolint:paralleltest // Since os.Setenv is used.
func TestGetResultDataBucketURL(t *testing.T) {
	t.Run("GetResultDataBucketURL", func(t *testing.T) {
		os.Unsetenv(resultDataBucketURL)
		bucket, err := GetResultDataBucketURL()
		if err != nil {
			t.Errorf("failed to get production bucket URL from config: %v", err)
		}
		if bucket != prodBucket {
			t.Errorf("test failed: expected - %s, got = %s", prodBucket, bucket)
		}
	})
}

//nolint:paralleltest // Since os.Setenv is used.
func TestGetRequestTopicURL(t *testing.T) {
	t.Run("GetRequestTopicURL", func(t *testing.T) {
		os.Unsetenv(requestTopicURL)
		topic, err := GetRequestTopicURL()
		if err != nil {
			t.Errorf("failed to get production topic URL from config: %v", err)
		}
		if topic != prodTopic {
			t.Errorf("test failed: expected - %s, got = %s", prodTopic, topic)
		}
	})
}

//nolint:paralleltest // Since os.Setenv is used.
func TestGetRequestSubscriptionURL(t *testing.T) {
	t.Run("GetRequestSubscriptionURL", func(t *testing.T) {
		os.Unsetenv(requestSubscriptionURL)
		subscription, err := GetRequestSubscriptionURL()
		if err != nil {
			t.Errorf("failed to get production subscription URL from config: %v", err)
		}
		if subscription != prodSubscription {
			t.Errorf("test failed: expected - %s, got = %s", prodSubscription, subscription)
		}
	})
}

//nolint:paralleltest // Since os.Setenv is used.
func TestGetBigQueryDataset(t *testing.T) {
	t.Run("GetBigQueryDataset", func(t *testing.T) {
		os.Unsetenv(bigqueryDataset)
		dataset, err := GetBigQueryDataset()
		if err != nil {
			t.Errorf("failed to get production BQ datset from config: %v", err)
		}
		if dataset != prodBigQueryDataset {
			t.Errorf("test failed: expected - %s, got = %s", prodBigQueryDataset, dataset)
		}
	})
}

//nolint:paralleltest // Since os.Setenv is used.
func TestGetBigQueryTable(t *testing.T) {
	t.Run("GetBigQueryTable", func(t *testing.T) {
		os.Unsetenv(bigqueryTable)
		table, err := GetBigQueryTable()
		if err != nil {
			t.Errorf("failed to get production BQ table from config: %v", err)
		}
		if table != prodBigQueryTable {
			t.Errorf("test failed: expected - %s, got = %s", prodBigQueryTable, table)
		}
	})
}

//nolint:paralleltest // Since os.Setenv is used.
func TestGetShardSize(t *testing.T) {
	t.Run("GetShardSize", func(t *testing.T) {
		os.Unsetenv(shardSize)
		size, err := GetShardSize()
		if err != nil {
			t.Errorf("failed to get production shard size from config: %v", err)
		}
		if size != prodShardSize {
			t.Errorf("test failed: expected - %d, got = %d", prodShardSize, size)
		}
	})
}

//nolint:paralleltest // Since os.Setenv is used.
func TestGetMetricExporter(t *testing.T) {
	t.Run("GetMetricExporter", func(t *testing.T) {
		os.Unsetenv(metricExporter)
		exporter, err := GetMetricExporter()
		if err != nil {
			t.Errorf("failed to get production metric exporter from config: %v", err)
		}
		if exporter != prodMetricExporter {
			t.Errorf("test failed: expected - %s, got = %s", prodMetricExporter, exporter)
		}
	})
}

//nolint:paralleltest // Since os.Setenv is used.
func TestGetMetricStackdriverPrefix(t *testing.T) {
	t.Run("GetMetricStackdriverPrefix", func(t *testing.T) {
		os.Unsetenv(metricStackdriverPrefix)
		prefix, err := GetMetricStackdriverPrefix()
		if err != nil {
			t.Errorf("failed to get production metric stackdriver prefix from config: %v", err)
		}
		if prefix != prodMetricStackdriverPrefix {
			t.Errorf("test failed: expected - %s, got = %s", prodMetricStackdriverPrefix, prefix)
		}
	})
}

//nolint:paralleltest // Since os.Setenv is used.
func TestGetAPIResultsBucketURL(t *testing.T) {
	t.Run("GetBigQueryExportsBucketURL", func(t *testing.T) {
		bigqueryExportsBucketURL := apiResultsBucketURL
		os.Unsetenv(bigqueryExportsBucketURL)
		bucket, err := GetAPIResultsBucketURL()
		if err != nil {
			t.Errorf("failed to get production bucket URL from config: %v", err)
		}
		if bucket != prodAPIBucketURL {
			t.Errorf("test failed: expected - %s, got = %s", prodAPIBucketURL, bucket)
		}
	})
}

//nolint:paralleltest // Since os.Setenv is used.
func TestInputBucket(t *testing.T) {
	tests := []struct {
		f       func() (string, error)
		name    string
		envVar  string
		want    string
		wantErr bool
	}{
		{
			name:    "GetInputBucketURL",
			envVar:  inputBucketURL,
			want:    prodInputBucketURL,
			wantErr: false,
			f:       GetInputBucketURL,
		},
		{
			name:    "GetInputBucketPrefix",
			envVar:  inputBucketPrefix,
			want:    prodInputBucketPrefix,
			wantErr: false,
			f:       GetInputBucketPrefix,
		},
	}
	for _, testcase := range tests {
		testcase := testcase
		t.Run(testcase.name, func(t *testing.T) {
			os.Unsetenv(testcase.envVar)
			got, err := testcase.f()
			if (err != nil) != testcase.wantErr {
				t.Errorf("failed to get production value from config: %v", err)
			}
			if got != testcase.want {
				t.Errorf("test failed: expected - %s, got = %s", testcase.want, got)
			}
		})
	}
}

func TestEnvVarName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		mapName string
		subKey  string
		want    string
	}{
		{
			name:    "basic",
			mapName: "foo",
			subKey:  "bar",
			want:    "FOO_BAR",
		},
		{
			name:    "with dashes",
			mapName: "foo-bar",
			subKey:  "baz-qux",
			want:    "FOO_BAR_BAZ_QUX",
		},
	}
	for _, testcase := range tests {
		testcase := testcase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()
			got := envVarName(testcase.mapName, testcase.subKey)
			if got != testcase.want {
				t.Errorf("test failed: expected - %s, got = %s", testcase.want, got)
			}
		})
	}
}

func TestGetAdditionalParams(t *testing.T) {
	t.Parallel()
	//nolint:govet
	tests := []struct {
		name    string
		mapName string
		want    map[string]string
		wantErr bool
	}{
		{
			name:    "scorecard values",
			mapName: "scorecard",
			want:    prodScorecardParams,
			wantErr: false,
		},
		{
			name:    "nonexistant value",
			mapName: "this-value-should-never-exist",
			want:    map[string]string{},
			wantErr: true,
		},
	}
	for _, testcase := range tests {
		testcase := testcase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()
			got, err := GetAdditionalParams(testcase.mapName)
			if testcase.wantErr != (err != nil) {
				t.Fatalf("unexpected error value for GetAdditionalParams: %v", err)
			}
			if !cmp.Equal(got, testcase.want) {
				diff := cmp.Diff(got, testcase.want)
				t.Errorf("test failed: expected - %v, got - %v. \n%s", testcase.want, got, diff)
			}
		})
	}
}
