#!/usr/bin/env bash
# check_deltas.sh — verifies a diff against a spec's Proposed changes sections.
#
# Usage: check_deltas.sh spec-nnn[,spec-mmm,...] [diff-range]
#   diff-range defaults to comparing the working tree against HEAD.
#
# Several ids, comma-separated, are for the change that implements several
# specs at once — completing a multi-spec task in one pull request. The
# contract then is: every promise of every listed spec is honoured
# (MISSING stays per spec), and every permanent doc the diff touches was
# promised by at least one of them (UNDECLARED checks the union). Checking
# each spec alone against the whole diff would report every sibling's
# promise as undeclared — an invariant nobody stated.
#
# Exit codes:
#   0 — every promised path was touched, and no untouched permanent doc
#       outside the promise list was modified.
#   1 — a promised path was NOT touched by the diff (incomplete change).
#   2 — a permanent doc (anything under docs/) was touched but was not
#       listed in either Proposed changes section (undeclared change).
#   3 — usage error, or spec file not found.
#
# This script is intentionally dumb: it does not understand markdown
# semantics, section anchors as targets for navigation, or intent. It only
# checks "was this path present in the diff, yes or no". That's the point —
# the check is meant to be mechanical, not judgment-based.

set -euo pipefail

# The bullet reader is queue_lib.sh's, shared with the promise gate: one
# copy, so a bullet that gate refused cannot be read here, and one it
# never saw cannot pass here (decision 0078). The skills and the scripts
# ship as one tree, so the path is always there to source.
. "$(cd "$(dirname "$0")" && pwd)/../../scripts/stage-2-pull-requests/queue_lib.sh"

SPEC_IDS="${1:-}"
DIFF_RANGE="${2:-HEAD}"

if [[ -z "$SPEC_IDS" ]]; then
  echo "Usage: check_deltas.sh spec-nnn[,spec-mmm,...] [diff-range]" >&2
  exit 3
fi

# No associative arrays — macOS ships bash 3.2, and these scripts promise
# to run there. The file is re-resolved where needed instead.
# <id>.md or <id>-<subject>.md — the subject slug is not identity, and the
# generator writes one, so a resolver that reads only the bare form finds
# nothing for every spec this project now creates.
spec_file_of() {
  find work/specs \( -iname "$1.md" -o -iname "$1-*.md" \) 2>/dev/null | head -n1
}

SPEC_LIST=$(printf '%s' "$SPEC_IDS" | tr ',' ' ')
for id in $SPEC_LIST; do
  if [[ -z "$(spec_file_of "$id")" ]]; then
    echo "Spec file for ${id} not found under work/specs/" >&2
    exit 3
  fi
done

# The two Proposed-changes sections, as `ql_promised_paths` reads them —
# a bullet's backticked path, anchor stripped, written relative to docs/
# (product/..., technical/..., about.md) per the schema in
# docs/technical/schemas/README.md. `git diff --name-only` reports
# relative to the repository root, and the trailing `sed 's|^|docs/|'`
# normalises the former to the latter — without it every promised path
# reports MISSING and every touched doc reports UNDECLARED,
# simultaneously.
#
# promised_of <spec-id> — both sections, normalised.
promised_of() {
  local f
  f=$(spec_file_of "$1")
  ql_promised_paths < "$f" | sed 's|^|docs/|'
}

ALL_PROMISED=$(for id in $SPEC_LIST; do promised_of "$id"; done | sed '/^$/d' | sort -u)

# A failing git call must not be swallowed: an empty file list is
# indistinguishable from "nothing was touched", which would report every
# promised path as MISSING and look like a real result.
err_tmp=$(mktemp "${TMPDIR:-/tmp}/check_deltas.XXXXXX")
if ! CHANGED_FILES=$(git diff --name-only "$DIFF_RANGE" 2>"$err_tmp"); then
  echo "git diff --name-only ${DIFF_RANGE} failed:" >&2
  head -n 2 "$err_tmp" >&2
  rm -f "$err_tmp"
  echo "(no git history yet, or a bad diff range — this check cannot run;" >&2
  echo " verify by hand against the spec's Proposed changes sections)" >&2
  exit 3
fi
rm -f "$err_tmp"

status=0

# Check every promised path was actually touched — per spec, so the report
# names which contract went unhonoured.
#
# A promise ending in `/` names a folder — the shape a rename or a
# chapter-wide sweep is promised in. It is honoured when the diff touches
# anything under it, and it declares everything under it.
for id in $SPEC_LIST; do
  while IFS= read -r p; do
    [[ -z "$p" ]] && continue
    if [[ "$p" == */ ]]; then
      if ! grep -q "^${p}" <<< "$CHANGED_FILES"; then
        echo "MISSING: ${id}'s promised change under '$p' not found in diff" >&2
        status=1
      fi
      continue
    fi
    if ! grep -qxF "$p" <<< "$CHANGED_FILES"; then
      echo "MISSING: ${id}'s promised change to '$p' not found in diff" >&2
      status=1
    fi
  done <<< "$(promised_of "$id")"
done

# Check no permanent doc outside the promise list was touched. Permanent
# is structural: everything under docs/ — the queue lives in work/, so no
# enumeration of docs/'s inside is needed, and adopters may shape it.
while IFS= read -r f; do
  [[ -z "$f" ]] && continue
  # docs/writrun-instructions.md is process metadata, not project truth —
  # a spec never has to promise it.
  [[ "$f" == "docs/writrun-instructions.md" ]] && continue
  case "$f" in
    docs/*)
      declared=false
      if grep -qxF "$f" <<< "$ALL_PROMISED"; then
        declared=true
      else
        while IFS= read -r p; do
          [[ "$p" == */ ]] || continue
          case "$f" in "$p"*) declared=true; break ;; esac
        done <<< "$ALL_PROMISED"
      fi
      if [[ "$declared" != "true" ]]; then
        echo "UNDECLARED: '$f' was modified but not listed in the Proposed changes of ${SPEC_IDS}" >&2
        # Both failures print; the exit code reports MISSING when both are
        # present, since a forgotten doc update is the drift this exists to
        # catch. Never overwrite a 1 with a 2.
        [[ "$status" -eq 0 ]] && status=2
      fi
      ;;
  esac
done <<< "$CHANGED_FILES"

if [[ "$status" -eq 0 ]]; then
  echo "OK: diff matches the promised deltas of ${SPEC_IDS}."
fi

exit "$status"
