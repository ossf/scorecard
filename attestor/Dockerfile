# Copyright 2022 OpenSSF Scorecard Authors
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

FROM golang@sha256:ea3d912d500b1ae0a691b2e53eb8a6345b579d42d7e6a64acca83d274b949740 AS base
WORKDIR /src/scorecard
COPY . ./

FROM base AS build
ARG TARGETOS
ARG TARGETARCH
RUN make build-attestor

FROM gcr.io/google-appengine/debian11@sha256:fed7dd5b2c4bbfb70bd26a277cdaff98dced71f113632ccd5451dcc013fce0a4
COPY --from=build /src/scorecard/attestor /
ENTRYPOINT [ "/scorecard-attestor" ]

