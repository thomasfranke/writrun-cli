---
id: report-0017
status: declined
task_ref: []
doc_ref: null
created: 2026-09-06T01:30:00Z
triaged: 2026-09-06T04:26:48Z
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

**Triage — declined.** The undo does not break the invariant the script
declares. `record_provenance.sh` promises never to rewrite an entry it
*found*; the undo removes the entry the same run appended, before it
reached any commit, and puts the file back with every prior entry in
place.

`check_state.sh` rule I draws that line at the base commit: it compares
what `git show <base>:<path>` holds against the working tree and refuses
only a change to an entry the base held. Driven over the finish fixture
with the ledger on and a base entry already present, it exits 0 on the
undone tree, exits 0 on a tree still carrying the run's entry, exits 1
when a base entry's login is rewritten, and exits 1 when a base entry is
deleted. The undone state is not the state it refuses. The append is
invisible to it in any case — rule I reads the committed diff, and
`finish` commits nothing.

Appending after the confirmation instead was weighed and rejected.
`record_provenance.sh` validates `by=`, the login, the model id and the
four counts, so step 3 gates as well as writes: a bad model id exits 1
today with an empty `gh` log and a clean tree, where after the
confirmation the same refusal would land after `gh pr ready`, which no
undo reaches. `preflight.sh` stage 1 applies the same entry rules, and
only because the entry is on disk when it runs. The decline path's end
state is identical either way.
