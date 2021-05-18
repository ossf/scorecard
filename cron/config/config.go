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
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strconv"

	"gopkg.in/yaml.v2"
)

const (
	ShardNumFilename       string = ".shard_num"
	resultDataBucketURL    string = "SCORECARD_DATA_BUCKET_URL"
	requestTopicURL        string = "SCORECARD_REQUEST_TOPIC_URL"
	requestSubscriptionURL string = "SCORECARD_REQUEST_SUBSCRIPTION_URL"
	inputReposFile         string = "SCORECARD_REPOS_FILE"
	shardSize              string = "SCORECARD_SHARD_SIZE"
	configYAML             string = "config.yaml"
)

var (
	ErrorEmptyConfigValue = errors.New("config value set to empty")
	ErrorValueConversion  = errors.New("unexpected type, cannot convert value")
)

type config struct {
	ResultDataBucketURL    string `yaml:"result-data-bucket-url"`
	RequestTopicURL        string `yaml:"request-topic-url"`
	RequestSubscriptionURL string `yaml:"request-subscription-url"`
	InputReposFile         string `yaml:"input-repos-file"`
	ShardSize              int    `yaml:"shard-size"`
}

func getParsedConfigFromFile(filename string) (config, error) {
	yamlFile, err := os.Open(filename)
	if err != nil {
		return config{}, fmt.Errorf("error during os.Open: %w", err)
	}
	defer yamlFile.Close()

	byteValue, err := ioutil.ReadAll(yamlFile)
	if err != nil {
		return config{}, fmt.Errorf("error during ioutil.ReadAll: %w", err)
	}

	var ret config
	err = yaml.Unmarshal(byteValue, &ret)
	if err != nil {
		return config{}, fmt.Errorf("error during yaml.Unmarshal: %w", err)
	}
	return ret, nil
}

func getConfigValue(envVar, filename, fieldName string) (reflect.Value, error) {
	if val, present := os.LookupEnv(envVar); present {
		return reflect.ValueOf(val), nil
	}
	parsedConfig, err := getParsedConfigFromFile(filename)
	if err != nil {
		return reflect.ValueOf(parsedConfig), fmt.Errorf("error parsing config file: %w", err)
	}
	return reflect.ValueOf(parsedConfig).FieldByName(fieldName), nil
}

func getStringConfigValue(envVar, filename, fieldName, configName string) (string, error) {
	value, err := getConfigValue(envVar, filename, fieldName)
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

func getIntConfigValue(envVar, filename, fieldName, configName string) (int, error) {
	value, err := getConfigValue(envVar, filename, fieldName)
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

func GetResultDataBucketURL() (string, error) {
	return getStringConfigValue(resultDataBucketURL, configYAML, "ResultDataBucketURL", "result-data-bucket-url")
}

func GetRequestTopicURL() (string, error) {
	return getStringConfigValue(requestTopicURL, configYAML, "RequestTopicURL", "request-topic-url")
}

func GetRequestSubscriptionURL() (string, error) {
	return getStringConfigValue(requestSubscriptionURL, configYAML, "RequestSubscriptionURL", "request-subscription-url")
}

func GetInputReposFile() (string, error) {
	return getStringConfigValue(inputReposFile, configYAML, "InputReposFile", "input-repos-file")
}

func GetShardSize() (int, error) {
	return getIntConfigValue(shardSize, configYAML, "ShardSize", "shard-size")
}
