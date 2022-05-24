// Copyright 2022 OpenSSF Authors
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
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"log"

	"github.com/ossf/scorecard-action/install/cli"
	"github.com/ossf/scorecard-action/install/options"
)

func main() {
	opts := options.New()
	if err := cli.New(opts).Execute(); err != nil {
		log.Fatalf("error during command execution: %v", err)
	}
}
