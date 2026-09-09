#!/usr/bin/env bash
# resolve_doc.sh — prints which file answers a project-home address.
#
# Usage: resolve_doc.sh <project-home path> [--origin]
#   Run from the repository root; the path is relative to it.
#
#   resolve_doc.sh writrun/gates.md                  the file in force
#   resolve_doc.sh writrun/conventions/commits.md    the same, one deeper
#   resolve_doc.sh writrun/gates.md --origin         the path, a tab, then
#                                                    `declared` or `default`
#
# **The project's home is the address; the kit's home is the fallback.**
# Every text WritRun answers on a project's behalf is addressed at
# `writrun/<path>` and read from there — unless that file defers, in
# which case the answer is `.writrun/defaults/<path>`, the same relative
# name in the kit's own home
# (docs/product/adoption.md#two-homes). Callers therefore name one
# address and never branch on which home won.
#
# **Deferring is positional.** A file whose FIRST line is exactly
# `/// writrun:default` defers; anything else is the project's answer,
# whole. The rule is the draft chapter's, for the reason that one is
# positional too: a file that *documents* the marker names it in its
# prose, and a reader searching the whole file would call the
# documentation a stub
# (docs/product/stage-1-tasks-and-specs/authoring.md#a-chapter-that-is-not-a-rule-yet).
#
# **An absent file defers too.** A project that deleted a stub, or
# adopted before the stub existed, is answered by the default rather
# than by nothing — the same posture read_setting.sh takes when a key
# is missing. What has no answer at all is an address the kit ships no
# default for: that is a caller's mistake, and it is loud, because
# printing nothing would let a check pass by having read no rule.
#
# **--origin is how a renderer tells an answer from a default nobody
# wrote**, exactly as it is for a setting. Both print the same path
# without it, which is right for a caller that only wants to read the
# file.
#
# Exit codes: 0 resolved; 3 a usage error; 4 no file and no default —
# the address names nothing.
#
# Portable bash 3.2, POSIX sed. See the standing rule in
# docs/technical/decisions/.

set -euo pipefail

MARKER="/// writrun:default"
DEFAULTS_HOME=".writrun/defaults"
PROJECT_HOME="writrun"

usage() { echo "usage: resolve_doc.sh <writrun/path.md> [--origin]" >&2; exit 3; }

ADDR=""
WITH_ORIGIN=""
for arg in "$@"; do
  case "$arg" in
    --origin) WITH_ORIGIN=yes ;;
    -*)       usage ;;
    *)        [ -z "$ADDR" ] || usage
              ADDR="$arg" ;;
  esac
done
[ -n "$ADDR" ] || usage

# The address is the project's home, always. A caller that hands over a
# kit path has confused the two homes, and guessing which it meant would
# hide that from it.
case "$ADDR" in
  "$PROJECT_HOME"/*) ;;
  *) echo "resolve_doc: '$ADDR' is not an address in the project's home ($PROJECT_HOME/)" >&2; exit 3 ;;
esac

REL="${ADDR#"$PROJECT_HOME"/}"
DEFAULT="$DEFAULTS_HOME/$REL"

emit() {
  if [ -n "$WITH_ORIGIN" ]; then printf '%s\t%s\n' "$1" "$2"; else printf '%s\n' "$1"; fi
}

# Present and not deferring: the project's answer, whole. The first
# line is read with any trailing carriage return stripped: a checkout
# under core.autocrlf rewrites every line ending, and a stub must not
# stop deferring because of how it was checked out.
if [ -f "$ADDR" ] && [ "$(sed -n '1p' "$ADDR" | tr -d '\r')" != "$MARKER" ]; then
  emit "$ADDR" declared
  exit 0
fi

# Deferring, or absent. Either way the kit answers — when it has one.
if [ -f "$DEFAULT" ]; then
  emit "$DEFAULT" default
  exit 0
fi

if [ -f "$ADDR" ]; then
  echo "resolve_doc: '$ADDR' defers, but the kit ships no default at '$DEFAULT'" >&2
else
  echo "resolve_doc: neither '$ADDR' nor the kit's default '$DEFAULT' exists" >&2
fi
exit 4
