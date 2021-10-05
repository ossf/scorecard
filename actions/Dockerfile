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

# Testing: docker run -e INPUT_SARIF_FILE=results.sarif -e GITHUB_WORKSPACE=/ -e INPUT_POLICY_FILE="/policy.yml" -e INPUT_REPO_TOKEN=$GITHUB_AUTH_TOKEN -e GITHUB_REPOSITORY="ossf/scorecard" laurentsimon/scorecard-action:latest

# Use scorecard's official Docker image as base.
ARG DOCKER_SHA
FROM gcr.io/openssf/scorecard@sha256:${DOCKER_SHA} as base

# Build our image and update the root certs.
# TODO: use distroless.
FROM debian:9.5-slim
RUN apt-get update && \
    apt-get install -y --no-install-recommends jq ca-certificates

# Copy the scorecard binary from the official scorecard image.
#COPY --from=base /scorecard /scorecard
COPY ./scorecard /scorecard
# Copy a test policy for local testing.
COPY actions/policies/template.yml /policy.yml

# Our entry point.
# Note: the file is executable in the repo
# and permission carry over to the image.
ADD actions/entrypoint.sh /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
