// Copyright 2020 OpenSSF Scorecard Authors
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

// Package main of OSSF Scorecard.
package main

import (
	"fmt"
	"log"

	"github.com/google/osv-scanner/v2/pkg/osvscanner"
	"sigs.k8s.io/release-utils/version"

	"github.com/ossf/scorecard/v5/cmd"
	"github.com/ossf/scorecard/v5/options"
)

func main() {
	info := version.GetVersionInfo()
	actions := osvscanner.ExperimentalScannerActions{}
	actions.RequestUserAgent = fmt.Sprintf("scorecard-cli/%s", info.GitVersion)
	opts := options.New()
	if err := cmd.New(opts).Execute(); err != nil {
		log.Fatalf("error during command execution: %v", err)
	}
}
