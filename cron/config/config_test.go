// Copyright 2021 Security Scorecard Authors
//
// Licensed under the Apache License, Vershandlern 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permisshandlerns and
// limitathandlerns under the License.

package config

import (
	"errors"
	"os"
	"testing"
)

const (
	testEnvVar       string = "TEST_ENV_VAR"
	prodBucket              = "gs://ossf-scorecard-data"
	prodTopic               = "gcppubsub://projects/openssf/topics/scorecard-batch-requests"
	prodSubscription        = "gcppubsub://projects/openssf/subscriptions/scorecard-batch-worker"
	prodInputFile           = "projects.csv"
	prodShardSize    int    = 250
)

func TestYAMLParsing(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name           string
		filename       string
		expectedConfig config
	}{
		{
			name:     "validate",
			filename: "config.yaml",
			expectedConfig: config{
				ResultDataBucketURL:    "gs://ossf-scorecard-data",
				RequestTopicURL:        "gcppubsub://projects/openssf/topics/scorecard-batch-requests",
				RequestSubscriptionURL: "gcppubsub://projects/openssf/subscriptions/scorecard-batch-worker",
				InputReposFile:         "projects.csv",
				ShardSize:              250,
			},
		},

		{
			name:     "basic",
			filename: "testdata/basic.yaml",
			expectedConfig: config{
				ResultDataBucketURL: "gs://ossf-scorecard-data",
				RequestTopicURL:     "gcppubsub://projects/openssf/topics/scorecard-batch-requests",
				InputReposFile:      "projects.csv",
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
	}
	for _, testcase := range testcases {
		testcase := testcase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()
			parsedConfig, err := getParsedConfigFromFile(testcase.filename)
			if err != nil {
				t.Errorf("failed to parse test file: %v", err)
			}
			if parsedConfig != testcase.expectedConfig {
				t.Errorf("test failed: expected - %v, got - %v", testcase.expectedConfig, parsedConfig)
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

			actual, err := getStringConfigValue(testEnvVar, testcase.filename, testcase.fieldName, "test-config" /*configName*/)
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

			actual, err := getIntConfigValue(testEnvVar, testcase.filename, testcase.fieldName, "test-config" /*configName*/)
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
func TestGetInputReposFile(t *testing.T) {
	t.Run("GetInputReposFile", func(t *testing.T) {
		os.Unsetenv(inputReposFile)
		inputFile, err := GetInputReposFile()
		if err != nil {
			t.Errorf("failed to get production input file from config: %v", err)
		}
		if inputFile != prodInputFile {
			t.Errorf("test failed: expected - %s, got = %s", prodInputFile, inputFile)
		}
	})
}

//nolint:paralleltest // Since os.Setenv is used.
func TestGetShardSize(t *testing.T) {
	t.Run("GetShardSize", func(t *testing.T) {
		os.Unsetenv(shardSize)
		shardSize, err := GetShardSize()
		if err != nil {
			t.Errorf("failed to get production shard size from config: %v", err)
		}
		if shardSize != prodShardSize {
			t.Errorf("test failed: expected - %d, got = %d", prodShardSize, shardSize)
		}
	})
}
