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

# golang:1.19
FROM golang@sha256:4b1fc02d16fca272e5e6e6adc98396219b43ef663a377eef4a97e881d364393f AS base
WORKDIR /src
ENV CGO_ENABLED=0
COPY go.* ./
RUN go mod download
COPY . ./

FROM base AS shuffle
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 make build-shuffler
RUN ./cron/internal/shuffle/shuffle 2500 cron/internal/data/projects.csv cron/internal/data/projects.release.csv

FROM base AS controller
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 make build-controller

FROM gcr.io/distroless/base:nonroot@sha256:99133cb0878bb1f84d1753957c6fd4b84f006f2798535de22ebf7ba170bbf434
COPY ./cron/internal/data/projects*csv cron/internal/data/
COPY ./cron/internal/data/gitlab-projects-selected.csv cron/internal/data/
COPY --from=shuffle /src/cron/internal/data/projects.release.csv cron/internal/data/projects.release.csv
COPY --from=controller /src/cron/internal/controller/controller cron/internal/controller/controller
ENTRYPOINT ["cron/internal/controller/controller"]
