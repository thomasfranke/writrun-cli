# Rules for every command

`writrun` is WritRun's optional client. It packages what the methodology
already defines; it never decides in the methodology's place.

## Where a command runs

- Every command except `init` requires an adopted repository — one with
  a `.writrun/` directory. Outside one, the command aborts naming the
  cause and changes nothing.
- `init` is the exception: it runs where the kit is absent, and refuses
  where it is already present.
- `--version` and `--help` always answer, anywhere.

## What no command ever does

- **No command approves.** Nothing flips a spec from `draft` to
  `approved`, and nothing merges a pull request. Both gates stay on the
  forge, operated by a human on purpose.
- **No command reimplements a check.** Every check a command runs is
  the repository's own, in the order the methodology fixed.
- **No command is an agent.** `work` launches one; `writrun` itself
  never reasons about a task's content.
- **No command overwrites the project's own files.** Whatever the
  methodology declares the adopter's — conventions, existing docs, an
  existing `AGENTS.md` — is grafted or left alone, never replaced.

## How a command reports

- A command that would change the repository or the forge shows what it
  will do and asks for confirmation first. `--yes` skips the prompt for
  automated use.
- A failing check stops the command at that check. Nothing later runs,
  nothing partial is left behind, and the failure names the check.
- Exit status is zero only when the command did what it said it would.
- Output is plain text, readable in a terminal without a pager.
