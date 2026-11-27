#!/usr/bin/env bash
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

# Lists the files that .github/workflows/links.yml passes to lychee.
#
# Usage: scripts/link-check-inputs.sh OUTDIR [BASE HEAD]
#
# Without BASE and HEAD, lists every in-scope file. With them, lists the
# in-scope files changed between the two commits, or every in-scope file
# if the link check itself changed. The list goes to OUTDIR/files.txt.
#
# lychee skips URLs inside HTML comments, so the comments in each listed
# Markdown file are copied to OUTDIR/comments/<path>.txt, with the rest
# of the file blanked so line numbers still match, and that copy is
# added to the list.
#
# OUTDIR/offline.txt lists the same files plus the generated docs, for
# checking relative links and anchors. Links in checks.yaml and def.yml
# only become relative links once rendered into the generated docs.
set -euo pipefail

if [[ $# -ne 1 && $# -ne 3 ]]; then
  echo "usage: $0 OUTDIR [BASE HEAD]" >&2
  exit 2
fi

outdir=$1
list=$outdir/files.txt
offline=$outdir/offline.txt
mkdir -p "$outdir"
: >"$list"

# Generated docs are left out of files.txt so external link failures
# point at the source to edit: docs/checks.md comes from
# docs/checks/internal/checks.yaml, and docs/probes.md from
# probes/*/def.yml. Meeting notes and openspec changes are
# point-in-time records.
generated=(docs/checks.md docs/probes.md)
in_scope='\.md$|^docs/checks/internal/checks\.yaml$|^probes/[^/]+/def\.yml$|^security-insights\.yml$|^\.github/ISSUE_TEMPLATE/[^/]+\.yml$'
out_of_scope='^(docs/checks\.md|docs/probes\.md|governance/meetings/|openspec/changes/)'
check_itself='^(lychee\.toml|\.github/workflows/links\.yml|scripts/link-check-inputs\.sh)$'

git_files() { git -c core.quotePath=false "$@"; }

if [[ $# -eq 3 ]]; then
  candidates=$(git_files diff --name-only --diff-filter=d "$2" "$3")
  if grep -qE "$check_itself" <<<"$candidates"; then
    candidates=$(git_files ls-files)
  fi
else
  candidates=$(git_files ls-files)
fi

# grep exits 1 when nothing matches, which is not an error here.
filter() { grep -E "$@" || [[ $? -eq 1 ]]; }
filter "$in_scope" <<<"$candidates" | filter -v "$out_of_scope" >"$list.tmp"

while IFS= read -r file; do
  echo "$file" >>"$list"
  [[ $file == *.md ]] || continue
  grep -q '<!--' "$file" || continue
  dest=$outdir/comments/$file.txt
  mkdir -p "$(dirname "$dest")"
  perl -0777 -pe 's{(<!--.*?-->)|[^\n]}{$1 // ""}gse' "$file" >"$dest"
  echo "$dest" >>"$list"
done <"$list.tmp"
rm "$list.tmp"

{ grep -v "^$outdir/comments/" "$list" || [[ $? -eq 1 ]]; } >"$offline"
printf '%s\n' "${generated[@]}" >>"$offline"

echo "$(wc -l <"$list" | tr -d ' ') files to check, listed in $list"
