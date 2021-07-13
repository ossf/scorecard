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

FROM  golang:1.17beta1@sha256:d7219d0176879a1731821ef5e9e3a740c3bed269870f22c753a883052abd0688 AS base
WORKDIR /src
ENV CGO_ENABLED=0
COPY go.* ./
RUN go mod download
COPY . ./

FROM base AS build
ARG TARGETOS
ARG TARGETARCH
RUN --mount=type=cache,target=/root/.cache/go-build  CGO_ENABLED=0 go build -a -tags netgo -ldflags '-w -extldflags "-static"' -o /out/gitblobcache .

FROM gcr.io/distroless/base:nonroot@sha256:38778ff7aa549bf6904c9d1c68bfe5946e96cac91dc32ba1f58e83bb9c9e6abe
COPY --from=build /out/gitblobcache /
ENTRYPOINT [ "/gitblobcache" ]
