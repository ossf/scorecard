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

FROM golang:1.24.1@sha256:fa145a3c13f145356057e00ed6f66fbd9bf017798c9d7b2b8e956651fe4f52da AS base
WORKDIR /src
ENV CGO_ENABLED=0
COPY go.* ./
RUN go mod download
COPY . ./

FROM base AS transfer
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 make build-bq-transfer

FROM gcr.io/distroless/base:nonroot@sha256:23fa4a8575bc94e586b94fb9b1dbce8a6d219ed97f805369079eebab54c2cb23
COPY --from=transfer /src/cron/internal/bq/data-transfer cron/internal/bq/data-transfer
ENTRYPOINT ["cron/internal/bq/data-transfer"]
