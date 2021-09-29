#!/bin/bash
# Copyright 2021 Security Scorecard Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -euo pipefail

# https://docs.github.com/en/actions/learn-github-actions/environment-variables
# GITHUB_EVENT_PATH contains the json file for the event.
# GITHUB_SHA contains the commit hash.
# GITHUB_WORKSPACE contains the repo folder.
# GITHUB_EVENT_NAME contains the event name.
# GITHUB_ACTIONS is true in GitHub env.

export GITHUB_AUTH_TOKEN="$INPUT_REPO_TOKEN"
export SCORECARD_V3=1
export SCORECARD_POLICY_FILE="$INPUT_POLICY_FILE"
export SCORECARD_SARIF_FILE="$INPUT_SARIF_FILE"

# Note: this will fail if we push to a branch on the same repo, so it will show as failing
# on forked repos.
if [[ "$GITHUB_EVENT_NAME" != "pull_request"* ]] && ! [[ "$GITHUB_REF" =~ ^refs/heads/(main|master)$ ]]; then
    echo "$GITHUB_REF not supported with '$GITHUB_EVENT_NAME' event."
    echo "Only the default branch is supported"
    exit 1
fi

# It's important to change directories here, to ensure
# the files in SARIF start at the source of the repo.
# This allows GitHub to highlight the file.
cd "$GITHUB_WORKSPACE"
/scorecard --repo="$GITHUB_REPOSITORY" --format sarif --show-details --policy="$SCORECARD_POLICY_FILE" > "$SCORECARD_SARIF_FILE"
jq '.' "$SCORECARD_SARIF_FILE"