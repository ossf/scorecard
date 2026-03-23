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

# Verify PR title contains required emoji prefix
# This replaces the deprecated kubebuilder-release-tools action

set -euo pipefail

PR_TITLE="${1:-}"

if [ -z "$PR_TITLE" ]; then
    printf "Error: PR title is required as first argument\n" >&2
    exit 1
fi

# Sanity check: limit title length to prevent DoS
if [ "${#PR_TITLE}" -gt 500 ]; then
    printf "Error: PR title exceeds maximum length of 500 characters\n" >&2
    exit 1
fi

# Remove WIP prefix if present (case-insensitive)
TITLE=$(echo "$PR_TITLE" | sed -E 's/^[[:space:]]*\W*WIP\W*[[:space:]]*//' | sed 's/^[[:space:]]*//')

# Remove tag prefix like [tag] if present
TITLE=$(echo "$TITLE" | sed -E 's/^\[[\w\-\.]*\][[:space:]]*//')

# Check for required emoji prefixes
# Using both emoji and :emoji: formats for compatibility
if echo "$TITLE" | grep -qE '^(⚠|:warning:)'; then
    PR_TYPE="⚠ Breaking change"
elif echo "$TITLE" | grep -qE '^(✨|:sparkles:)'; then
    PR_TYPE="✨ Non-breaking feature"
elif echo "$TITLE" | grep -qE '^(🐛|:bug:)'; then
    PR_TYPE="🐛 Patch fix"
elif echo "$TITLE" | grep -qE '^(📖|:book:)'; then
    PR_TYPE="📖 Documentation changes (user or developer)"
elif echo "$TITLE" | grep -qE '^(🌱|:seedling:)'; then
    PR_TYPE="🌱 Infra/Tests/Other"
elif echo "$TITLE" | grep -qE '^(👻|:ghost:)'; then
    PR_TYPE="👻 No release note"
else
    printf "❌ PR Title Verification Failed\n\n"
    printf "Title: '%s'\n\n" "$PR_TITLE"
    printf "No matching PR type indicator found in title.\n\n"
    printf "You need to have one of these as the prefix of your PR title:\n\n"
    printf "%s\n" "- Breaking change: ⚠ (':warning:')"
    printf "%s\n" "- Non-breaking feature: ✨ (':sparkles:')"
    printf "%s\n" "- Patch fix: 🐛 (':bug:')"
    printf "%s\n" "- Documentation changes (user or developer): 📖 (':book:')"
    printf "%s\n" "- Infra/Tests/Other: 🌱 (':seedling:')"
    printf "%s\n\n" "- No release note: 👻 (':ghost:')"
    printf "More details: https://github.com/ossf/scorecard/blob/main/CONTRIBUTING.md#pr-process\n"
    exit 1
fi

printf "✅ PR Title Verification Passed\n\n"
printf "Found %s\n\n" "$PR_TYPE"
printf "Final title: %s\n" "$TITLE"
