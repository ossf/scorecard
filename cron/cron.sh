#!/bin/bash
# Copyright 2020 Security Scorecard Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http:#www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

SOURCE="${BASH_SOURCE[0]}"
if [[ ! -v TEST_CRON ]]; then
  input=$(dirname "$SOURCE")/projects.txt
elif [[ -z "$TEST_CRON" ]]; then
  input=$(dirname "$SOURCE")/projects.txt
else
  input=$(dirname "$SOURCE")/test_projects.txt
fi


output=$(date +"%m-%d-%Y").json
touch "$output"
echo "{ \"results\": [" >> "$output"

# sort and uniqify these, in case there are duplicates
# shellcheck disable=SC2002
projects=$(cat "$input" | sort | uniq)
while read -r proj; do
    if [[ $proj =~ ^#.* ]]; then
        continue
    fi
    echo "$proj"
    ../scorecard --repo="$proj" --show-details --format=json >> "$output"
    echo "," >> "$output"
done <<< "$projects"
sed -i '$d' "$output" # removing the trailing comma which will be last line.
echo "]}" >> "$output"

gsutil cp "$output" gs://"$GCS_BUCKET"
# Also copy the most recent run into a "latest.json" file
gsutil cp gs://"$GCS_BUCKET"/"$output" gs://"$GCS_BUCKET"/latest.json
