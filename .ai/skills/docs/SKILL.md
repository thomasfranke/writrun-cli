---
name: docs
description: Load before writing or editing any markdown this project owns — docs/, README.md, AGENTS.md, CONTRIBUTING.md, work/ bodies, skills, code comments. Defines the mandatory style — CLI only, no prose, objective, cohesive — and the checks a finished text must pass.
---

# Documentation

Four laws. Every text this project owns passes all four before it is
handed over.

## 0. CLI only

- The docs describe the `writrun` CLI — what the binary does and how it
  is built. Nothing else.
- The methodology's rules, concepts and flows live upstream in
  [WritRun](https://github.com/thomasfranke/writrun): link there, never
  restate or document them here.
- This repository's own working flow lives in `AGENTS.md`, never in
  `docs/`.
- A report in `work/reports/` records a finding about this binary. A
  finding whose subject is a kit script or a methodology rule is not
  filed here at all — it goes to Thomas, who decides whether a report
  opens upstream (`AGENTS.md`, human gates).
- The subject is what a finding judges, not the file it names: `finish`
  reversing a ledger entry is this binary's, and `record_provenance.sh`
  contradicting itself is not.

## 1. No prose

- State rules, never stories. Delete narrative, metaphor, scene-setting,
  and rhetorical questions.
- One sentence, one claim. A sentence needing two dashes or a second
  comma clause is two sentences.
- Lead with the claim; reasoning follows it, never builds to it.
- No throat-clearing: "it is important to note", "in order to", "as
  mentioned above" carry nothing.
- A rationale earns one clause on the rule it explains. Anything longer
  is a dated entry in `docs/technical/decisions/`, not doc body.

## 2. Objective

- Every statement is checkable: a reader answers "does the repo comply —
  yes or no" without interpretation.
- Exact over vague: name the path, the flag, the exit code, the number.
  "some", "various", "etc." are banned.
- Present tense, active voice. A doc states what the system does; what
  it will do is a task in `work/`, never doc text.
- Enumerations are tables. Rules are bullets. Paragraphs only where one
  claim depends on another.

## 3. Cohesive

- One term per concept, project-wide. The first doc to name a concept
  fixes the term; every other file uses that term or links to it.
- A fact lives in exactly one file. Everywhere else links to it —
  restating is forking.
- Machine-readable values live in their file (`settings.json`, `VERSION`,
  scripts); prose points, never restates.
- Headings are the contract: a reader navigates by scanning them, so a
  heading names its content exactly.

## Hand-over check

Before handing a text over, verify:

- [ ] no statement documents the methodology or this repo's flow — CLI only
- [ ] no sentence carries two claims
- [ ] no unverifiable statement
- [ ] no fact restated from another file
- [ ] no future tense about the system
- [ ] every enumeration of 3+ items is a table or list, not a paragraph

## Scope

Everything the project writes: `docs/`, `README.md`, `AGENTS.md`,
`CONTRIBUTING.md`, task/spec/report bodies in `work/`, skills, commit
messages' bodies, code comments. Files owned by WritRun's kit
(`.writrun/` except `conventions/` and `settings.json`, and
`docs/writrun-instructions.md`) are upstream's; this skill does not
reach them.
