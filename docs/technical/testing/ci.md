# CI

- CI (`tests.yml`) runs on pull requests, in order, stopping at the
  first failure: `gofmt -l` clean → `go vet` → unit with the coverage
  gate → integration → e2e. `release-readiness.yml` re-verifies `main`
  on every push.
