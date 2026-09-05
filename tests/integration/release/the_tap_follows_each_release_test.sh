#!/usr/bin/env bash
. "$(dirname "$0")/../../release_lib.sh"

# `brew install thomasfranke/tap/writrun` resolves only when GoReleaser
# writes into thomasfranke/homebrew-tap under the name `writrun`
# (technical/distribution/routes.md).
#
# A cask, not a formula: GoReleaser deprecated formula generation, and
# `goreleaser check` fails a configuration that still asks for one.

check "the tap repository is thomasfranke/homebrew-tap" 0 "" \
  -- grep -qE '^ +owner: thomasfranke$' "$GORELEASER_YAML"
check "the tap is the homebrew-tap repository" 0 "" \
  -- grep -qE '^ +name: homebrew-tap$' "$GORELEASER_YAML"
check "the installed name is writrun" 0 "" \
  -- grep -qE '^ +- name: writrun$' "$GORELEASER_YAML"
check "the tap is written by GoReleaser, not by hand" 0 "" \
  -- grep -qE '^homebrew_casks:$' "$GORELEASER_YAML"

# The tap lives in another repository, which GITHUB_TOKEN cannot reach:
# the push needs its own credential, and the workflow passes it.
check "the tap push has a credential of its own" 0 "" \
  -- grep -qF 'HOMEBREW_TAP_TOKEN' "$GORELEASER_YAML"
check "the release workflow supplies it" 0 "" \
  -- grep -qF 'HOMEBREW_TAP_TOKEN' "$REPO_ROOT/.github/workflows/release.yml"

finish
