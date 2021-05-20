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
	_ "embed" // Used to embed config.yaml
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"

	"gopkg.in/yaml.v2"
)

const (
	ShardNumFilename       string = ".shard_num"
	projectID              string = "SCORECARD_PROJECT_ID"
	resultDataBucketURL    string = "SCORECARD_DATA_BUCKET_URL"
	requestTopicURL        string = "SCORECARD_REQUEST_TOPIC_URL"
	requestSubscriptionURL string = "SCORECARD_REQUEST_SUBSCRIPTION_URL"
	bigqueryDataset        string = "SCORECARD_BIGQUERY_DATASET"
	bigqueryTable          string = "SCORECARD_BIQQUERY_TABLE"
	shardSize              string = "SCORECARD_SHARD_SIZE"
)

var (
	ErrorEmptyConfigValue = errors.New("config value set to empty")
	ErrorValueConversion  = errors.New("unexpected type, cannot convert value")
	//go:embed config.yaml
	configYAML []byte
)

type config struct {
	ProjectID              string `yaml:"project-id"`
	ResultDataBucketURL    string `yaml:"result-data-bucket-url"`
	RequestTopicURL        string `yaml:"request-topic-url"`
	RequestSubscriptionURL string `yaml:"request-subscription-url"`
	BigQueryDataset        string `yaml:"bigquery-dataset"`
	BigQueryTable          string `yaml:"bigquery-table"`
	ShardSize              int    `yaml:"shard-size"`
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

func GetProjectID() (string, error) {
	return getStringConfigValue(projectID, configYAML, "ProjectID", "project-id")
}

func GetResultDataBucketURL() (string, error) {
	return getStringConfigValue(resultDataBucketURL, configYAML, "ResultDataBucketURL", "result-data-bucket-url")
}

func GetRequestTopicURL() (string, error) {
	return getStringConfigValue(requestTopicURL, configYAML, "RequestTopicURL", "request-topic-url")
}

func GetRequestSubscriptionURL() (string, error) {
	return getStringConfigValue(requestSubscriptionURL, configYAML, "RequestSubscriptionURL", "request-subscription-url")
}

func GetBigQueryDataset() (string, error) {
	return getStringConfigValue(bigqueryDataset, configYAML, "BigQueryDataset", "bigquery-dataset")
}

func GetBigQueryTable() (string, error) {
	return getStringConfigValue(bigqueryTable, configYAML, "BigQueryTable", "bigquery-table")
}

func GetShardSize() (int, error) {
	return getIntConfigValue(shardSize, configYAML, "ShardSize", "shard-size")
}
