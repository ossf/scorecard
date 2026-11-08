#!/bin/bash
# Copyright 2026 OpenSSF Scorecard Authors
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

# Compare Scorecard results between two commits with scdiff.
#
# Usage: scripts/scdiff.sh BASE HEAD [CHECKS]
#
# BASE and HEAD are any git revisions (tags, branches, or SHAs) present in
# this clone. CHECKS is an optional comma-separated list of checks. Each
# revision is built in a temporary worktree, so the current checkout is left
# alone. The repository list is cmd/internal/scdiff/repositories.txt from the
# current checkout; set SCDIFF_REPOS to use another file.
#
# Requires a GitHub token (GITHUB_AUTH_TOKEN) and, for the GitLab entries,
# GITLAB_AUTH_TOKEN. Exits 0 when the results match and non-zero when they
# differ or a step fails.

set -euo pipefail

if [ $# -lt 2 ] || [ $# -gt 3 ]; then
    printf "Usage: %s BASE HEAD [CHECKS]\n" "$0" >&2
    exit 2
fi

ROOT="$(git rev-parse --show-toplevel)"
REPOS="$(realpath "${SCDIFF_REPOS:-$ROOT/cmd/internal/scdiff/repositories.txt}")"
CHECKS="${3:-}"

if [ ! -s "$REPOS" ]; then
    printf "Error: repository list %s is missing or empty\n" "$REPOS" >&2
    exit 2
fi

BASE_SHA="$(git rev-parse --verify --quiet "$1^{commit}")" || {
    printf "Error: %s is not a commit in this clone\n" "$1" >&2
    exit 2
}
HEAD_SHA="$(git rev-parse --verify --quiet "$2^{commit}")" || {
    printf "Error: %s is not a commit in this clone\n" "$2" >&2
    exit 2
}

WORK="$(mktemp -d)"
cleanup() {
    git -C "$ROOT" worktree remove --force "$WORK/base" 2>/dev/null || true
    git -C "$ROOT" worktree remove --force "$WORK/head" 2>/dev/null || true
    rm -rf "$WORK"
}
trap cleanup EXIT

generate() {
    local side="$1" sha="$2"
    printf "Generating results for %s (%s)\n" "$side" "$sha" >&2
    git -C "$ROOT" worktree add --quiet --detach "$WORK/$side" "$sha"
    (
        cd "$WORK/$side"
        go run ./cmd/internal/scdiff generate \
            --repos "$REPOS" \
            --checks "$CHECKS" \
            --output "$WORK/$side.json"
    )
}

generate base "$BASE_SHA"
generate head "$HEAD_SHA"

printf "Comparing results\n" >&2
if ! (cd "$ROOT" && go run ./cmd/internal/scdiff compare "$WORK/base.json" "$WORK/head.json"); then
    printf "Results differ, or the comparison failed (see above)\n" >&2
    exit 1
fi
printf "Results match\n" >&2
