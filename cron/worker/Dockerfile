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

FROM golang@sha256:3c4de86eec9cbc619cdd72424abd88326ffcf5d813a8338a7743c55e5898734f AS base
WORKDIR /src
ENV CGO_ENABLED=0
COPY go.* ./
RUN go mod download
COPY . ./

FROM base AS worker
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 make build-worker

FROM gcr.io/distroless/base:nonroot@sha256:792dfe78a236dfb6fb180a250d105e0a03585dcbc73f8fce033fe62d4fd59bcb
COPY --from=worker /src/cron/worker/worker cron/worker/worker
ENTRYPOINT ["cron/worker/worker"]
