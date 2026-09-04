# technical subjects are folders; `architecture.md` alone stays at the root.

**2026-09-04**

Each technical subject becomes a folder — `README.md` as its index,
one file per sub-topic — the same shape `product/` gives its flows and
`decisions/` its subjects. `architecture.md` stays a file at the root:
it is the map of the whole, read before any one subject, and belongs
above them. This supersedes the grain of
[0006](0006-one-file-per-subject.md) — one file per subject becomes
one folder per subject — while its README-as-index rule survives at
both levels. A subject's finer files also tighten the queue-impact
signal: a `doc_ref` names the sub-topic it depends on, not the whole
subject. Rejected: keeping the flat files (a subject splitting later
would migrate paths then; the folder gives it the room now) and
foldering `architecture.md` too (a one-file folder for the map adds a
level and groups nothing).
