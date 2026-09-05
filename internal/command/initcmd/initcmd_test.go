package initcmd

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/thomasfranke/writrun-cli/internal/command"

	"github.com/thomasfranke/writrun-cli/internal/gitx"

	"github.com/thomasfranke/writrun-cli/internal/vfs"
)

// runInit drives the whole command against a target repository, the
// terminal faked, and returns what was printed and the error.
func runInit(t *testing.T, target string, d Deps, term *command.FakeTerminal, yes bool, args ...string) (string, error) {
	t.Helper()
	if d.Git == nil {
		d.Git = gitx.Run
	}
	if d.LookPath == nil {
		d.LookPath = func(name string) (string, error) { return "/usr/bin/" + name, nil }
	}
	if d.Gh == nil {
		d.Gh = func(args ...string) (string, error) { return "", nil }
	}
	if d.Files == nil {
		d.Files = vfs.OS{}
	}
	var out bytes.Buffer
	ctx := &command.Ctx{
		Stdout:   &out,
		Stderr:   &out,
		Terminal: term,
		Root:     target,
		Yes:      yes,
	}
	cmd := New(d)
	err := cmd.Run(ctx, args)
	return out.String(), err
}

func TestInitAdoptsEndToEnd(t *testing.T) {
	target := makeTarget(t)
	d := Deps{Tag: testTag, Source: makeSource(t)}
	term := &command.FakeTerminal{}
	out, err := runInit(t, target, d, term, true, "--stage", "1")
	if err != nil {
		t.Fatalf("init = %v\n%s", err, out)
	}
	if got := strings.TrimSpace(read(t, target, ".writrun/VERSION")); got != testTag {
		t.Errorf("VERSION = %q, want %q", got, testTag)
	}
	if !strings.Contains(out, "Adopted WritRun "+testTag) {
		t.Errorf("no adoption report:\n%s", out)
	}
	if len(term.Asked) != 0 {
		t.Errorf("questions were asked under --stage and --yes: %v", term.Asked)
	}
}

func TestInitDeclineLeavesTheRepositoryUntouched(t *testing.T) {
	target := makeTarget(t)
	d := Deps{Tag: testTag, Source: makeSource(t)}
	term := &command.FakeTerminal{In: true, ConfirmAnswer: false}
	out, err := runInit(t, target, d, term, false, "--stage", "1")
	if !errors.Is(err, command.ErrDeclined) {
		t.Fatalf("init = %v, want ErrDeclined\n%s", err, out)
	}
	if _, statErr := os.Stat(filepath.Join(target, ".writrun")); statErr == nil {
		t.Error(".writrun/ exists after a decline")
	}
	if _, statErr := os.Stat(filepath.Join(target, ".git", "hooks", "commit-msg")); statErr == nil {
		t.Error("the hook was installed before the confirmation")
	}
}

func TestInitStageIsArrowSelectedWithoutTheFlag(t *testing.T) {
	target := makeTarget(t)
	d := Deps{Tag: testTag, Source: makeSource(t)}
	term := &command.FakeTerminal{In: true, SelectIndex: 1, ConfirmAnswer: true}
	out, err := runInit(t, target, d, term, false)
	if err != nil {
		t.Fatalf("init = %v\n%s", err, out)
	}
	if !strings.Contains(read(t, target, ".writrun/settings.json"), `"stage": 2`) {
		t.Error("the selected stage did not land in the settings")
	}
	if len(term.Asked) == 0 || !strings.Contains(term.Asked[0], "stage") {
		t.Errorf("the stage question was not asked: %v", term.Asked)
	}
}

func TestInitWithoutTerminalNamesTheAnsweringFlag(t *testing.T) {
	target := makeTarget(t)
	d := Deps{Tag: testTag, Source: makeSource(t)}
	_, err := runInit(t, target, d, &command.FakeTerminal{}, true)
	if err == nil || !strings.Contains(err.Error(), "--stage") {
		t.Errorf("init = %v, want an abort naming --stage", err)
	}
}

func TestInitRefusesADirtyTree(t *testing.T) {
	target := makeTarget(t)
	write(t, target, "uncommitted.txt", "dirt\n")
	d := Deps{Tag: testTag, Source: makeSource(t)}
	_, err := runInit(t, target, d, &command.FakeTerminal{}, true, "--stage", "1")
	if err == nil || !strings.Contains(err.Error(), "dirty") {
		t.Errorf("init = %v, want the dirty-tree refusal", err)
	}
	if _, statErr := os.Stat(filepath.Join(target, ".writrun")); statErr == nil {
		t.Error(".writrun/ exists after a refusal")
	}
}

func TestInitRefusesAForeignHook(t *testing.T) {
	target := makeTarget(t)
	write(t, target, ".git/hooks/commit-msg", "#!/bin/sh\nexit 0\n")
	d := Deps{Tag: testTag, Source: makeSource(t)}
	_, err := runInit(t, target, d, &command.FakeTerminal{}, true, "--stage", "1")
	if err == nil || !strings.Contains(err.Error(), "already installed") {
		t.Errorf("init = %v, want the foreign-hook refusal", err)
	}
}

func TestInitAbortsBeforeAnyWriteWithoutTheSource(t *testing.T) {
	target := makeTarget(t)
	d := Deps{Tag: testTag, Source: filepath.Join(t.TempDir(), "nowhere")}
	_, err := runInit(t, target, d, &command.FakeTerminal{}, true, "--stage", "1")
	if err == nil || !strings.Contains(err.Error(), "nothing was written") {
		t.Errorf("init = %v, want the fetch failure naming itself", err)
	}
	if _, statErr := os.Stat(filepath.Join(target, ".writrun")); statErr == nil {
		t.Error(".writrun/ exists after a failed fetch")
	}
}

func TestInitRefusesASourceWithoutATemplate(t *testing.T) {
	target := makeTarget(t)
	src := t.TempDir()
	gitT(t, src, "init", "-q")
	write(t, src, "README.md", "not a kit\n")
	gitT(t, src, "add", ".")
	gitT(t, src, "commit", "-q", "-m", "x")
	gitT(t, src, "tag", testTag)
	d := Deps{Tag: testTag, Source: src}
	_, err := runInit(t, target, d, &command.FakeTerminal{}, true, "--stage", "1")
	if err == nil || !strings.Contains(err.Error(), "no template/") {
		t.Errorf("init = %v, want the no-template refusal", err)
	}
}

func TestInitRefusesABadStageValue(t *testing.T) {
	target := makeTarget(t)
	d := Deps{Tag: testTag, Source: makeSource(t)}
	for _, bad := range []string{"0", "4", "two"} {
		_, err := runInit(t, target, d, &command.FakeTerminal{}, true, "--stage", bad)
		if err == nil || !strings.Contains(err.Error(), "--stage must be 1, 2 or 3") {
			t.Errorf("--stage %s: err = %v, want the range refusal", bad, err)
		}
	}
}

func TestInitRefusesUnexpectedArguments(t *testing.T) {
	target := makeTarget(t)
	d := Deps{Tag: testTag, Source: makeSource(t)}
	_, err := runInit(t, target, d, &command.FakeTerminal{}, true, "bogus")
	if err == nil || !strings.Contains(err.Error(), "unexpected argument") {
		t.Errorf("init = %v, want the argument refusal", err)
	}
}

func TestInitNamesGapsAndStillCompletes(t *testing.T) {
	// The target has no About file and no real chapters, so stage 1
	// finds gaps — named, never blocking (spec-0002).
	target := makeTarget(t)
	d := Deps{Tag: testTag, Source: makeSource(t)}
	out, err := runInit(t, target, d, &command.FakeTerminal{}, true, "--stage", "1")
	if err != nil {
		t.Fatalf("init = %v, want completion despite gaps\n%s", err, out)
	}
	if !strings.Contains(out, "docs/about.md") {
		t.Errorf("the About gap is not named:\n%s", out)
	}
	if !strings.Contains(out, "named, not fixed") {
		t.Errorf("the gaps are not reported as unfixed:\n%s", out)
	}
}

func TestInitSaysShippedDefaultsWhenNothingToExtract(t *testing.T) {
	target := makeTarget(t, "just words", "more words")
	d := Deps{Tag: testTag, Source: makeSource(t)}
	out, err := runInit(t, target, d, &command.FakeTerminal{}, true, "--stage", "1")
	if err != nil {
		t.Fatalf("init = %v\n%s", err, out)
	}
	if !strings.Contains(out, "shipped defaults") {
		t.Errorf("the plan does not say the defaults stand:\n%s", out)
	}
	if !strings.Contains(read(t, target, ".writrun/scripts/stage-2-pull-requests/check_observance.sh"), `TYPES="docs feat fix refactor chore"`) {
		t.Error("the shipped vocabulary did not survive")
	}
}

func TestInitRefusesATemplateWithoutAgents(t *testing.T) {
	src := t.TempDir()
	gitT(t, src, "init", "-q")
	write(t, src, "template/WRITRUN.md", "# a kit missing its skeleton\n")
	gitT(t, src, "add", ".")
	gitT(t, src, "commit", "-q", "-m", "x")
	gitT(t, src, "tag", testTag)

	target := makeTarget(t)
	d := Deps{Tag: testTag, Source: src}
	_, err := runInit(t, target, d, &command.FakeTerminal{}, true, "--stage", "1")
	if err == nil || !strings.Contains(err.Error(), "no template/AGENTS.md") {
		t.Fatalf("init = %v, want the missing-skeleton refusal", err)
	}
	if _, statErr := os.Stat(filepath.Join(target, ".writrun")); statErr == nil {
		t.Error(".writrun/ exists after a refusal")
	}
}

func TestInitDirtyTreeRefusalNamesTheRemedyThatClearsIt(t *testing.T) {
	// git status counts untracked files, and a plain `git stash` leaves
	// them exactly where they are.
	target := makeTarget(t)
	write(t, target, "untracked.txt", "dirt\n")
	d := Deps{Tag: testTag, Source: makeSource(t)}
	_, err := runInit(t, target, d, &command.FakeTerminal{}, true, "--stage", "1")
	if err == nil || !strings.Contains(err.Error(), "git stash -u") {
		t.Errorf("init = %v, want the refusal naming `git stash -u`", err)
	}
}
