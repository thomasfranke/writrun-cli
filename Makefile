# Thin aliases only — the scripts are the interface, this file just names
# them. CI calls these targets too (tests.yml, release-readiness.yml), so
# renaming one is a workflow change.

.PHONY: tests test test-unit test-integration cover

# Everything — every tier: Go unit tests, then the bash suites. The
# coverage gate is not here: it lives in CI (technical/testing/tiers.md).
# `make cover` is how a session reads the percentage before pushing.
tests:
	go test ./...
	bash tests/run.sh

# `make test` is the muscle-memory alias for the same thing.
test: tests

# unit is Go, table-driven, beside the code (technical/testing/tiers.md).
test-unit:
	go test ./...

# The percentage and the floor, printed — the same script CI runs, so a
# session can read before pushing what the pipeline will read after.
# `make cover 90` raises the floor for one run; the pipeline's floor is
# the script's default.
cover:
	@bash scripts/coverage.sh $(filter-out cover,$(MAKECMDGOALS))

# WRITRUN_BIN_DIR lets the CLI cases share one compiled binary
# (tests/cli_lib.sh) instead of relinking per case file.
test-integration:
	@fail=0; WRITRUN_BIN_DIR=$$(mktemp -d); export WRITRUN_BIN_DIR; \
	trap 'rm -rf "$$WRITRUN_BIN_DIR"' EXIT; \
	for f in tests/integration/*/*_test.sh; do \
	  [ -e "$$f" ] || continue; \
	  bash "$$f" || fail=1; \
	done; \
	exit $$fail

# Cut a release: `make release` (= patch), or `make release patch|minor|major`.
# The whole path — compute, changelog, test, commit, tag, push,
# publish — lives in the script.
.PHONY: release patch minor major
release:
	@MAKE="$(MAKE)" bash scripts/release.sh $(filter major minor patch,$(MAKECMDGOALS))

# The bump words are goals only so `make release minor` parses — no-ops alone.
patch minor major:
	@:

# make test-unit / test-integration (a tier), or test-release, ... (one
# suite directory, whichever tier it lives in).
test-%:
	@fail=0; found=0; WRITRUN_BIN_DIR=$$(mktemp -d); export WRITRUN_BIN_DIR; \
	trap 'rm -rf "$$WRITRUN_BIN_DIR"' EXIT; \
	for f in tests/$*/*_test.sh tests/$*/*/*_test.sh tests/*/$*/*_test.sh; do \
	  [ -e "$$f" ] || continue; found=1; \
	  bash "$$f" || fail=1; \
	done; \
	if [ "$$found" -eq 0 ]; then echo "no such suite: $*"; exit 3; fi; \
	exit $$fail
