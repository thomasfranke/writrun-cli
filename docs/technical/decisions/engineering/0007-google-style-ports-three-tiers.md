# Google Go style, ports at the boundaries, three test tiers gated in CI.

**2026-09-03**

Unit, integration and e2e; coverage gate at 85% over `internal/`.
Rejected: full Clean Architecture / DDD layering — the domain lives
upstream in WritRun, and a porcelain with entities and repositories of
its own would be a second model free to disagree. The principles stay
(dependency inversion, use cases, small interfaces at the point of
consumption); the ceremony does not.
