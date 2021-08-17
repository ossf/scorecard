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
	resultDataBucketURL    string = "SCORECARD_DATA_BUCKET_URL"
	requestTopicURL        string = "SCORECARD_REQUEST_TOPIC_URL"
	requestSubscriptionURL string = "SCORECARD_REQUEST_SUBSCRIPTION_URL"
	bigqueryDataset        string = "SCORECARD_BIGQUERY_DATASET"
	bigqueryTable          string = "SCORECARD_BIGQUERY_TABLE"
	shardSize              string = "SCORECARD_SHARD_SIZE"
	webhookURL             string = "SCORECARD_WEBHOOK_URL"
	metricExporter         string = "SCORECARD_METRIC_EXPORTER"

	bigqueryTableV2       string = "SCORECARD_BIGQUERY_TABLEV2"
	resultDataBucketURLV2 string = "SCORECARD_DATA_BUCKET_URLV2"
)

var (
	// ErrorEmptyConfigValue indicates the value for the configuration option was empty.
	ErrorEmptyConfigValue = errors.New("config value set to empty")
	// ErrorValueConversion indicates an unexpected type was found for the value of the config option.
	ErrorValueConversion = errors.New("unexpected type, cannot convert value")
	//go:embed config.yaml
	configYAML []byte
)

//nolint
type config struct {
	ProjectID              string `yaml:"project-id"`
	ResultDataBucketURL    string `yaml:"result-data-bucket-url"`
	RequestTopicURL        string `yaml:"request-topic-url"`
	RequestSubscriptionURL string `yaml:"request-subscription-url"`
	BigQueryDataset        string `yaml:"bigquery-dataset"`
	BigQueryTable          string `yaml:"bigquery-table"`
	WebhookURL             string `yaml:"webhook-url"`
	MetricExporter         string `yaml:"metric-exporter"`
	ShardSize              int    `yaml:"shard-size"`
	// UPGRADEv2: to remove.
	ResultDataBucketURLV2 string `yaml:"result-data-bucket-url-v2"`
	BigQueryTableV2       string `yaml:"bigquery-table-v2"`
}

func getParsedConfigFromFile(byteValue []byte) (config, error) {
	var ret config
	err := yaml.Unmarshal(byteValue, &ret)
	if err != nil {
		return config{}, fmt.Errorf("error during yaml.Unmarshal: %w", err)
	}
	return ret, nil
}

func getConfigValue(envVar string, byteValue []byte, fieldName string) (reflect.Value, error) {
	if val, present := os.LookupEnv(envVar); present {
		return reflect.ValueOf(val), nil
	}
	parsedConfig, err := getParsedConfigFromFile(byteValue)
	if err != nil {
		return reflect.ValueOf(parsedConfig), fmt.Errorf("error parsing config file: %w", err)
	}
	return reflect.ValueOf(parsedConfig).FieldByName(fieldName), nil
}

func getStringConfigValue(envVar string, byteValue []byte, fieldName, configName string) (string, error) {
	value, err := getConfigValue(envVar, byteValue, fieldName)
	if err != nil {
		return "", fmt.Errorf("error getting config value %s: %w", configName, err)
	}
	if value.Type().Name() != "string" {
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
	switch value.Type().Name() {
	case "string":
		//nolint:wrapcheck
		return strconv.Atoi(value.String())
	case "int":
		return int(value.Int()), nil
	default:
		return 0, fmt.Errorf("%w: %s, %s", ErrorValueConversion, value.Type().Name(), configName)
	}
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

// GetBigQueryTableV2 returns the table name to transfer cron job results.
// UPGRADEv2: to remove.
func GetBigQueryTableV2() (string, error) {
	return getStringConfigValue(bigqueryTableV2, configYAML, "BigQueryTableV2", "bigquery-table-v2")
}

// GetResultDataBucketURLV2 returns the bucketURL for storing cron job results.
// UPGRADEv2: to remove.
func GetResultDataBucketURLV2() (string, error) {
	return getStringConfigValue(resultDataBucketURLV2, configYAML, "ResultDataBucketURLV2", "result-data-bucket-url-v2")
}

// GetShardSize returns the shard_size for the cron job.
func GetShardSize() (int, error) {
	return getIntConfigValue(shardSize, configYAML, "ShardSize", "shard-size")
}

// GetWebhookURL returns the webhook URL to ping on a successful cron job completion.
func GetWebhookURL() (string, error) {
	url, err := getStringConfigValue(webhookURL, configYAML, "WebhookURL", "webhook-url")
	if err != nil && !errors.As(err, &ErrorEmptyConfigValue) {
		return url, err
	}
	return url, nil
}

// GetMetricExporter returns the opencensus exporter type.
func GetMetricExporter() (string, error) {
	return getStringConfigValue(metricExporter, configYAML, "MetricExporter", "metric-exporter")
}
