# Screens

**What `writrun` renders, drawn.** On any question of layout, keys or
wording these files are the reference the implementation is checked
against. The rules a drawing cannot carry are stated below them.

**A drawing decides where colour goes, never which colour.** Its hues
are a canvas's, chosen to read beside the other frames; the terminal's
are the reader's, and the binary paints with the indexed palette their
theme defines so a value chosen for one background is not imposed on
another ([rules](../rules.md)).

**One folder per section of the entry screen, one file per row it
lists.** The grouping is [`product/README.md`](../README.md)'s, because
one set grouped two ways is two answers about one tool.

| | |
|---|---|
| [Entry](entry.excalidraw) | What the entry screen lists, and how it reaches the rest. |
| [First run](first-run.excalidraw) | What `writrun` opens where `.writrun/` is absent: the environment answered, and `init`. |
| [--help and --version](help.excalidraw) | The two answers that run anywhere, and the grouped help proposed in their place. |

### Tasks

| | |
|---|---|
| [list](tasks/list.excalidraw) | The lister's sections, navigated by keys. One keystroke in from the entry screen. |
| [take](tasks/take.excalidraw) | The composition that waits for a yes, the task arrow-selected, the draft opened. |
| [work](tasks/work.excalidraw) | The task launched on the adopter's agent, and the abort where none is configured. |
| [status](tasks/status.excalidraw) | Where the work stands, on a branch carrying a task and on one carrying none. |
| [finish](tasks/finish.excalidraw) | The deltas checked, the two fields written, preflight, and the draft marked ready. |

### Authoring

| | |
|---|---|
| [author](authoring/author.excalidraw) | The four checks, the composition, and the pull request that opens ready. |
| [amend](authoring/amend.excalidraw) | The spec returned to draft, and the pull request the amendment suspends. |

### Reports

| | |
|---|---|
| [report](reports/report.excalidraw) | The observation recorded: an id minted, no branch, no triage. |

### Adoption

| | |
|---|---|
| [init](adoption/init.excalidraw) | The plan, the stage question, and the gaps the stage's checks name. |
| [update](adoption/update.excalidraw) | What the refresh will write, what it never touches, and the dirty tree it refuses. |
| [doctor](adoption/doctor.excalidraw) | Every requirement named and marked, the stage above the declaration previewed, and the report navigable with a detail pane. |
| [config](adoption/config.excalidraw) | [`config`](../config.md)'s keys, and what judges a change. |
| [uninstall](adoption/uninstall.excalidraw) | What is removed, what stays, and the record that survives the tooling. |

## `writrun` with no command

Opens the queue as a screen navigated by keys; every action dispatches
a command. Which key does what is drawn in
[tasks/list.excalidraw](tasks/list.excalidraw).

- Requires a terminal on stdin and stdout; without one, prints what
  `--help` prints instead.
- Requires an adopted repository; outside one, prints what `--help`
  prints instead. A screen in its place is drawn, proposed, in
  [first-run.excalidraw](first-run.excalidraw).
- Shows the sections [`list`](../queue/list.md) shows, in the same
  order.
- Dispatches [`take`](../pull-requests/take.md),
  [`work`](../queue/work.md) and [`status`](../queue/status.md), and
  nothing else.
- A key runs the command it names — its checks, its questions, its
  confirmation, unchanged ([rules](../rules.md)).
- **A command owns the keyboard alone, and the screen comes back when
  it is done.** A command asks through a terminal program of its own,
  and two of those in one process do not share a keyboard: a screen
  still holding the input is a question that answers itself. So the
  screen is paused and the terminal released — not closed — and reading
  the next thing is a keypress rather than another `writrun`.
- **A command that asks nothing answers into a screen.** Its output is
  captured and shown scrollable, with the screen saying what is running
  while it waits. The screen never gives up the terminal for such a
  command, which is what lets it do either.
- **A command that asks keeps the terminal to itself.** There is no
  capturing a question: it would wait on a reader who cannot see it. It
  takes the whole terminal, as a screen does, and hands it back when the
  reader says they have read it.
- **Which of the two is the command's own declaration**, and the safe
  answer is the second. A command whose declaration is missing keeps the
  terminal, which is only slower to read; the other mistake asks a
  question into the dark.
- The screen offers no action a command does not already provide.
- The screen reads only; every change goes through the dispatched
  command.

## `writrun config`

Opens the settings as a screen: every key the kit documents, under the
section the kit gives it, navigated by keys. Which key does what is
drawn in [adoption/config.excalidraw](adoption/config.excalidraw).

- Holds no list of keys and no list of allowed values. Both are the
  kit's — the keys read out of the settings file the kit's checker
  governs, the values known only to that checker
  ([report-0030](../../../work/reports/report-0030-settings-vocabulary-unreadable.md)).
- Moving selects a key; choosing one asks for its value, as
  [`config`](../config.md) already does from the command line.
- A change is written and then judged: the kit's checker decides, and a
  refusal restores the file to its previous bytes and prints the
  checker's own sentence, unedited.
- The screen offers no key the settings file does not name, and no
  judgement of its own about a value.

## The drawings and the binary

- **A drawing states the design, whole.** A frame is what the screen
  must render — not a record of what it rendered before, and not a
  proposal waiting beside the real thing. One screen, one answer.
- **Where the binary does not match a drawing, the gap is a task**, in
  [`work/`](../../../work/tasks/README.md), never a second frame. The
  drawing does not soften and the binary does not get the benefit of the
  doubt.
- **A change that closes a gap re-transcribes the frame** from a real
  run of the new binary. Its spec names the drawing in *Proposed product
  changes*, so
  [`writrun-check-spec-deltas`](../../../.writrun/skills/writrun-check-spec-deltas/SKILL.md)
  reads the diff and refuses a merge that left the drawing behind.
- **Re-transcribed, never relabelled.** Editing a caption to claim the
  binary prints what the drawing asked for is how a reference stops
  being one.
- **A drawing is not a changelog.** The state before a change is in the
  git history, which is where a reader looks for it.

## Rules for this folder

- A drawing answers how a screen looks; a rule answers what holds,
  including where no screen exists. Neither states the other's part.
- One file per screen, under the folder its entry-screen section
  names. A screen drawn in two files is two answers about one screen.
- **The drawings lead the binary.** They are documentation, and a
  document may state what is asked of the implementation before the
  implementation answers it. A task derives from a drawing; a drawing
  never waits for one.
- A frame whose design the binary already renders is transcribed from a
  real run; one it does not is built from values read out of this
  repository's own files. Neither is written from memory.
- A screen that does not exist yet is drawn like any other: the design
  is the document, and the implementation follows it.
- `.excalidraw` files open at [excalidraw.com](https://excalidraw.com)
  and in the VS Code extension.
