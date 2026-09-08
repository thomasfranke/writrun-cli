# Coupling to the kit

`writrun` operates a `.writrun/` directory it did not write and cannot
change, beside a `writrun/` directory that is the project's and that the
binary never writes on the kit's behalf. Every fact about either stated
in Go is a fact a WritRun tag can invalidate, and the binary pins one
tag at a time. Four rules decide which of those facts may exist.

## 1. A kit path enters Go only where the binary calls it

**Called** — the kit's API. Each path is declared once, in the package
that owns the act, and every command references that name.

| Path | Declared in | What calls it |
|---|---|---|
| `.writrun/scripts/**`, `.writrun/skills/**/*.sh` | `internal/kit`, beside the runner that executes them | the commands that run them — the scripts are the execution authority ([0002](../decisions/architecture/0002-scripts-are-the-authority.md)) |
| `.writrun/templates/pull_request_template.md` | `internal/kit` | the body `take` and `author` compose |
| `.writrun/VERSION` | `internal/kittag` | `init` and `update` write it; everything else reads it |
| `writrun/settings.json` | `internal/kit` | the stage, the conduct flags and the commit vocabulary |
| `writrun/gates.md` | `internal/kit` | `doctor`'s stage-1 gates |
| `work/tasks/`, `work/specs/`, `work/reports/` | `internal/queue` | the queue readers |

A path re-declared in a second package is the same defect as a path
listed in Go at all, one level down: the kit still moves, and the edit
is a hunt instead of a line.

**Handled** — copied, refreshed, removed, walked. These do not enter Go.
A file a refresh installs is discovered from the fetched template; a
file `uninstall` deletes is matched by the `writrun-` namespace the kit
puts on its own files outside `.writrun/`.

> A WritRun tag that adds a file needs no Go change for `init` to
> install it, `update` to refresh it, or `uninstall` to remove it.

> A path the kit moves is one edit, in the package that declares it.

## 2. A shape a kit file states is read from that file

The kit ships its own declarations. Where one exists, Go reads it rather
than holding a second copy that a tag can contradict in silence.

| The shape | The file that states it | Not |
|---|---|---|
| which human gates a project owes | the rows of `writrun/gates.md`, or of the default it defers to | a list of gates in Go |
| which files a refresh may write | the fetched template's tree | a list of directories in Go |
| what a spec, task or report carries | `internal/queue`, one reader | a parser per command |

> A WritRun tag that adds a human gate needs no Go change for `doctor`
> to check it.

## 3. Two homes, and a deferring file is resolved, never opened

`.writrun/` is the kit's whole and a refresh replaces every file in it.
`writrun/` is the project's whole and no refresh writes into it. No
folder is part of both.

- A path under `writrun/` is never written by a refresh, and never
  seeded by one: writing the project's home is adoption's act.
- A file under `writrun/` either **defers** — a first line of
  `/// writrun:default`, and there only — or is the project's answer,
  whole. Nothing merges.
- A deferring file is read through the kit's own resolver, which is what
  makes the default reach a project that never customized. Go computes
  no default's address: an address in Go freezes at the tag that wrote
  it, which is the drift the split exists to prevent.
- A value is read from `writrun/settings.json`, never from a
  convention's prose. A convention explains a vocabulary; it never
  carries a second copy for a check to disagree with.

> A WritRun tag that corrects a default reaches every project that
> defers, and needs no Go change to do it.

> A path the adopter owns is named in `internal/kit` and nowhere else,
> so the next tag that moves one is a constant, not a migration.

## 4. `AGENTS.md` is the exception, and it is the only one

`init` grafts WritRun's section into a file the project already owns,
and `uninstall` cuts it back out. No kit file describes that edit,
because the file being edited is the adopter's, not the kit's — so the
binary carries the shape: the heading whose body links
`.writrun/AGENTS.md`, and the `writrun:begin`/`writrun:end` markers kits
before `v0.0.04` left behind.

> `internal/pointer` is the only package that knows the shape of a file
> it does not call.

## What these rules are not

They do not ask the binary to parse the kit's prose. A declaration is
read where the kit ships it as data — a tree of files, a markdown table,
a JSON object. A rule the kit states only in a sentence stays the
scripts' to enforce, and the binary runs the script
([architecture](../architecture.md)).
