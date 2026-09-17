---
id: report-0046
status: open
task_ref: []
doc_ref: product/pull-requests/finish.md
created: 2026-09-17T18:40:33Z
triaged: null
---

# preflight refuses the only range shape that reaches the working tree

**References:** [product/pull-requests/finish.md](../../docs/product/pull-requests/finish.md)

The kit reads a **bare** range — `origin/main`, with no `..` — as
"this ref against the working tree". `ql_range_ends` leaves its head
end empty for that shape and for no other, and
`check_deltas.sh`'s header describes the behaviour deliberately: "The
bare shape hands back an empty HEADREF: its diff compares the ref to
the *working tree*."

`preflight.sh` cannot be given one. It decides what its arguments are
by shape, at `:44-52`:

    *..*)  RANGE="$arg" ;;
    *)     IDS_ARG="$arg" ;;

So `origin/main` arrives as a second task list and the run dies with
`PREFLIGHT: two task lists given ('task-0001' and 'origin/main')`.
Observed while implementing spec-0045: `writrun finish` handed
preflight the bare range, preflight refused it, and the command's own
undo put both completion edits back — correct behaviour in front of a
gate that would not answer.

No other shape reaches the tree. `origin/main..` and `origin/main...`
both match `*..*` and are accepted as ranges, and both resolve their
head end to `HEAD`: `ql_range_ends` reads `${QL_HEADREF:-HEAD}`
for the two-dot form and `${right:-HEAD}` for the three-dot one. The
one shape that reaches the working tree is the one shape preflight
will not take as a range.

The consequence is [report-0014](report-0014-uncommitted-completion-edits.md)'s,
unreached. `finish` writes the spec's `implemented` and the task's
`completed` date into the working tree and commits nothing, and the two
stages preflight runs after stage 1 read a commit range — so they still
judge the branch as it was before the command ran, and still answer
`deltas checked: none — no spec reached 'implemented' in this range`.
spec-0045 reached step 1, which calls `check_deltas.sh` directly and
does read the tree; step 4 is behind this.

A patch here would not survive: `preflight.sh` is the kit's and the
next `writrun update` replaces it.
