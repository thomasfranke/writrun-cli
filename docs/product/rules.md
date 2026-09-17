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
- `writrun` with no command opens a [screen](screens/README.md) where
  stdin and stdout are terminals, and prints what `--help` prints where
  either is not.
- Which screen is the kit's answer: the entry screen inside an
  adoption, and outside one the
  [first run](screens/README.md#writrun-where-the-kit-is-absent), which
  offers `init` alone and holds it out of reach while an environment
  requirement is unmet.
- A screen only dispatches the commands these rules govern.

## What no command ever does

- **No command approves.** Nothing flips a spec from `draft` to
  `approved`, and nothing merges a pull request. Both gates stay on the
  forge, operated by a human on purpose.
- **No command reimplements a check.** Every check a command runs is
  the repository's own, in the order the methodology fixed.
- **No command is an agent.** `work` launches one; `writrun` itself
  never reasons about a task's content.
- **No command overwrites the project's own files with the kit's.**
  Whatever the methodology declares the adopter's — everything under
  `writrun/`, existing docs, an existing `AGENTS.md` — is grafted or
  left alone, never replaced by a copy the kit ships.
  [`config`](config.md) is not an exception to this and never was: it
  writes the adopter's own answer at the adopter's asking, and hands
  the result to the kit's checker before keeping it.

## How a command reports

- A command that would change the repository or the forge shows what it
  will do and asks for confirmation first. `--yes` skips the prompt for
  automated use.
- Where stdin is a terminal, a question is navigated, not typed: arrow
  keys move, Enter confirms — a stage, a task, a yes. Typing is
  required only for free text, such as a title.
- Every question has a flag that answers it, so automation never meets
  a prompt. Without a terminal, an unanswered question aborts instead
  of hanging.
- A failing check stops the command at that check. Nothing later runs,
  nothing partial is left behind, and the failure names the check.
- A failure after the first write — a branch pushed, a file created —
  names the exact command that resumes the flow.
- Exit status is zero only when the command did what it said it would.
- Output is plain text, readable in a terminal without a pager.
- Color appears only where stdout is a terminal; `NO_COLOR` set or
  `--no-color` given disables it.
- `--version` names the product — `writrun-cli` — its own version, and
  the WritRun tag it pins. The command a person types stays `writrun`.
- `--help` prints one line per command, grouped by what a person is
  doing, and where the docs live. Each row's text is the command's own
  summary, and the entry screen shows that same string.
- `writrun <command> --help` prints that command's long description:
  one sentence naming it, then why you would reach for it, what it
  does, what it leaves alone, and what comes next.
- A command that asks something prints its long description before its
  first question, where it was run with no arguments and stdin is a
  terminal. `list`, `status`, `doctor` and `work` print none: they do
  their work bare, and a daily explanation is noise.
