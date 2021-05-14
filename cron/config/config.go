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

// TODO(Azeem Shaikh): Use config.yaml to store these values and allow users to override these values using ENV vars.
const (
	ResultDataBucketURL string = "gs://ossf-scorecard-data"
	RequestTopicURL            = "gcppubsub://projects/openssf/topics/scorecard-batch-requests"
	InputReposFile             = "projects.csv"
	ShardNumFilename           = ".shard_num"
	ShardSize           int    = 250
)
