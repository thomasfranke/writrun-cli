# Coupling to the kit

`writrun` operates a `.writrun/` directory it did not write and cannot
change. Every fact about that directory stated in Go is a fact a WritRun
tag can invalidate, and the binary pins one tag at a time. Three rules
decide which of those facts may exist.

## 1. A kit path enters Go only where the binary calls it

**Called** — the kit's API. Each path is declared once, in the package
that owns the act, and every command references that name.

| Path | Declared in | What calls it |
|---|---|---|
| `.writrun/scripts/**`, `.writrun/skills/**/*.sh` | `internal/kit`, beside the runner that executes them | the commands that run them — the scripts are the execution authority ([0002](../decisions/architecture/0002-scripts-are-the-authority.md)) |
| `.writrun/templates/pull_request_template.md` | `internal/kit` | the body `take` and `author` compose |
| `.writrun/VERSION` | `internal/kittag` | `init` and `update` write it; everything else reads it |
| `.writrun/settings.json` | `internal/kit` | the stage and the conduct flags |
| `.writrun/gates.md` | `internal/kit` | `doctor`'s stage-1 gates |
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
| which human gates a project owes | the rows of `.writrun/gates.md` | a list of gates in Go |
| which files a refresh may write | the fetched template's tree | a list of directories in Go |
| what a spec, task or report carries | `internal/queue`, one reader | a parser per command |

> A WritRun tag that adds a human gate needs no Go change for `doctor`
> to check it.

## 3. `AGENTS.md` is the exception, and it is the only one

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
