# Copyright 2020 OpenSSF Scorecard Authors
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

# golang:1.19
FROM golang@sha256:25de7b6b28219279a409961158c547aadd0960cf2dcbc533780224afa1157fd4 AS build
WORKDIR /src
ENV CGO_ENABLED=0
COPY go.* ./
RUN go mod download
COPY . ./
RUN make build-scorecard

# https://github.com/chainguard-images/images/tree/main/images/static
# latest
FROM cgr.dev/chainguard/static:latest@sha256:17844d8faa68296eddddad9c1678edaf18cc556433d68c5bd3808a6efac200a8
COPY --from=build /src/scorecard /
ENTRYPOINT [ "/scorecard" ]
