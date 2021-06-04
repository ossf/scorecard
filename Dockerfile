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


FROM  golang:1.16.5@sha256:6ff0e09b47d0e93823978da613c449056cda25c058ebc4031f93c05a76905a8c AS base
WORKDIR /src
ENV CGO_ENABLED=0
COPY go.* ./
RUN go mod download
COPY . ./

FROM base AS build
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 make build-scorecard

FROM gcr.io/distroless/base:nonroot@sha256:bc84925113289d139a9ef2f309f0dd7ac46ea7b786f172ba9084ffdb4cbd9490
COPY --from=build /src/scorecard /
ENTRYPOINT [ "/scorecard" ]
