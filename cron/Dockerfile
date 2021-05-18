# syntax = docker/dockerfile:1-experimental
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


FROM  golang:1.16.4@sha256:6f0b0a314b158ff6caf8f12d7f6f3a966500ec6afb533e986eca7375e2f7560f AS base
WORKDIR /src
ENV CGO_ENABLED=0
COPY go.* ./
RUN go mod download
COPY . ./

FROM base AS build
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 make build-scorecard

FROM base AS cron
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 make build-cron

FROM gcr.io/google.com/cloudsdktool/cloud-sdk:slim@sha256:acb34ca33bc83255110d16a749b0da618a96fcc481348ecfe614b3edcabdeaa6
COPY ./cron/projects.txt /cron/projects.txt
COPY --from=build /src/scorecard /
COPY --from=cron /src/cron/scorecardcron ./cron/cron
ENTRYPOINT [ "/cron/cron", "/cron/projects.txt"] 
