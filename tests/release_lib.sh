#!/usr/bin/env bash
# release_lib.sh — the fixture behind the release cases
# (tests/integration/release/). It names the three files the release
# path is made of, and `release_setup` builds the repository the cut
# needs.
#
# `release_setup` leaves a repository on main with a clean tree, ready
# for scripts/release.sh. A local bare `origin` makes the push real, and
# the script's two collaborators are stubbed — `make` (via $MAKE) and
# `gh` (via PATH) — both appending their invocations to $WORK/calls.log.
# The cases that only read the release configuration never call it.
#
# Each case sources this and runs standalone, or under tests/run.sh.

. "$(dirname "${BASH_SOURCE[0]}")/harness.sh"

RELEASE_SH="$REPO_ROOT/scripts/release.sh"
GORELEASER_YAML="$REPO_ROOT/.goreleaser.yaml"
WRITRUN_TAG_SH="$REPO_ROOT/scripts/writrun-tag.sh"
MAIN_GO="$REPO_ROOT/cmd/writrun/main.go"

# yaml_list <key> — the flow list declared for <key>, spaces stripped:
# `goos: [darwin, linux, windows]` reads back as `darwin,linux,windows`.
yaml_list() {
  sed -n "s/^ *$1: *\[\(.*\)\].*/\1/p" "$GORELEASER_YAML" | tr -d ' '
}

# equals <got> <want> — an assertion `check` can run, naming both sides
# when they differ.
equals() {
  [ "$1" = "$2" ] && return 0
  printf 'got "%s", want "%s"\n' "$1" "$2"
  return 1
}

# cd's into the repository.
release_setup() {
  WORK_PREV="$WORK"
  WORK=$(mktemp -d)
  [ -n "$WORK_PREV" ] && rm -rf "$WORK_PREV"
  git init -q "$WORK/repo"
  cd "$WORK/repo" || exit 1
  git symbolic-ref HEAD refs/heads/main
  git config user.email t@example.com
  git config user.name Test
  printf 'baseline\n' > README.md
  git add -A >/dev/null
  git commit -qm baseline
  git init -q --bare "$WORK/origin.git"
  git remote add origin "$WORK/origin.git"
  git push -q origin main
  mkdir -p "$WORK/stub-bin"
  printf '#!/usr/bin/env bash\necho "make $*" >> "%s/calls.log"\n' "$WORK" > "$WORK/stub-bin/make"
  printf '#!/usr/bin/env bash\necho "gh $*" >> "%s/calls.log"\n' "$WORK" > "$WORK/stub-bin/gh"
  chmod +x "$WORK/stub-bin/make" "$WORK/stub-bin/gh"
  export PATH="$WORK/stub-bin:$PATH"
  export MAKE="$WORK/stub-bin/make"
}
