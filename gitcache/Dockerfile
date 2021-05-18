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

FROM  golang:1.16.4@sha256:96129f3766cf8b69c32bd0b4be8fbf8f19ef143555b76a7f7b0046ed142e296a AS base
WORKDIR /src
ENV CGO_ENABLED=0
COPY go.* ./
RUN go mod download
COPY . ./

FROM base AS build
ARG TARGETOS
ARG TARGETARCH
RUN --mount=type=cache,target=/root/.cache/go-build  CGO_ENABLED=0 go build -a -tags netgo -ldflags '-w -extldflags "-static"' -o /out/gitblobcache .

FROM gcr.io/distroless/base:nonroot@sha256:fba21a4d26b3b4f24dd86e756f17107537dfc0f19a7882a8cd7914da1ccd8f92
COPY --from=build /out/gitblobcache /
ENTRYPOINT [ "/gitblobcache" ]
