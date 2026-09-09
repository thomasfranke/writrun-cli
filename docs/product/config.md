# `writrun config`

Shows the adopter's settings, and changes one.

- Lists every key the settings declare, under the section that owns it,
  with the value the kit's own reader answers.
- **Holds no key and no allowed value.** Which keys exist is the
  settings file, which the kit's checker requires complete; which values
  are legal is the checker's, and it says so in its own words when it
  refuses.
- Changes one key at a time, named as `section.key` — the form the kit's
  reader takes. Without a value, the value is the one free-text
  question; `--yes` does not answer it.
- **Shows the key, the old value and the new one, then asks**
  ([rules](rules.md)).
- Writes, then hands the file to the kit's checker. The change is kept
  only where the checker accepts it.
- **A refused change restores the file byte for byte** and prints the
  checker's own refusal. The settings are never left in a shape the
  project's own kit calls invalid.
- A value that is already what the key holds is reported and not
  written.
- Where the settings file is absent, it says so and offers no edit:
  the kit's reader documents its defaults and keeps working without
  one, and writing the project's home is
  [`init`](adoption/init.md)'s act.
- Changes nothing else. The stage is a declaration, not an
  installation — raising it installs no script, and
  [`update`](adoption/update.md) is what fetches the kit.
