package initcmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/thomasfranke/writrun-cli/internal/fence"
	"github.com/thomasfranke/writrun-cli/internal/hook"

	"github.com/thomasfranke/writrun-cli/internal/gitx"

	"github.com/thomasfranke/writrun-cli/internal/vfs"
)

// planFixture clones the test source the way run does and plans an
// adoption of the target at the given stage.
func planFixture(t *testing.T, target string, stage int) *adoption {
	t.Helper()
	src := makeSource(t)
	clone := filepath.Join(t.TempDir(), "writrun")
	gitT(t, "", "clone", "-q", "--depth", "1", "--branch", testTag, src, clone)
	hookAt, err := hook.Path(target, gitx.Run)
	if err != nil {
		t.Fatalf("hook.Path = %v", err)
	}
	a, err := plan(vfs.OS{}, target, filepath.Join(clone, "template"), testTag, src, stage, hookAt, gitx.Run)
	if err != nil {
		t.Fatalf("plan = %v", err)
	}
	return a
}

func TestPlanKeepsTheProjectsOwnFiles(t *testing.T) {
	target := makeTarget(t)
	write(t, target, "docs/product/README.md", "# The project's real product doc\n")
	gitT(t, target, "add", ".")
	gitT(t, target, "commit", "-q", "-m", "docs")

	a := planFixture(t, target, 1)
	if len(a.kept) != 1 || filepath.ToSlash(a.kept[0]) != "docs/product/README.md" {
		t.Errorf("kept = %v, want the owned skeleton", a.kept)
	}
	for _, c := range a.copies {
		if filepath.ToSlash(c.rel) == "docs/product/README.md" {
			t.Error("an owned file is also in the copy list")
		}
	}
}

func TestPlanDecidesTheAgentsAction(t *testing.T) {
	t.Run("absent means skeleton", func(t *testing.T) {
		a := planFixture(t, makeTarget(t), 1)
		if a.agents != agentsSkeleton {
			t.Errorf("agents = %v, want skeleton", a.agents)
		}
	})
	t.Run("present means graft", func(t *testing.T) {
		target := makeTarget(t)
		write(t, target, "AGENTS.md", "# Mine\n")
		gitT(t, target, "add", ".")
		gitT(t, target, "commit", "-q", "-m", "agents")
		a := planFixture(t, target, 1)
		if a.agents != agentsGraft {
			t.Errorf("agents = %v, want graft", a.agents)
		}
	})
	t.Run("markers already present means kept", func(t *testing.T) {
		target := makeTarget(t)
		write(t, target, "AGENTS.md", templateAgents)
		gitT(t, target, "add", ".")
		gitT(t, target, "commit", "-q", "-m", "agents")
		a := planFixture(t, target, 1)
		if a.agents != agentsKept {
			t.Errorf("agents = %v, want kept", a.agents)
		}
	})
}

func TestApplyPerformsTheWholePlan(t *testing.T) {
	target := makeTarget(t, "feat: begin", "feat(api): more")
	a := planFixture(t, target, 2)
	if err := a.apply(); err != nil {
		t.Fatalf("apply = %v", err)
	}

	if got := strings.TrimSpace(read(t, target, ".writrun/VERSION")); got != testTag {
		t.Errorf("VERSION = %q, want %q", got, testTag)
	}
	if settings := read(t, target, ".writrun/settings.json"); !strings.Contains(settings, `"stage": 2`) {
		t.Errorf("settings do not record stage 2:\n%s", settings)
	}
	if agents := read(t, target, "AGENTS.md"); !strings.Contains(agents, fence.Begin) {
		t.Error("AGENTS.md lacks the fenced section")
	}
	if observance := read(t, target, ".writrun/scripts/stage-2-pull-requests/check_observance.sh"); !strings.Contains(observance, `TYPES="feat"`) {
		t.Errorf("the extracted vocabulary did not land:\n%s", observance)
	}
	info, err := os.Stat(a.hookPath)
	if err != nil {
		t.Fatalf("the hook was not installed: %v", err)
	}
	if info.Mode()&0o100 == 0 {
		t.Error("the hook is not executable")
	}
	scriptInfo, err := os.Stat(filepath.Join(target, ".writrun/scripts/stage-2-pull-requests/check_observance.sh"))
	if err != nil || scriptInfo.Mode()&0o100 == 0 {
		t.Errorf("a copied script lost its executable bit: %v", err)
	}

	entries, err := os.ReadDir(filepath.Join(target, "work", "tasks"))
	if err != nil {
		t.Fatalf("the queue folder is missing: %v", err)
	}
	for _, e := range entries {
		if e.Name() != "README.md" {
			t.Errorf("the queue is not empty: %s", e.Name())
		}
	}
}

func TestApplyGraftKeepsEveryByteOutsideTheFence(t *testing.T) {
	target := makeTarget(t)
	existing := "# Mine\n\nMy own rules, byte for byte.\n"
	write(t, target, "AGENTS.md", existing)
	gitT(t, target, "add", ".")
	gitT(t, target, "commit", "-q", "-m", "agents")

	a := planFixture(t, target, 1)
	if err := a.apply(); err != nil {
		t.Fatalf("apply = %v", err)
	}
	after := read(t, target, "AGENTS.md")
	if !strings.HasPrefix(after, existing) {
		t.Error("bytes outside the fenced section changed")
	}
	if !strings.Contains(after, fence.End) {
		t.Error("the fenced section was not grafted")
	}
}

func TestRenderNamesEveryDecision(t *testing.T) {
	target := makeTarget(t)
	a := planFixture(t, target, 1)
	var out bytes.Buffer
	a.render(&out)
	for _, want := range []string{
		"WritRun " + testTag,
		"stage        1 — files",
		"shipped defaults",
		"AGENTS.md    absent",
		"commit-msg",
		"records stage 1",
	} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("the plan does not say %q:\n%s", want, out.String())
		}
	}
}

func TestSummarizeCopiesGroupsByTopLevel(t *testing.T) {
	got := summarizeCopies([]copyStep{
		{rel: filepath.Join(".writrun", "VERSION")},
		{rel: filepath.Join(".writrun", "settings.json")},
		{rel: "WRITRUN.md"},
	})
	if !strings.Contains(got, "3 files") || !strings.Contains(got, ".writrun/ (2 files)") || !strings.Contains(got, "WRITRUN.md") {
		t.Errorf("summarizeCopies = %q", got)
	}
}

func TestRenderNamesTheHookPathThatWillBeWritten(t *testing.T) {
	t.Run("inside the repository it is relative", func(t *testing.T) {
		a := planFixture(t, makeTarget(t), 1)
		var out bytes.Buffer
		a.render(&out)
		if !strings.Contains(out.String(), "hook         .git/hooks/commit-msg") {
			t.Errorf("the plan does not name the hook path:\n%s", out.String())
		}
	})
	// core.hooksPath moves the hook out of the repository, and a plan
	// still saying .git/hooks/ would be consent to something else.
	t.Run("a redirected hooks directory is named in full", func(t *testing.T) {
		target := makeTarget(t)
		hooks := t.TempDir()
		gitT(t, target, "config", "core.hooksPath", hooks)
		a := planFixture(t, target, 1)
		want := filepath.Join(hooks, "commit-msg")
		if a.hookPath != want {
			t.Fatalf("hookPath = %q, want %q", a.hookPath, want)
		}
		var out bytes.Buffer
		a.render(&out)
		if !strings.Contains(out.String(), want) {
			t.Errorf("the plan hides the redirected hook path:\n%s", out.String())
		}
	})
}

func TestRenderNamesTheSourceItFetchedFrom(t *testing.T) {
	target := makeTarget(t)
	a := planFixture(t, target, 1)
	var out bytes.Buffer
	a.render(&out)
	if !strings.Contains(out.String(), a.source) {
		t.Errorf("the plan does not name the source %q:\n%s", a.source, out.String())
	}
}

func TestApplyRefusesSettingsWithNoStageKey(t *testing.T) {
	target := makeTarget(t)
	a := planFixture(t, target, 3)
	// The copy step reads from the template at apply time, so this is
	// a kit whose settings init cannot write a stage into.
	write(t, a.template, ".writrun/settings.json", "{\n  \"stage_1\": {}\n}\n")
	err := a.apply()
	if err == nil || !strings.Contains(err.Error(), `no "stage" key`) {
		t.Fatalf("apply = %v, want the refusal rather than a silent no-op", err)
	}
	if settings := read(t, target, ".writrun/settings.json"); strings.Contains(settings, `"stage": 3`) {
		t.Errorf("a stage was written after all:\n%s", settings)
	}
}
