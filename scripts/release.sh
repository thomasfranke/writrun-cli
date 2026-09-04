#!/usr/bin/env bash
# release.sh — cut a release.
#
#   release.sh [patch|minor|major]     default: patch
#
# SemVer: patch bumps the 3rd digit, minor the middle one, major the
# 1st. The next number is computed from the latest tag — the very
# first release is v0.1.0 — then: write the changelog, run the suite,
# and only after that commit, tag, push, and publish the GitHub
# Release with notes generated from the conventional commits.
#
# The tag is the only place the number is written — there is no VERSION
# file to stamp. The cut writes CHANGELOG.md, the one file the release
# commit carries. That file is generated and never edited by hand: one
# writer is what keeps it from becoming a second history that agrees
# with the tags until the first time somebody forgets.
#
# Every guard aborts before anything is mutated. A failed suite aborts
# before the commit, leaving only the changelog dirty in the tree
# (`git checkout CHANGELOG.md` undoes it).
set -euo pipefail

[ "$#" -le 1 ] || { echo "release: pick one of patch|minor|major" >&2; exit 1; }
bump="${1:-patch}"
case "$bump" in
  patch|minor|major) ;;
  *) echo "release: pick one of patch|minor|major" >&2; exit 1 ;;
esac

[ -z "$(git status --porcelain)" ] || { echo "release: working tree not clean" >&2; exit 1; }
[ "$(git branch --show-current)" = "main" ] || { echo "release: tags are cut from main" >&2; exit 1; }

# The cut ends at the forge (`gh release create`), which runs after the
# push — so an unusable gh must abort here, before anything is mutated,
# not fail there, after the tag is already public.
command -v gh >/dev/null 2>&1 && gh auth status >/dev/null 2>&1 \
  || { echo "release: gh must be installed and authenticated — the cut ends at the forge" >&2; exit 1; }

last="$(git tag --list 'v*' --sort=-v:refname | head -n 1)"
if [ -z "$last" ]; then
  next="v0.1.0"
else
  next="$(echo "$last" | awk -F. -v b="$bump" '{
    sub(/^v/, "")
    if (b == "major")      { $1++; $2 = 0; $3 = 0 }
    else if (b == "minor") { $2++; $3 = 0 }
    else                   { $3++ }
    printf "v%d.%d.%d", $1, $2, $3 }')"
fi
echo "release: ${last:-none} -> $next ($bump)"


# The entries are the subjects between the last tag and HEAD, read with
# git and never from the forge: they must exist before the tag does, and
# a network call here would fail differently than the rest of the cut.
range="HEAD"
[ -n "$last" ] && range="$last..HEAD"
types="docs feat fix refactor chore"

changelog_section() {
  printf '## %s — %s\n\n' "$next" "$(date -u +%Y-%m-%d)"

  local subjects rest t matched
  subjects="$(git log --format=%s "$range")"
  if [ -z "$subjects" ]; then
    printf 'No commit landed between %s and this tag.\n\n' "$last"
    return
  fi

  # Grouped in the order conventions/commits.md declares the types, then
  # whatever is left — a subject the vocabulary does not know is listed
  # verbatim rather than dropped. Judging it is the door's job at the
  # pull request, not the tag's.
  rest="$subjects"
  for t in $types; do
    matched="$(printf '%s\n' "$subjects" | grep -E "^${t}(\([^)]*\))?!?: " || true)"
    [ -n "$matched" ] || continue
    printf '### %s\n\n' "$t"
    printf '%s\n' "$matched" | sed 's/^/- /'
    printf '\n'
    rest="$(printf '%s\n' "$rest" | grep -Ev "^${t}(\([^)]*\))?!?: " || true)"
  done
  if [ -n "$rest" ]; then
    printf '### Other\n\n'
    printf '%s\n' "$rest" | sed 's/^/- /'
    printf '\n'
  fi
}

tmp="$(mktemp)"
if [ -f CHANGELOG.md ]; then
  # Newest first: the section goes above every existing one, below
  # whatever preamble the file opens with.
  # No section yet — a file holding only its preamble — is not a failure,
  # and under pipefail an unguarded grep here would abort the cut.
  at="$(grep -n '^## ' CHANGELOG.md | head -n 1 | cut -d: -f1 || true)"
  if [ -n "$at" ]; then
    head -n "$((at - 1))" CHANGELOG.md > "$tmp"
    changelog_section >> "$tmp"
    tail -n "+$at" CHANGELOG.md >> "$tmp"
  else
    cat CHANGELOG.md > "$tmp"
    changelog_section >> "$tmp"
  fi
else
  printf '# Changelog\n\n' > "$tmp"
  changelog_section >> "$tmp"
fi
mv "$tmp" CHANGELOG.md

"${MAKE:-make}" tests

git add CHANGELOG.md
git diff --cached --quiet || git commit -m "chore(release): $next"
git tag -a "$next" -m "$next"
git push origin main --follow-tags
gh release create "$next" --generate-notes
