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

// Package config defines the configuration values for the cron job.
package config

import (
	// Used to embed config.yaml.
	_ "embed"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

const (
	// ShardMetadataFilename file contains metadata for the created shard.
	ShardMetadataFilename string = ".shard_metadata"
	// ShardNumFilename is the name of the file that stores the number of shards.
	ShardNumFilename string = ".shard_num"
	// TransferStatusFilename file identifies if shard transfer to BigQuery is completed.
	TransferStatusFilename string = ".transfer_complete"
	projectID              string = "SCORECARD_PROJECT_ID"
	requestTopicURL        string = "SCORECARD_REQUEST_TOPIC_URL"
	requestSubscriptionURL string = "SCORECARD_REQUEST_SUBSCRIPTION_URL"
	bigqueryDataset        string = "SCORECARD_BIGQUERY_DATASET"
	completionThreshold    string = "SCORECARD_COMPLETION_THRESHOLD"
	shardSize              string = "SCORECARD_SHARD_SIZE"
	webhookURL             string = "SCORECARD_WEBHOOK_URL"
	metricExporter         string = "SCORECARD_METRIC_EXPORTER"
	ciiDataBucketURL       string = "SCORECARD_CII_DATA_BUCKET_URL"
	blacklistedChecks      string = "SCORECARD_BLACKLISTED_CHECKS"
	bigqueryTable          string = "SCORECARD_BIGQUERY_TABLE"
	resultDataBucketURL    string = "SCORECARD_DATA_BUCKET_URL"
	apiResultsBucketURL    string = "SCORECARD_API_RESULTS_BUCKET_URL"
	inputBucketURL         string = "SCORECARD_INPUT_BUCKET_URL"
	inputBucketPrefix      string = "SCORECARD_INPUT_BUCKET_PREFIX"
	// Raw results.
	rawBigqueryTable       string = "RAW_SCORECARD_BIGQUERY_TABLE"
	rawResultDataBucketURL string = "RAW_SCORECARD_DATA_BUCKET_URL"
)

var (
	// ErrorEmptyConfigValue indicates the value for the configuration option was empty.
	ErrorEmptyConfigValue = errors.New("config value set to empty")
	// ErrorValueConversion indicates an unexpected type was found for the value of the config option.
	ErrorValueConversion = errors.New("unexpected type, cannot convert value")
	//go:embed config.yaml
	configYAML []byte
)

//nolint:govet
type config struct {
	ProjectID              string  `yaml:"project-id"`
	ResultDataBucketURL    string  `yaml:"result-data-bucket-url"`
	RequestTopicURL        string  `yaml:"request-topic-url"`
	RequestSubscriptionURL string  `yaml:"request-subscription-url"`
	BigQueryDataset        string  `yaml:"bigquery-dataset"`
	BigQueryTable          string  `yaml:"bigquery-table"`
	CompletionThreshold    float32 `yaml:"completion-threshold"`
	WebhookURL             string  `yaml:"webhook-url"`
	CIIDataBucketURL       string  `yaml:"cii-data-bucket-url"`
	BlacklistedChecks      string  `yaml:"blacklisted-checks"`
	MetricExporter         string  `yaml:"metric-exporter"`
	ShardSize              int     `yaml:"shard-size"`
	// Raw results.
	RawResultDataBucketURL string            `yaml:"raw-result-data-bucket-url"`
	RawBigQueryTable       string            `yaml:"raw-bigquery-table"`
	APIResultsBucketURL    string            `yaml:"api-results-bucket-url"`
	InputBucketURL         string            `yaml:"input-bucket-url"`
	InputBucketPrefix      string            `yaml:"input-bucket-prefix"`
	Test                   map[string]string `yaml:"test"`
}

func getParsedConfigFromFile(byteValue []byte) (config, error) {
	var ret config
	err := yaml.Unmarshal(byteValue, &ret)
	if err != nil {
		return config{}, fmt.Errorf("error during yaml.Unmarshal: %w", err)
	}
	return ret, nil
}

func getReflectedValueFromConfig(byteValue []byte, fieldName string) (reflect.Value, error) {
	parsedConfig, err := getParsedConfigFromFile(byteValue)
	if err != nil {
		return reflect.ValueOf(parsedConfig), fmt.Errorf("error parsing config file: %w", err)
	}
	return reflect.ValueOf(parsedConfig).FieldByName(fieldName), nil
}

func getConfigValue(envVar string, byteValue []byte, fieldName string) (reflect.Value, error) {
	if val, present := os.LookupEnv(envVar); present {
		return reflect.ValueOf(val), nil
	}
	return getReflectedValueFromConfig(byteValue, fieldName)
}

func getStringConfigValue(envVar string, byteValue []byte, fieldName, configName string) (string, error) {
	value, err := getConfigValue(envVar, byteValue, fieldName)
	if err != nil {
		return "", fmt.Errorf("error getting config value %s: %w", configName, err)
	}
	if value.Kind() != reflect.String {
		return "", fmt.Errorf("%w: %s, %s", ErrorValueConversion, value.Type().Name(), configName)
	}
	if value.String() != "" {
		return value.String(), nil
	}
	return value.String(), fmt.Errorf("%w: %s", ErrorEmptyConfigValue, configName)
}

func getIntConfigValue(envVar string, byteValue []byte, fieldName, configName string) (int, error) {
	value, err := getConfigValue(envVar, byteValue, fieldName)
	if err != nil {
		return 0, fmt.Errorf("error getting config value %s: %w", configName, err)
	}

	switch value.Kind() {
	case reflect.String:
		//nolint:wrapcheck
		return strconv.Atoi(value.String())
	case reflect.Int:
		return int(value.Int()), nil
	default:
		return 0, fmt.Errorf("%w: %s, %s", ErrorValueConversion, value.Type().Name(), configName)
	}
}

func getFloat64ConfigValue(envVar string, byteValue []byte, fieldName, configName string) (float64, error) {
	value, err := getConfigValue(envVar, byteValue, fieldName)
	if err != nil {
		return 0, fmt.Errorf("error getting config value %s: %w", configName, err)
	}

	switch value.Kind() {
	case reflect.String:
		//nolint: wrapcheck, gomnd
		return strconv.ParseFloat(value.String(), 64)
	case reflect.Float32, reflect.Float64:
		return value.Float(), nil
	default:
		return 0, fmt.Errorf("%w: %s, %s", ErrorValueConversion, value.Type().Name(), configName)
	}
}

// envVarName converts a yaml map and nested key to an expected env variable name.
func envVarName(mapName, key string) string {
	base := fmt.Sprintf("%s_%s", mapName, key)
	underscored := strings.ReplaceAll(base, "-", "_")
	return strings.ToUpper(underscored)
}

func getMapConfigValue(byteValue []byte, fieldName, configName string) (map[string]string, error) {
	value, err := getReflectedValueFromConfig(byteValue, fieldName)
	if err != nil {
		return map[string]string{}, fmt.Errorf("error getting config value %s: %w", configName, err)
	}
	if value.Kind() != reflect.Map {
		return map[string]string{}, fmt.Errorf("%w: %s, %s", ErrorValueConversion, value.Type().Name(), configName)
	}
	ret := map[string]string{}
	iter := value.MapRange()
	for iter.Next() {
		key := iter.Key().String()
		val := iter.Value().String()
		if v, present := os.LookupEnv(envVarName(fieldName, key)); present {
			val = v
		}
		ret[key] = val
	}
	return ret, nil
}

// GetProjectID returns the cloud projectID for the cron job.
func GetProjectID() (string, error) {
	return getStringConfigValue(projectID, configYAML, "ProjectID", "project-id")
}

// GetResultDataBucketURL returns the bucketURL for storing cron job results.
func GetResultDataBucketURL() (string, error) {
	return getStringConfigValue(resultDataBucketURL, configYAML, "ResultDataBucketURL", "result-data-bucket-url")
}

// GetRequestTopicURL returns the topic name for sending cron job PubSub requests.
func GetRequestTopicURL() (string, error) {
	return getStringConfigValue(requestTopicURL, configYAML, "RequestTopicURL", "request-topic-url")
}

// GetRequestSubscriptionURL returns the subscription name of the PubSub topic for cron job reuests.
func GetRequestSubscriptionURL() (string, error) {
	return getStringConfigValue(requestSubscriptionURL, configYAML, "RequestSubscriptionURL", "request-subscription-url")
}

// GetBigQueryDataset returns the BQ dataset name to transfer cron job results.
func GetBigQueryDataset() (string, error) {
	return getStringConfigValue(bigqueryDataset, configYAML, "BigQueryDataset", "bigquery-dataset")
}

// GetBigQueryTable returns the table name to transfer cron job results.
func GetBigQueryTable() (string, error) {
	return getStringConfigValue(bigqueryTable, configYAML, "BigQueryTable", "bigquery-table")
}

// GetCompletionThreshold returns fraction of shards to be populated before transferring cron job results.
func GetCompletionThreshold() (float64, error) {
	return getFloat64ConfigValue(completionThreshold, configYAML, "CompletionThreshold", "completion-threshold")
}

// GetRawBigQueryTable returns the table name to transfer cron job results.
func GetRawBigQueryTable() (string, error) {
	return getStringConfigValue(rawBigqueryTable, configYAML,
		"RawBigQueryTable", "raw-bigquery-table")
}

// GetRawResultDataBucketURL returns the bucketURL for storing cron job's raw results.
func GetRawResultDataBucketURL() (string, error) {
	return getStringConfigValue(rawResultDataBucketURL, configYAML,
		"RawResultDataBucketURL", "raw-result-data-bucket-url")
}

// GetShardSize returns the shard_size for the cron job.
func GetShardSize() (int, error) {
	return getIntConfigValue(shardSize, configYAML, "ShardSize", "shard-size")
}

// GetWebhookURL returns the webhook URL to ping on a successful cron job completion.
func GetWebhookURL() (string, error) {
	url, err := getStringConfigValue(webhookURL, configYAML, "WebhookURL", "webhook-url")
	if err != nil && !errors.Is(err, ErrorEmptyConfigValue) {
		return url, err
	}
	return url, nil
}

// GetCIIDataBucketURL returns the bucket URL where CII data is stored.
func GetCIIDataBucketURL() (string, error) {
	url, err := getStringConfigValue(ciiDataBucketURL, configYAML, "CIIDataBucketURL", "cii-data-bucket-url")
	if err != nil && !errors.Is(err, ErrorEmptyConfigValue) {
		return url, err
	}
	return url, nil
}

// GetBlacklistedChecks returns a list of checks which are not to be run.
func GetBlacklistedChecks() ([]string, error) {
	checks, err := getStringConfigValue(blacklistedChecks, configYAML, "BlacklistedChecks", "blacklisted-checks")
	if err != nil && !errors.Is(err, ErrorEmptyConfigValue) {
		return nil, err
	}
	return strings.Split(checks, ","), nil
}

// GetMetricExporter returns the opencensus exporter type.
func GetMetricExporter() (string, error) {
	return getStringConfigValue(metricExporter, configYAML, "MetricExporter", "metric-exporter")
}

// GetAPIResultsBucketURL returns the bucket URL for storing cron job results.
func GetAPIResultsBucketURL() (string, error) {
	return getStringConfigValue(apiResultsBucketURL, configYAML,
		"APIResultsBucketURL", "api-results-bucket-url")
}

// GetInputBucketURL() returns the bucket URL for input files.
func GetInputBucketURL() (string, error) {
	return getStringConfigValue(inputBucketURL, configYAML, "InputBucketURL", "input-bucket-url")
}

// GetInputBucketPrefix() returns the prefix used when fetching files from a bucket.
func GetInputBucketPrefix() (string, error) {
	prefix, err := getStringConfigValue(inputBucketPrefix, configYAML, "InputBucketPrefix", "input-bucket-prefix")
	if err != nil && !errors.Is(err, ErrorEmptyConfigValue) {
		return "", err
	}
	return prefix, nil
}
