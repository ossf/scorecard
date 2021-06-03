# Copyright 2020 Security Scorecard Authors
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

# syntax = docker/dockerfile:1-experimental

FROM  golang:1.16.4@sha256:8a106c4b4005efb43c0ba4bb5763b84742c7e222bad5a8dff73cc9f7710c64ee AS base
WORKDIR /src
ENV CGO_ENABLED=0
COPY go.* ./
RUN go mod download
COPY . ./

FROM base AS build
ARG TARGETOS
ARG TARGETARCH
RUN --mount=type=cache,target=/root/.cache/go-build  CGO_ENABLED=0 go build -a -tags netgo -ldflags '-w -extldflags "-static"' -o /out/gitblobcache .

FROM gcr.io/distroless/base:nonroot@sha256:bc84925113289d139a9ef2f309f0dd7ac46ea7b786f172ba9084ffdb4cbd9490
COPY --from=build /out/gitblobcache /
ENTRYPOINT [ "/gitblobcache" ]
