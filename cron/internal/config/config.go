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

	// ConfigFlag is the name of the CLI flag to specify a config file.
	ConfigFlag string = "config"
	// ConfigDefault is the default value for the config file if not provided as a CLI arg.
	ConfigDefault string = "/etc/scorecard/config.yaml"
	// ConfigUsage is a description of the config CLI flag.
	ConfigUsage string = "Location of config file"

	projectID              string = "SCORECARD_PROJECT_ID"
	requestTopicURL        string = "SCORECARD_REQUEST_TOPIC_URL"
	requestSubscriptionURL string = "SCORECARD_REQUEST_SUBSCRIPTION_URL"
	bigqueryDataset        string = "SCORECARD_BIGQUERY_DATASET"
	completionThreshold    string = "SCORECARD_COMPLETION_THRESHOLD"
	shardSize              string = "SCORECARD_SHARD_SIZE"
	webhookURL             string = "SCORECARD_WEBHOOK_URL"
	metricExporter         string = "SCORECARD_METRIC_EXPORTER"
	bigqueryTable          string = "SCORECARD_BIGQUERY_TABLE"
	resultDataBucketURL    string = "SCORECARD_DATA_BUCKET_URL"
	apiResultsBucketURL    string = "SCORECARD_API_RESULTS_BUCKET_URL"
	inputBucketURL         string = "SCORECARD_INPUT_BUCKET_URL"
	inputBucketPrefix      string = "SCORECARD_INPUT_BUCKET_PREFIX"
)

var (
	// ErrorEmptyConfigValue indicates the value for the configuration option was empty.
	ErrorEmptyConfigValue = errors.New("config value set to empty")
	// ErrorValueConversion indicates an unexpected type was found for the value of the config option.
	ErrorValueConversion = errors.New("unexpected type, cannot convert value")
	// stores config file contents, set with ReadConfig.
	configYAML []byte
)

//nolint:govet
type config struct {
	ProjectID              string                       `yaml:"project-id"`
	ResultDataBucketURL    string                       `yaml:"result-data-bucket-url"`
	RequestTopicURL        string                       `yaml:"request-topic-url"`
	RequestSubscriptionURL string                       `yaml:"request-subscription-url"`
	BigQueryDataset        string                       `yaml:"bigquery-dataset"`
	BigQueryTable          string                       `yaml:"bigquery-table"`
	CompletionThreshold    float32                      `yaml:"completion-threshold"`
	WebhookURL             string                       `yaml:"webhook-url"`
	MetricExporter         string                       `yaml:"metric-exporter"`
	ShardSize              int                          `yaml:"shard-size"`
	InputBucketURL         string                       `yaml:"input-bucket-url"`
	InputBucketPrefix      string                       `yaml:"input-bucket-prefix"`
	AdditionalParams       map[string]map[string]string `yaml:"additional-params"`
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

func envVarName(subMapName, subKeyName string) string {
	base := fmt.Sprintf("%s_%s", subMapName, subKeyName)
	underscored := strings.ReplaceAll(base, "-", "_")
	return strings.ToUpper(underscored)
}

// getMapConfigValue returns a map from a nested yaml file. The values can be overridden if an env variable
// is set which corresponds to the name of the nested map and nested key. For example, the baz-qux value in
// the returned map can be overridden if FOO_BAR_BAZ_QUX is set.
// In the example below, "additional-params" is the fieldName, and "foo-bar" is the subMapName:
//
//	additional-params:
//	  foo-bar:
//	    baz-qux:
func getMapConfigValue(byteValue []byte, fieldName, configName, subMapName string) (map[string]string, error) {
	value, err := getReflectedValueFromConfig(byteValue, fieldName)
	if err != nil {
		return map[string]string{}, fmt.Errorf("error getting config value %s: %w", configName, err)
	}
	if value.Kind() != reflect.Map {
		return map[string]string{}, fmt.Errorf("%w: %s, %s", ErrorValueConversion, value.Type().Name(), configName)
	}
	subMap := value.MapIndex(reflect.ValueOf(subMapName))
	if subMap.Kind() != reflect.Map {
		return map[string]string{}, fmt.Errorf("%w: %s, %s", ErrorValueConversion, value.Type().Name(), configName)
	}
	ret := map[string]string{}
	iter := subMap.MapRange()
	for iter.Next() {
		subKey := iter.Key().String()
		val := iter.Value().String()
		if v, present := os.LookupEnv(envVarName(subMapName, subKey)); present {
			val = v
		}
		ret[subKey] = val
	}
	return ret, nil
}

func getScorecardParam(key string) (string, error) {
	s, err := GetScorecardValues()
	if err != nil {
		return "", err
	}
	return s[key], nil
}

// ReadConfig reads the contents of a configuration file for later use by getters.
// This function must be called before any other exported function.
func ReadConfig(filename string) error {
	var err error
	configYAML, err = os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("config file \"%s\": %w", filename, err)
	}
	return nil
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
	return getScorecardParam("raw-bigquery-table")
}

// GetRawResultDataBucketURL returns the bucketURL for storing cron job's raw results.
func GetRawResultDataBucketURL() (string, error) {
	return getScorecardParam("raw-result-data-bucket-url")
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
	return getScorecardParam("cii-data-bucket-url")
}

// GetBlacklistedChecks returns a list of checks which are not to be run.
func GetBlacklistedChecks() ([]string, error) {
	checks, err := getScorecardParam("blacklisted-checks")
	return strings.Split(checks, ","), err
}

// GetMetricExporter returns the opencensus exporter type.
func GetMetricExporter() (string, error) {
	return getStringConfigValue(metricExporter, configYAML, "MetricExporter", "metric-exporter")
}

// GetAPIResultsBucketURL returns the bucket URL for storing cron job results.
func GetAPIResultsBucketURL() (string, error) {
	return getScorecardParam("api-results-bucket-url")
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

func GetAdditionalParams(subMapName string) (map[string]string, error) {
	return getMapConfigValue(configYAML, "AdditionalParams", "additional-params", subMapName)
}

// GetScorecardValues() returns a map of key, value pairs containing additional, scorecard specific values.
func GetScorecardValues() (map[string]string, error) {
	return GetAdditionalParams("scorecard")
}

// GetCriticalityValues() returns a map of key, value pairs containing additional, criticality specific values.
func GetCriticalityValues() (map[string]string, error) {
	return GetAdditionalParams("criticality")
}
