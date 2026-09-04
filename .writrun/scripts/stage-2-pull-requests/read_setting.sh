#!/usr/bin/env bash
# read_setting.sh — prints one value from .writrun/settings.json.
#
# Usage: read_setting.sh <address> [--origin]
#   Run from the repository root; the path is relative to it.
#
#   read_setting.sh stage                     a top-level key, bare
#   read_setting.sh stage_2.pr_title_style    a key inside its section
#   read_setting.sh stage --origin            the value, a tab, then
#                                             `declared` or `default`
#
# **--origin is how a reader tells a choice from a documented default.**
# Both print identically without it, which is right for a caller that
# only wants the value to act on — and wrong for one that renders the
# file for a person, where "this is what the project decided" and "this
# is what nobody decided" are different sentences (session_card.sh).
# A value carried over by one of the two rename bridges below is
# `declared`: the adopter wrote it, under the name of the day.
#
# **The address, not the name, is a key's identity.** The file carries one
# top-level `stage` and one object per stage holding the keys that stage's
# readers act on (docs/technical/settings.md#settings), so the same name in
# two sections would be two keys.
#
# The file is JSON in a restricted shape — two levels and nothing deeper,
# one `"key": value` per line, values `true`, `false`, an unquoted integer,
# or a double-quoted string
# (docs/technical/settings.md#the-shape-is-a-checked-contract). That
# restriction is what lets this read it with `awk` alone: one nesting
# level, entered and left on lines of fixed shape, is a two-state line
# reader. Requiring `jq` would be this project's first runtime dependency,
# which the toolchain is built to avoid.
#
# **Absence is not an error.** No file, or no such key, prints the
# documented default — the same posture list_tasks.sh takes when no forge
# answers. An adopter who deletes the file keeps working, with the
# behaviour the schema documents; a reader that failed instead would make
# the file mandatory, which it is not.
#
# An address the schema does not document has no documented default, so it
# prints nothing. Exit is still 0: this reads a value, it does not judge
# the file — check_settings.sh does that.
#
# Exit codes: 0 always, except 3 for a usage error.
#
# Portable bash 3.2, POSIX awk/sed. See the standing rule in
# docs/technical/decisions/.

set -euo pipefail

KEY=""
WITH_ORIGIN=""
for arg in "$@"; do
  case "$arg" in
    --origin) WITH_ORIGIN=yes ;;
    -*)       echo "usage: read_setting.sh <address> [--origin]" >&2; exit 3 ;;
    *)        [ -z "$KEY" ] || { echo "usage: read_setting.sh <address> [--origin]" >&2; exit 3; }
              KEY="$arg" ;;
  esac
done
[ -n "$KEY" ] || { echo "usage: read_setting.sh <address> [--origin]" >&2; exit 3; }

# emit <value> <declared|default> — the one place the output shape lives,
# so the two forms can never drift apart.
emit() {
  if [ -n "$WITH_ORIGIN" ]; then printf '%s\t%s\n' "$1" "$2"; else printf '%s\n' "$1"; fi
}

SETTINGS=".writrun/settings.json"
LEGACY=".writrun/conventions/settings.json"

# The documented defaults, addressed the way the schema addresses them
# (docs/technical/settings.md#settings). Each is the behaviour from before
# its key existed, so a project without the file, or without the key,
# behaves exactly as it did.
default_for() {
  case "$1" in
    stage)                   printf '3' ;;
    stage_1.spec_required)   printf 'when-warranted' ;;
    stage_1.decisions_style) printf 'per-subsystem' ;;
    stage_1.product_layout)  printf 'by-concept' ;;
    stage_1.provenance_ledger) printf 'false' ;;
    stage_2.agent_coauthor)  printf 'true' ;;
    stage_2.auto_commit)     printf 'true' ;;
    stage_2.auto_pr)         printf 'true' ;;
    stage_2.auto_push)       printf 'true' ;;
    stage_2.pr_title_style)  printf 'conventional' ;;
  esac
}

# The section is everything before the first dot, the name everything
# after it; a bare address names the top level, whose section is "".
case "$KEY" in
  *.*) SECTION="${KEY%%.*}"; NAME="${KEY#*.}" ;;
  *)   SECTION=""; NAME="$KEY" ;;
esac

# The migration bridge (decision 0053): a file left at the old address is
# read flat, under the contract frozen at the move — shape included, so a
# sectioned address finds its key at the top level there. The check is
# what names the move; this is what keeps the adopter's choice honoured
# until they make it.
FILE="$SETTINGS"
if [ ! -f "$SETTINGS" ]; then
  if [ -f "$LEGACY" ]; then
    FILE="$LEGACY"
    SECTION=""
  else
    emit "$(default_for "$KEY")" default; exit 0
  fi
fi

# pair <section> <name> — the raw text after the addressed key's colon,
# empty when the file does not carry it. A section is opened by a line
# whose value is a bare `{` and closed by a line that is a bare `}`, so a
# quoted value holding a brace or a dot is never mistaken for either: the
# state machine reads line shapes, never value content.
pair() {
  awk -v want_sec="$1" -v want_key="$2" '
    /^[[:space:]]*"[^"]*"[[:space:]]*:[[:space:]]*\{[[:space:]]*$/ {
      sec = $0
      sub(/^[^"]*"/, "", sec); sub(/".*$/, "", sec)
      next
    }
    /^[[:space:]]*\}[[:space:]]*,?[[:space:]]*$/ { sec = ""; next }
    /^[[:space:]]*"[^"]*"[[:space:]]*:/ {
      if (found || sec != want_sec) next
      key = $0
      sub(/^[^"]*"/, "", key); sub(/".*$/, "", key)
      if (key != want_key) next
      val = $0
      sub(/^[[:space:]]*"[^"]*"[[:space:]]*:[[:space:]]*/, "", val)
      print val
      found = 1
    }
  ' "$FILE"
}

raw=$(pair "$SECTION" "$NAME")

if [ -z "$raw" ] && [ "$KEY" = "stage" ]; then
  # The older bridge, kept: a settings file written before the rename says
  # `level`, and reading it as "absent, so the default" would turn every
  # workflow on for an adopter who chose the full opt-out.
  # check_settings.sh names the rename; this keeps their choice honoured
  # until they make it.
  legacy=$(pair "" level | sed 's/[[:space:]]*$//; s/,$//; s/^"//; s/"$//')
  case "$legacy" in
    tasks-and-specs) emit 1 declared; exit 0 ;;
    pull-requests)   emit 2 declared; exit 0 ;;
    github-issues)   emit 3 declared; exit 0 ;;
  esac
fi

if [ -z "$raw" ] && [ "$KEY" = "stage_2.agent_coauthor" ]; then
  # The same bridge, for the same reason (spec-0035): a settings file
  # written before the rename says `credit_ai`, and reading it as
  # "absent, so the default" would print `true` for an adopter who wrote
  # `false` to switch the obligation off — a deliberate opt-out inverted
  # into an obligation, and check_observance.sh flipped from forbidding
  # credit to demanding trailers. The value carries over unchanged,
  # which is what check_settings.sh's rename fault promises;
  # this is where the promise is kept.
  raw=$(pair "$SECTION" credit_ai)
fi

if [ -z "$raw" ]; then
  emit "$(default_for "$KEY")" default; exit 0
fi

# Trailing whitespace, then the separating comma, then the quotes.
val=$(printf '%s' "$raw" | sed 's/[[:space:]]*$//')
val=${val%,}
val=$(printf '%s' "$val" | sed 's/[[:space:]]*$//')
case "$val" in
  \"*\") val=${val#\"}; val=${val%\"} ;;
esac

emit "$val" declared
