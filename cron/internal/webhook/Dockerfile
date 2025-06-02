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

FROM golang:1.24.3@sha256:81bf5927dc91aefb42e2bc3a5abdbe9bb3bae8ba8b107e2a4cf43ce3402534c6 AS base
WORKDIR /src
ENV CGO_ENABLED=0
COPY go.* ./
RUN go mod download
COPY . ./

FROM base AS webhook
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 make build-webhook

FROM gcr.io/distroless/base:nonroot@sha256:5c9b112e85b26632c6ba9ac874be9c6b20d61599f6087534ce2b9feeb7f6babf
COPY --from=webhook /src/cron/internal/webhook/webhook cron/internal/webhook/webhook
ENTRYPOINT ["cron/internal/webhook/webhook"]
