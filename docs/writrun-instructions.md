# docs — the permanent truth

Everything in this folder is **input**: the pipeline reads it to derive
tasks and specs. Write a rule here and work flows from it — never the
other way around: code is checked against these docs, not these docs
against the code.

**The one habit that runs the whole pipeline:** finished writing or
changing a doc here? Tell your AI agent — *"update the tasks"*. That
request is the human gate that lets derivation start: the agent derives
the tasks and specs from what you wrote, shows them to you, and opens
the PR. You write; everything downstream is derived.

What that means in practice:

- **Any file here is fair game for derivation.** A task's `doc_ref`
  points into this folder, and the queue-impact check warns when a
  change here touches docs that non-completed tasks still reference.
- **The shape is the stakeholders'.** WritRun prescribes nothing inside
  `docs/` — organize it however this project's readers need. The
  `product/` and `technical/` folders the kit ships are **optional, a
  suggestion only**: keep them, rename them, or replace them with your
  own structure entirely. One rule survives any shape: product intent
  and technical design never share a file.
- **One file is required: an About.** What this project *is*, who it is
  for, what it refuses to become — short, stable, the first read of
  every audience. The kit deliberately ships no skeleton for it (yours
  may already exist); adoption's minimum bar requires one, under any
  name, stated where a reader would look.
- **Always present tense.** A doc states what the system does; the gap
  between doc and code lives in `work/`, as tasks — never here as
  roadmap prose.
- **Human-gated.** Nothing here merges on agent approval alone, and an
  authored rule derives work only after a human declares it finished.
- **Except this file.** `writrun-instructions.md` is process metadata,
  not project truth: no task ever derives from it, no `doc_ref` points
  at it, and the checks ignore it — editing it owes no declaration and
  breaks no spec's contract.

The queue these docs feed: [`../work/`](../work/README.md).
