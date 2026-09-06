---
id: report-0016
status: declined
task_ref: []
doc_ref: null
created: 2026-09-06T00:37:16Z
triaged: 2026-09-06T04:26:48Z
---

# The template's Derived-work comment reads as a none declaration

`check_derived_work.sh` reads the lines between `## Derived work` and
the next `## ` heading, and passes when it greps the word `none` there
(`grep -qiE '(^|[^[:alnum:]])none([^[:alnum:]]|$)'`). The instruction
comment `pull_request_template.md` ships inside that section contains
`write "none" and say why in Notes`, which matches. Reproduced against
`.writrun/` at VERSION v0.0.03: over a range adding a permanent doc and
no task, with `PR_BODY` set to the template's Derived-work section
verbatim, the check prints `Derived work explicitly declared as none.`
and exits 0. `writrun author` drops that comment when it fills the
section, so the bodies this binary opens are unaffected; a body a human
copies from the template is not.

**Triage — declined.** Not this repository's. The finding judges two kit
artefacts agreeing badly with each other: the shipped template's
instruction comment and `check_derived_work.sh`'s grep. The binary
appears only as the thing unaffected.

`.writrun/templates` is in `internal/kitpaths`'s `RefreshDirs` — "the
kit-owned directories `writrun update` replaces whole" — and the file
has one commit here, the adoption. A fix written here is overwritten by
design.

Reproduced before declining: over a range adding a permanent doc and no
task, a body carrying the template section verbatim passes
`check_derived_work.sh` with "Derived work explicitly declared as none";
deleting only the comment from that same body makes it exit 1. Upstream
still carries both halves unchanged.

Every body this binary opens is unaffected, checked per command:
`author` drops the comment with the section, `take` delegates to
`take_task.sh` which drops the heading whole, `amend` strips every
comment but the `writrun:` markers, and `finish` composes no body. What
reaches the forge carrying it is a body a human copied out of the
template by hand.

Closed without opening a report upstream, by Thomas's decision.
