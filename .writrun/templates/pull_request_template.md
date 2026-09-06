<!--
Shipped by WritRun — the canonical PR body template. Agents fill it when
opening any pull request; a human opening one by hand copies it from
here (GitHub does not pre-fill from .writrun/, by its own rules — this
project chose one home over the platform's pre-fill).

Everything here is the project's to edit, except one marker: the
"## Derived work" heading — `writrun check` finds the declaration by that
exact heading. Rename it and the check goes blind.

Every task, spec and report named below is a bullet carrying three
things: the id, its title, and a full URL on main that opens the file. A
body is a page, not a file in the tree, so a relative path resolves under
the pull request's own address and reaches nothing.
See docs/product/stage-2-pull-requests/body.md.

Three kinds of PR land here. Keep the section that applies, delete the others.
  Authoring    — writes a rule that is not true yet. No task precedes it;
                 tasks derive from it. Fill "Derived work".
  Implementing — ships an approved spec and closes the loop on the docs it
                 promised. Fill "Spec" and "How to verify".
  Reporting    — records what was observed. Fill "Report", with the task
                 and spec the `tracked` route mints, if it minted any.

See docs/product/stage-1-tasks-and-specs/authoring.md#two-ways-a-permanent-doc-changes.
-->

## What

## Why

<!-- writrun:begin -->

## Derived work

<!-- AUTHORING PRs ONLY. Every task and spec this change creates, the
     spec nested under the task it derives from.
     If the rule derives no work, write "none" and say why in Notes —
     an empty section and a forgotten one look the same. -->

- [task-NNNN](https://github.com/owner/repo/blob/main/work/tasks/task-NNNN-slug.md) — What it asks for
  - [spec-NNNN](https://github.com/owner/repo/blob/main/work/specs/spec-NNNN-slug.md) — What it states

## Spec

<!-- IMPLEMENTATION PRs ONLY. The spec(s) this PR implements. Each must
     already be `approved` — a PR may not approve its own spec. -->

- [spec-NNNN](https://github.com/owner/repo/blob/main/work/specs/spec-NNNN-slug.md) — What it states

## Report

<!-- REPORTING PRs ONLY. The report this change records, and the task and
     spec the `tracked` route minted from it — nested under it, if the
     route minted any. A report that ends fixed, declined or routed
     mints none, and says which end it took. -->

- [report-NNNN](https://github.com/owner/repo/blob/main/work/reports/report-NNNN-slug.md) — What was observed
  - [task-NNNN](https://github.com/owner/repo/blob/main/work/tasks/task-NNNN-slug.md) — What it asks for

## How to verify

<!-- The methodology's answer. Implementation PRs: the
     writrun-check-spec-deltas result. Authoring PRs: this check does not
     apply — an authoring change has no spec to check against. Either
     way, name anything a reviewer should re-read by hand. -->

## How to test

<!-- The reviewer's answer, and a different question from the one above:
     what to run to watch the change work, and what to expect back.
     Commands, not assurances. A change that ships nothing runnable says
     so in a line — a deleted section and a forgotten one look the same. -->

<!-- writrun:end -->

## Notes
