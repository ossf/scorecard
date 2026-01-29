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

FROM golang:1.25.6@sha256:ce63a16e0f7063787ebb4eb28e72d477b00b4726f79874b3205a965ffd797ab2 AS base
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

FROM gcr.io/distroless/base:nonroot@sha256:cd961bbef4ecc70d2b2ff41074dd1c932af3f141f2fc00e4d91a03a832e3a658
COPY ./cron/internal/data/projects*csv cron/internal/data/
COPY ./cron/internal/data/gitlab-projects-releasetest.csv cron/internal/data/
COPY ./cron/internal/data/gitlab-projects.csv cron/internal/data/
COPY --from=shuffle /src/cron/internal/data/projects.release.csv cron/internal/data/projects.release.csv
COPY --from=controller /src/cron/internal/controller/controller cron/internal/controller/controller
ENTRYPOINT ["cron/internal/controller/controller"]
