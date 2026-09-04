# Thin aliases only — the scripts are the interface, this file just names
# them. CI calls the scripts directly and must keep working without make.

.PHONY: tests test test-unit test-integration

# Everything — both tiers.
tests:
	bash tests/run.sh

# `make test` is the muscle-memory alias for the same thing.
test: tests

test-unit:
	@fail=0; \
	for f in tests/unit/*/*_test.sh; do \
	  [ -e "$$f" ] || continue; \
	  bash "$$f" || fail=1; \
	done; \
	exit $$fail

test-integration:
	@fail=0; \
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
	@fail=0; found=0; \
	for f in tests/$*/*_test.sh tests/$*/*/*_test.sh tests/*/$*/*_test.sh; do \
	  [ -e "$$f" ] || continue; found=1; \
	  bash "$$f" || fail=1; \
	done; \
	if [ "$$found" -eq 0 ]; then echo "no such suite: $*"; exit 3; fi; \
	exit $$fail
