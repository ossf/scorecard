# Copyright 2021 OpenSSF Scorecard Authors
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

FROM golang:1.20@sha256:344193a70dc3588452ea39b4a1e465a8d3c91f788ae053f7ee168cebf18e0a50 AS base
WORKDIR /src
ENV CGO_ENABLED=0
COPY go.* ./
RUN go mod download
COPY . ./

FROM base AS worker
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 make build-worker

FROM gcr.io/distroless/base:nonroot@sha256:99133cb0878bb1f84d1753957c6fd4b84f006f2798535de22ebf7ba170bbf434
COPY --from=worker /src/cron/internal/worker/worker cron/internal/worker/worker
ENTRYPOINT ["cron/internal/worker/worker"]
