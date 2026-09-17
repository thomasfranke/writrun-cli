# a command explains itself; `--help` stays one line each. Supersedes [0010](0010-help-is-one-line-per-command.md).

**2026-09-17**

`--help` is still one line per command, now grouped by what a person is
doing. What changed is that `writrun <command> --help` prints that
command's long description — one sentence, then why you would reach for
it, what it does, what it leaves alone, what comes next — and a command
that asks something prints the same before its first question.

[0010](0010-help-is-one-line-per-command.md) rejected exactly this, as
"a fork of the product docs", and the reasoning was sound about the
thing it was looking at: help text written beside `product/` is a second
copy that drifts. What it did not have was a third place for the text to
live. The drawings are that place. Every description is transcribed from
a frame under `product/screens/`, and the suite asserts the rendered
lines against the frame — so the text has one author, and a copy that
drifted would fail the build rather than mislead a reader.

The cost 0010 named is real and is paid differently: the descriptions
are long, and a person who meets them daily would stop reading. So the
four commands that do their work bare print none, and `--help` alone is
unchanged from what 0010 decided.

Rejected again, and for 0010's reasons: generating the text from
`product/` at build time — machinery maintaining prose — and writing it
in Go with nothing holding it to a source.
