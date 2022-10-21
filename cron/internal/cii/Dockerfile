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

# golang:1.19
FROM golang@sha256:25de7b6b28219279a409961158c547aadd0960cf2dcbc533780224afa1157fd4 AS base
WORKDIR /src
ENV CGO_ENABLED=0
COPY go.* ./
RUN go mod download
COPY . ./

FROM base AS cii
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 make build-cii-worker

FROM gcr.io/distroless/base:nonroot@sha256:a6441d19dfd755cd0782a4480fbeeb196a049c52d23b0dba81539098fda25d16
COPY --from=cii /src/cron/internal/cii/cii-worker cron/internal/cii/cii-worker
ENTRYPOINT ["cron/internal/cii/cii-worker"]
