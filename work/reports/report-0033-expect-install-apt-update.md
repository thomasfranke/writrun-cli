---
id: report-0033
status: fixed
task_ref: []
doc_ref: null
created: 2026-09-09T17:42:37Z
triaged: 2026-09-09T17:42:51Z
---

# Installing expect coupled the pipeline to an unrelated apt repository

The `expect` step I added to two workflows ran `apt-get update`
before installing. On the release commit it failed:

    E: Failed to fetch https://dl.google.com/linux/chrome-stable/.../Packages.gz
       Hash Sum mismatch
    E: Some index files failed to download.

That is a third-party repository the GitHub runner image ships,
unrelated to `expect`. `apt-get update` exits non-zero when any index
fails, the step runs under `bash -e`, and main went red — on the commit
that cut v0.0.1, so the repository's first release is tagged from a
red verdict.

The `release` workflow itself succeeded and the release published; only
the readiness verdict failed, and it failed on infrastructure rather
than on anything the suite checks.

The update was mine and defensive: I added it in case the runner's
package lists were stale. The runner already carries `expect`, so it
buys nothing and couples the step to every repository the image
happens to configure.

Fixed by installing first and only updating on failure, tolerating
update's own failure there — whether `expect` installs is not the
Chrome repository's business.
