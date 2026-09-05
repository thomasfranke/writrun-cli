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
)

func readOnly(t *testing.T, path string, mode os.FileMode) {
	t.Helper()
	if err := os.Chmod(path, mode); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(path, 0o755) })
}

func skipAsRoot(t *testing.T) {
	t.Helper()
	if os.Geteuid() == 0 {
		t.Skip("root writes where nobody else may")
	}
}

func TestNewDefaultsTheSourceToTheCanonicalRepository(t *testing.T) {
	// An empty Source is the canonical repository, resolved once at
	// wiring time; the fetch then fails on the network, not on an
	// empty URL.
	target := makeTarget(t)
	err, _ := runInit(t, target, Deps{Tag: testTag}, &command.FakeTerminal{}, true, "--stage", "1")
	if err == nil {
		t.Skip("this machine reached the canonical repository")
	}
	if !strings.Contains(err.Error(), "github.com/thomasfranke/writrun") {
		t.Errorf("the default source was not used: %v", err)
	}
}

func TestAnUnknownFlagIsRefused(t *testing.T) {
	target := makeTarget(t)
	err, _ := runInit(t, target, Deps{Tag: testTag, Source: makeSource(t)}, &command.FakeTerminal{}, true, "--nope")
	if err == nil {
		t.Fatal("an unknown flag was accepted")
	}
}

func TestAWorkingTreeGitCannotReadStopsTheAdoption(t *testing.T) {
	// Outside a repository there is no tree to read, and an adoption
	// may not proceed without knowing whether one is dirty.
	target := t.TempDir()
	err, _ := runInit(t, target, Deps{Tag: testTag, Source: makeSource(t)}, &command.FakeTerminal{}, true, "--stage", "1")
	if err == nil {
		t.Fatal("the adoption proceeded outside a repository")
	}
	if !strings.Contains(err.Error(), "working tree") && !strings.Contains(err.Error(), "hooks directory") {
		t.Errorf("the error names neither read: %v", err)
	}
}

func TestAnAdoptionThatCouldNotWriteSaysHowToUndoIt(t *testing.T) {
	skipAsRoot(t)
	target := makeTarget(t)
	readOnly(t, target, 0o555)

	err, _ := runInit(t, target, Deps{Tag: testTag, Source: makeSource(t)}, &command.FakeTerminal{}, true, "--stage", "1")
	if err == nil {
		t.Fatal("an adoption that could not write succeeded")
	}
	if !strings.Contains(err.Error(), "the adoption is partial") {
		t.Errorf("the error does not say what state the tree is in: %v", err)
	}
	if !strings.Contains(err.Error(), "git clean -fd") {
		t.Errorf("the error does not say how to undo it: %v", err)
	}
}

func TestATemplateThatCannotBeWalkedIsReported(t *testing.T) {
	skipAsRoot(t)
	src := makeSource(t)
	clone := t.TempDir()
	gitT(t, "", "clone", "-q", "--depth", "1", "--branch", testTag, src, filepath.Join(clone, "writrun"))
	template := filepath.Join(clone, "writrun", "template")
	readOnly(t, filepath.Join(template, ".writrun"), 0o000)

	_, err := plan(makeTarget(t), template, testTag, src, 1, filepath.Join(t.TempDir(), "commit-msg"), gitx.Run)
	if err == nil {
		t.Fatal("a template that cannot be walked was planned over")
	}
	if !strings.Contains(err.Error(), "reading the template") {
		t.Errorf("the error does not name the act: %v", err)
	}
}

func TestATemplateWithNoAgentsIsNotAWritRunRepository(t *testing.T) {
	src := makeSource(t)
	clone := t.TempDir()
	gitT(t, "", "clone", "-q", "--depth", "1", "--branch", testTag, src, filepath.Join(clone, "writrun"))
	template := filepath.Join(clone, "writrun", "template")
	if err := os.Remove(filepath.Join(template, "AGENTS.md")); err != nil {
		t.Fatal(err)
	}
	if _, err := plan(makeTarget(t), template, testTag, src, 1, filepath.Join(t.TempDir(), "commit-msg"), gitx.Run); err == nil {
		t.Fatal("a template with no AGENTS.md was planned over")
	}
}

func TestApplyReportsEachWriteItCouldNotMake(t *testing.T) {
	skipAsRoot(t)
	src := makeSource(t)
	clone := t.TempDir()
	gitT(t, "", "clone", "-q", "--depth", "1", "--branch", testTag, src, filepath.Join(clone, "writrun"))
	template := filepath.Join(clone, "writrun", "template")

	// A copy that cannot land.
	target := makeTarget(t)
	a, err := plan(target, template, testTag, src, 1, filepath.Join(t.TempDir(), "commit-msg"), gitx.Run)
	if err != nil {
		t.Fatal(err)
	}
	readOnly(t, target, 0o555)
	if applyErr := a.apply(); applyErr == nil {
		t.Fatal("a copy under an unwritable root succeeded")
	} else if !strings.Contains(applyErr.Error(), "copying") {
		t.Errorf("the error does not name the act: %v", applyErr)
	}
}

func TestApplyAgentsReportsATemplateItCannotRead(t *testing.T) {
	a := &adoption{root: t.TempDir(), template: t.TempDir(), agents: agentsSkeleton}
	if err := a.applyAgents(); err == nil {
		t.Fatal("a template with no AGENTS.md was grafted from")
	}
}

func TestApplyAgentsReportsAnExistingDocumentItCannotRead(t *testing.T) {
	// The plan said graft, so an AGENTS.md was there when it was made;
	// one that cannot be read by the time apply runs is a fault.
	template := t.TempDir()
	write(t, template, "AGENTS.md", templateAgents)
	a := &adoption{root: t.TempDir(), template: template, agents: agentsGraft}
	if err := a.applyAgents(); err == nil {
		t.Fatal("a document that is not there was grafted onto")
	}
}

func TestRewriteFileReportsWhatItCouldNotRead(t *testing.T) {
	same := func(s string) (string, error) { return s, nil }
	if err := rewriteFile(filepath.Join(t.TempDir(), "not-there"), same); err == nil {
		t.Fatal("a file that is not there was rewritten")
	}

	skipAsRoot(t)
	dir := t.TempDir()
	path := filepath.Join(dir, "commits.md")
	if err := os.WriteFile(path, []byte("# Commits\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	readOnly(t, path, 0o000)
	if err := rewriteFile(path, same); err == nil {
		t.Fatal("a file that cannot be read was rewritten")
	}
}

func TestApplyVocabularyReportsTheFirstFileItCouldNotRewrite(t *testing.T) {
	// The conventions file is missing, so the first rewrite fails and
	// the second is never reached.
	err := applyVocabulary(t.TempDir(), vocabulary{Types: []string{"feat"}, Source: "history"})
	if err == nil {
		t.Fatal("a vocabulary was applied to a kit that is not there")
	}
}

func TestTheForgeReadsThatCouldNotAnswerAreNamed(t *testing.T) {
	// gh answers, then fails on each read in turn: every gap names the
	// setting it could not read rather than reporting it as unset.
	failing := func(args ...string) (string, error) {
		if len(args) > 1 && args[0] == "auth" {
			return "", nil
		}
		return "", errors.New("the API said no\nand said more about it")
	}
	d := Deps{
		LookPath: func(name string) (string, error) { return "/usr/bin/" + name, nil },
		Gh:       failing,
	}
	gaps := checkStages(makeTarget(t), 3, d)
	var text string
	for _, g := range gaps {
		text += g.Text + "\n"
	}
	for _, want := range []string{"squash merging", "workflow permissions", "Issues are enabled"} {
		if !strings.Contains(text, want) {
			t.Errorf("no gap names %q:\n%s", want, text)
		}
	}
	// firstLine: the reason is one line, never the API's whole answer.
	if strings.Contains(text, "and said more about it") {
		t.Errorf("a multi-line reason was reported whole:\n%s", text)
	}
}

func TestAMissingAgentsIsNamedAsAGap(t *testing.T) {
	target := makeTarget(t)
	gaps := checkFiles(target)
	var text string
	for _, g := range gaps {
		text += g.Text + "\n"
	}
	if !strings.Contains(text, "AGENTS.md — the agents' entry point is missing") {
		t.Errorf("an absent AGENTS.md was not named:\n%s", text)
	}
}

func TestAllClearIsSaidWhenNothingGapes(t *testing.T) {
	var out bytes.Buffer
	reportGaps(&out, nil, 1)
	if !strings.Contains(out.String(), "all clear") {
		t.Errorf("a clean check was not reported: %s", out.String())
	}
}

func TestThePlanNamesWhatTheProjectAlreadyOwns(t *testing.T) {
	// A skeleton the project already has a real file for is kept, and
	// the plan says so before the confirmation.
	target := makeTarget(t)
	write(t, target, "docs/product/README.md", "# Our own product docs\n")
	gitT(t, target, "add", "-A")
	gitT(t, target, "commit", "-q", "-m", "our docs")

	err, out := runInit(t, target, Deps{Tag: testTag, Source: makeSource(t)}, &command.FakeTerminal{}, true, "--stage", "1")
	if err != nil {
		t.Fatalf("init: %v\n%s", err, out)
	}
	if !strings.Contains(out, "kept         docs/product/README.md") {
		t.Errorf("the plan does not name the file it left alone:\n%s", out)
	}
	if got := read(t, target, "docs/product/README.md"); got != "# Our own product docs\n" {
		t.Errorf("the project's file was overwritten: %q", got)
	}
}

func TestThePlanSaysWhichWayAgentsGoes(t *testing.T) {
	source := makeSource(t)

	// Present without the markers: grafted.
	target := makeTarget(t)
	write(t, target, "AGENTS.md", "# Ours\n\nRules we already had.\n")
	gitT(t, target, "add", "-A")
	gitT(t, target, "commit", "-q", "-m", "agents")
	err, out := runInit(t, target, Deps{Tag: testTag, Source: source}, &command.FakeTerminal{}, true, "--stage", "1")
	if err != nil {
		t.Fatalf("init: %v\n%s", err, out)
	}
	if !strings.Contains(out, "AGENTS.md    graft") {
		t.Errorf("the plan does not say it grafts:\n%s", out)
	}

	// Present with the markers already: left alone.
	kept := makeTarget(t)
	write(t, kept, "AGENTS.md", templateAgents)
	gitT(t, kept, "add", "-A")
	gitT(t, kept, "commit", "-q", "-m", "agents")
	err, out = runInit(t, kept, Deps{Tag: testTag, Source: source}, &command.FakeTerminal{}, true, "--stage", "1")
	if err != nil {
		t.Fatalf("init: %v\n%s", err, out)
	}
	if !strings.Contains(out, "already carries the fenced markers") {
		t.Errorf("the plan does not say it leaves it alone:\n%s", out)
	}
}

func TestThePlanNamesTheExtractedVocabulary(t *testing.T) {
	// A history that used scopes, and one that never did: the plan
	// reports each differently, because absence is no vote against the
	// shipped list.
	scoped := makeTarget(t, "feat(product): a thing", "fix(technical): another")
	err, out := runInit(t, scoped, Deps{Tag: testTag, Source: makeSource(t)}, &command.FakeTerminal{}, true, "--stage", "1")
	if err != nil {
		t.Fatalf("init: %v\n%s", err, out)
	}
	if !strings.Contains(out, "conventions  extracted from") {
		t.Errorf("the plan does not name the source of the vocabulary:\n%s", out)
	}
	if !strings.Contains(out, "scopes product, technical") && !strings.Contains(out, "scopes technical, product") {
		t.Errorf("the plan does not name the extracted scopes:\n%s", out)
	}

	bare := makeTarget(t, "feat: a thing", "fix: another")
	err, out = runInit(t, bare, Deps{Tag: testTag, Source: makeSource(t)}, &command.FakeTerminal{}, true, "--stage", "1")
	if err != nil {
		t.Fatalf("init: %v\n%s", err, out)
	}
	if !strings.Contains(out, "scopes stay shipped") {
		t.Errorf("a history with no scope did not leave the shipped list standing:\n%s", out)
	}
}

func TestAForeignHookStopsTheAdoptionBeforeTheNetwork(t *testing.T) {
	target := makeTarget(t)
	hookAt := strings.TrimSpace(gitT(t, target, "rev-parse", "--git-path", "hooks/commit-msg"))
	if !filepath.IsAbs(hookAt) {
		hookAt = filepath.Join(target, hookAt)
	}
	if err := os.MkdirAll(filepath.Dir(hookAt), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(hookAt, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	err, _ := runInit(t, target, Deps{Tag: testTag, Source: makeSource(t)}, &command.FakeTerminal{}, true, "--stage", "1")
	if err == nil {
		t.Fatal("an adoption over a foreign hook succeeded")
	}
	if !strings.Contains(err.Error(), "refuses to overwrite it") {
		t.Errorf("the refusal does not name the hook: %v", err)
	}
	if _, statErr := os.Stat(filepath.Join(target, ".writrun")); statErr == nil {
		t.Error("the refusal still wrote the kit")
	}
}

func TestAOneLineForgeReasonIsReportedWhole(t *testing.T) {
	// firstLine returns the whole string where there is no newline to
	// cut at — the common case, and the one that must not lose text.
	d := Deps{
		LookPath: func(name string) (string, error) { return "/usr/bin/" + name, nil },
		Gh: func(args ...string) (string, error) {
			if len(args) > 1 && args[0] == "auth" {
				return "", nil
			}
			return "", errors.New("a single line of reason")
		},
	}
	gaps := checkStages(makeTarget(t), 2, d)
	var text string
	for _, g := range gaps {
		text += g.Text + "\n"
	}
	if !strings.Contains(text, "a single line of reason") {
		t.Errorf("a one-line reason was not reported:\n%s", text)
	}
}
