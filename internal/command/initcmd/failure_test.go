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
	"github.com/thomasfranke/writrun-cli/internal/kitfetch"

	"github.com/thomasfranke/writrun-cli/internal/vfs"
)

func TestNewDefaultsTheSourceToTheCanonicalRepository(t *testing.T) {
	// An empty Source is the canonical repository, resolved once at
	// wiring time; the fetch is asked for that one, never for an empty
	// URL.
	target := makeTarget(t)
	kit := fakeKit(t)
	out, err := runInit(t, target, Deps{Tag: testTag, Kit: kit}, &command.FakeTerminal{}, true, "--stage", "1")
	if err != nil {
		t.Fatalf("init = %v\n%s", err, out)
	}
	if len(kit.Asked) != 1 {
		t.Fatalf("the fetch was asked %d times, want 1: %v", len(kit.Asked), kit.Asked)
	}
	if kit.Asked[0].Source != "https://github.com/thomasfranke/writrun" {
		t.Errorf("the default source was not used: %q", kit.Asked[0].Source)
	}
}

func TestAnUnknownFlagIsRefused(t *testing.T) {
	target := makeTarget(t)
	_, err := runInit(t, target, Deps{Tag: testTag}, &command.FakeTerminal{}, true, "--nope")
	if err == nil {
		t.Fatal("an unknown flag was accepted")
	}
}

func TestAWorkingTreeGitCannotReadStopsTheAdoption(t *testing.T) {
	// Outside a repository there is no tree to read, and an adoption
	// may not proceed without knowing whether one is dirty.
	target := t.TempDir()
	_, err := runInit(t, target, Deps{Tag: testTag}, &command.FakeTerminal{}, true, "--stage", "1")
	if err == nil {
		t.Fatal("the adoption proceeded outside a repository")
	}
	if !strings.Contains(err.Error(), "working tree") && !strings.Contains(err.Error(), "hooks directory") {
		t.Errorf("the error names neither read: %v", err)
	}
}

func TestATemplateWithNoAgentsIsNotAWritRunRepository(t *testing.T) {
	template := makeTemplate(t)
	if err := os.Remove(filepath.Join(template, "AGENTS.md")); err != nil {
		t.Fatal(err)
	}
	if _, err := plan(vfs.OS{}, makeTarget(t), template, testTag, "the source", 1, filepath.Join(t.TempDir(), "commit-msg"), gitx.Run); err == nil {
		t.Fatal("a template with no AGENTS.md was planned over")
	}
}

func TestApplyAgentsReportsATemplateItCannotRead(t *testing.T) {
	a := &adoption{disk: vfs.OS{}, root: t.TempDir(), template: t.TempDir(), agents: agentsSkeleton}
	if err := a.applyAgents(); err == nil {
		t.Fatal("a template with no AGENTS.md was grafted from")
	}
}

func TestApplyAgentsReportsAnExistingDocumentItCannotRead(t *testing.T) {
	// The plan said graft, so an AGENTS.md was there when it was made;
	// one that cannot be read by the time apply runs is a fault.
	template := t.TempDir()
	write(t, template, "AGENTS.md", templateAgents)
	a := &adoption{disk: vfs.OS{}, root: t.TempDir(), template: template, agents: agentsGraft}
	if err := a.applyAgents(); err == nil {
		t.Fatal("a document that is not there was grafted onto")
	}
}

func TestApplyVocabularyReportsTheFirstFileItCouldNotRewrite(t *testing.T) {
	// The conventions file is missing, so the first rewrite fails and
	// the second is never reached.
	err := applyVocabulary(vfs.OS{}, t.TempDir(), vocabulary{Types: []string{"feat"}, Source: "history"})
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
		Files:    vfs.OS{},
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
	gaps := checkFiles(vfs.OS{}, target)
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

	out, err := runInit(t, target, Deps{Tag: testTag}, &command.FakeTerminal{}, true, "--stage", "1")
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
	kit := fakeKit(t)

	// Present without the markers: grafted.
	target := makeTarget(t)
	write(t, target, "AGENTS.md", "# Ours\n\nRules we already had.\n")
	gitT(t, target, "add", "-A")
	gitT(t, target, "commit", "-q", "-m", "agents")
	out, err := runInit(t, target, Deps{Tag: testTag, Kit: kit}, &command.FakeTerminal{}, true, "--stage", "1")
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
	out, err = runInit(t, kept, Deps{Tag: testTag, Kit: kit}, &command.FakeTerminal{}, true, "--stage", "1")
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
	out, err := runInit(t, scoped, Deps{Tag: testTag}, &command.FakeTerminal{}, true, "--stage", "1")
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
	out, err = runInit(t, bare, Deps{Tag: testTag}, &command.FakeTerminal{}, true, "--stage", "1")
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

	_, err := runInit(t, target, Deps{Tag: testTag}, &command.FakeTerminal{}, true, "--stage", "1")
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
		Files:    vfs.OS{},
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

// fakeTemplate is a kit as the fake holds it — enough of one for the
// plan to be made and applied without a clone.
func fakeTemplate(t *testing.T) (*vfs.Fake, string, string) {
	t.Helper()
	disk := vfs.NewFake()
	root, template := "/repo", "/kit/template"
	disk.Seed(template+"/AGENTS.md", []byte(templateAgents), 0o644)
	disk.Seed(template+"/WRITRUN.md", []byte("# This project uses WritRun\n"), 0o644)
	disk.Seed(template+"/.writrun/settings.json", []byte(templateSettings), 0o644)
	disk.Seed(template+"/.writrun/conventions/commits.md", []byte(templateCommits), 0o644)
	disk.SeedDir(root)
	return disk, root, template
}

func TestApplyReportsTheCopyItCouldNotMake(t *testing.T) {
	disk, root, template := fakeTemplate(t)
	boom := errors.New("that file will not land")
	disk.Fail(root+"/WRITRUN.md", boom)

	a := &adoption{disk: disk, root: root, template: template, tag: testTag,
		copies:   []copyStep{{src: template + "/WRITRUN.md", rel: "WRITRUN.md", mode: 0o644}},
		hookPath: root + "/.git/hooks/commit-msg"}
	err := a.apply()
	if err == nil {
		t.Fatal("an adoption that cannot copy succeeded")
	}
	if !errors.Is(err, boom) {
		t.Errorf("the cause did not survive: %v", err)
	}
	if !strings.Contains(err.Error(), "copying WRITRUN.md") {
		t.Errorf("the error does not name the file: %v", err)
	}
}

func TestApplyReportsTheTagItCouldNotRecord(t *testing.T) {
	disk, root, template := fakeTemplate(t)
	disk.Seed(root+"/.writrun/settings.json", []byte(templateSettings), 0o644)
	boom := errors.New("VERSION will not be written")
	disk.Fail(root+"/.writrun/VERSION", boom)

	a := &adoption{disk: disk, root: root, template: template, tag: testTag, stage: 1,
		hookPath: root + "/.git/hooks/commit-msg"}
	err := a.apply()
	if !errors.Is(err, boom) {
		t.Fatalf("recording the tag: %v", err)
	}
	if !strings.Contains(err.Error(), "recording the tag") {
		t.Errorf("the error does not name the act: %v", err)
	}
}

func TestApplyAgentsReportsTheDocumentItCouldNotWrite(t *testing.T) {
	disk, root, template := fakeTemplate(t)
	boom := errors.New("AGENTS.md will not be written")
	disk.Fail(root+"/AGENTS.md", boom)

	a := &adoption{disk: disk, root: root, template: template, agents: agentsSkeleton}
	if err := a.applyAgents(); !errors.Is(err, boom) {
		t.Errorf("writing the skeleton: %v", err)
	}

	disk.Heal(root + "/AGENTS.md")
	disk.Seed(root+"/AGENTS.md", []byte("# Ours\n"), 0o644)
	disk.Fail(root+"/AGENTS.md", boom)
	a.agents = agentsGraft
	if err := a.applyAgents(); !errors.Is(err, boom) {
		t.Errorf("grafting onto the document: %v", err)
	}
}

func TestPlanReportsATemplateItCannotWalk(t *testing.T) {
	disk, root, template := fakeTemplate(t)
	disk.Fail(template, errors.New("the template cannot be read"))
	if _, err := plan(disk, root, template, testTag, "src", 1, "/hooks/commit-msg", gitx.Run); err == nil {
		t.Error("a template that cannot be walked was planned over")
	}
}

func TestRewriteFileReportsWhatItCouldNotTouch(t *testing.T) {
	disk := vfs.NewFake()
	same := func(s string) (string, error) { return s, nil }

	if err := rewriteFile(disk, "/repo/not-there.md", same); err == nil {
		t.Fatal("a file that is not there was rewritten")
	}

	disk.Seed("/repo/commits.md", []byte("# Commits\n"), 0o644)
	boom := errors.New("that file, no")
	disk.Fail("/repo/commits.md", boom)
	if err := rewriteFile(disk, "/repo/commits.md", same); !errors.Is(err, boom) {
		t.Errorf("a file that cannot be read was rewritten: %v", err)
	}
}

// fakeGit answers the two reads init makes through git: the working
// tree's status, and where the commit-msg hook lives.
func fakeGit(root string) gitx.Runner {
	return func(dir string, args ...string) (string, error) {
		if len(args) > 0 && args[0] == "rev-parse" {
			return filepath.Join(root, ".git", "hooks", "commit-msg") + "\n", nil
		}
		return "", nil
	}
}

func TestAPartialAdoptionNamesTheCommandsThatUndoIt(t *testing.T) {
	// The write fails after the fetch succeeded, which is the one
	// state the adoption cannot leave clean: the message is what tells
	// the user how to get back (spec-0016).
	disk, root, template := fakeTemplate(t)
	disk.FailOp("write", root+"/WRITRUN.md", errors.New("that file will not land"))

	d := Deps{Tag: testTag, Source: "the source", Git: fakeGit(root), Files: disk, Kit: kitfetch.NewFake(template)}
	_, err := runInit(t, root, d, &command.FakeTerminal{}, true, "--stage", "1")
	if err == nil {
		t.Fatal("an adoption that could not write succeeded")
	}
	for _, want := range []string{"the adoption is partial", "git checkout -- .", "git clean -fd", "rm -f " + root + "/.git/hooks/commit-msg", "rerun writrun init"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the partial-state message does not name %q: %v", want, err)
		}
	}
}

func TestTheFetchIsCleanedUpWhateverTheAdoptionDid(t *testing.T) {
	// The cleanup is the fetch's half of the contract: a checkout the
	// command never releases is a leak the fake has to be able to see.
	kit := fakeKit(t)
	out, err := runInit(t, makeTarget(t), Deps{Tag: testTag, Kit: kit}, &command.FakeTerminal{}, true, "--stage", "1")
	if err != nil {
		t.Fatalf("init = %v\n%s", err, out)
	}
	if kit.Cleaned != 1 {
		t.Errorf("the fetch was cleaned up %d times, want 1", kit.Cleaned)
	}

	failing := fakeKit(t)
	disk, root, template := fakeTemplate(t)
	failing.Template = template
	disk.FailOp("write", root+"/WRITRUN.md", errors.New("that file will not land"))
	d := Deps{Tag: testTag, Git: fakeGit(root), Files: disk, Kit: failing}
	if _, err := runInit(t, root, d, &command.FakeTerminal{}, true, "--stage", "1"); err == nil {
		t.Fatal("an adoption that could not write succeeded")
	}
	if failing.Cleaned != 1 {
		t.Errorf("a failed adoption cleaned up %d times, want 1", failing.Cleaned)
	}
}
