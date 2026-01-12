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

FROM golang:1.25.5@sha256:6cc2338c038bc20f96ab32848da2b5c0641bb9bb5363f2c33e9b7c8838f9a208 AS base
WORKDIR /src
ENV CGO_ENABLED=0
COPY go.* ./
RUN go mod download
COPY . ./

FROM base AS webhook
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 make build-webhook

FROM gcr.io/distroless/base:nonroot@sha256:cd961bbef4ecc70d2b2ff41074dd1c932af3f141f2fc00e4d91a03a832e3a658
COPY --from=webhook /src/cron/internal/webhook/webhook cron/internal/webhook/webhook
ENTRYPOINT ["cron/internal/webhook/webhook"]
