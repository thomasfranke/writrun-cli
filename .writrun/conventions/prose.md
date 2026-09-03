# Prose

How this project writes documentation, skills and comments. Taste, not
machinery — no script reads this file.

## One claim per sentence

Split a sentence carrying two claims. The reader who needs the second
one should not have to parse the first to reach it.

## Short over complete

A sentence the reader re-reads has cost more than the qualifier saved.
Drop the qualifier. Where the exception matters, it earns its own
sentence.

## Say it once

A rule stands in one place — the doc that owns it. Everywhere else links
there. Two copies are not redundancy; they are two rules that agree
today. See [`concepts/technical-doc.md`](https://github.com/thomasfranke/writrun/blob/main/docs/product/concepts/technical-doc.md#the-link-dont-restate-rule)
for the split this protects, and
[`concepts/skill.md`](https://github.com/thomasfranke/writrun/blob/main/docs/product/concepts/skill.md) for why a
skill is held tightest.

## Lead with the claim

State the rule, then the reasoning. A paragraph that builds to its point
makes every reader who already agreed read all of it.

## No throat-clearing

Cut "it is important to note", "in order to", "as mentioned above". They
carry nothing and they are read every time.

## Prose explains; settings decide

Where a value is machine-readable it lives in
[`settings.json`](../settings.json), and prose never restates it. A
value written twice will disagree with itself.
