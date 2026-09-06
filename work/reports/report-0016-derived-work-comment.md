---
id: report-0016
status: open
task_ref: []
doc_ref: null
created: 2026-09-06T00:37:16Z
triaged: null
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
