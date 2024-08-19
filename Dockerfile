# Copyright 2020 OpenSSF Scorecard Authors
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

FROM golang:1.22.6@sha256:367bb5295d3103981a86a572651d8297d6973f2ec8b62f716b007860e22cbc25 AS base
WORKDIR /src
ENV CGO_ENABLED=0
COPY go.* ./
RUN go mod download
COPY . ./

FROM base AS build
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 make build-scorecard

FROM cgr.dev/chainguard/static@sha256:5e9c88174a28c259c349f308dd661a6ec61ed5f8c72ecfaefb46cceb811b55a1
COPY --from=build /src/scorecard /
ENTRYPOINT [ "/scorecard" ]
