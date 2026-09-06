---
id: report-0017
status: open
task_ref: []
doc_ref: null
created: 2026-09-06T01:30:00Z
triaged: null
---

# record_provenance.sh declares itself append-only, and finish's undo now reverses an entry it appended

`record_provenance.sh`'s own header says it "only ever *appends*. It
never rewrites an entry it found, and `writrun-check-task-state` refuses
a diff that does" — the shape of that permission is what keeps the one
machine field a branch writes from becoming "a branch may edit front
matter". `writrun finish` runs that script at step 3, between the
completion writes and `preflight.sh`, and spec-0017's undo now puts the
task file back as the command found it whenever the run ends in anything
but a success: the appended entry goes with the completion date, from
outside the script, by a writer the script's contract does not name. The
behaviour observed before the undo reached that file was the other half
of the same corner — on a task the worker had already dated by hand,
step 2 wrote nothing there, so the entry survived a decline alone and
`git status --porcelain` showed a lone provenance line under the words
"declined — nothing changed". Both readings are defensible from the
texts at hand: the entry records tokens that were really spent, and it
records them against an act that did not happen. No document says which
writer wins when a command that must leave the tree clean meets a script
that only ever appends.
