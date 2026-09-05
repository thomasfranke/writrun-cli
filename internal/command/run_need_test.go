package command

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

// frameWith builds a frame whose two filesystem-facing ports answer
// exactly what a case is about.
func frameWith(getwd func() (string, error), find func(string) (string, bool, error), cmds ...Command) (Frame, *bytes.Buffer) {
	var out bytes.Buffer
	return Frame{
		Version: "test", WritRunTag: "v0.0.0",
		Commands: cmds,
		Stdout:   &out, Stderr: &out,
		Terminal: &FakeTerminal{},
		FindRepo: find, Getwd: getwd,
		Getenv: func(string) string { return "" },
	}, &out
}

func okWd() (string, error) { return "/somewhere", nil }

func adoptedRepo(string) (string, bool, error)   { return "/somewhere", true, nil }
func unadoptedRepo(string) (string, bool, error) { return "/somewhere", false, nil }
func noRepo(string) (string, bool, error) {
	return "", false, errors.New("not inside a git repository")
}

func aCommand(need Need, ran *bool) Command {
	return Command{Name: "probe", Summary: "a probe", Need: need,
		Run: func(*Ctx, []string) error { *ran = true; return nil }}
}

func TestAWorkingDirectoryThatCannotBeReadStopsEveryNeed(t *testing.T) {
	for _, need := range []Need{NeedAny, NeedAdopted, NeedAbsent} {
		ran := false
		f, out := frameWith(
			func() (string, error) { return "", errors.New("the cwd is gone") },
			adoptedRepo, aCommand(need, &ran))
		if code := Run(f, []string{"probe"}); code != 1 {
			t.Errorf("need %v: exit = %d, want 1", need, code)
		}
		if ran {
			t.Errorf("need %v: the command ran without a working directory", need)
		}
		if !strings.Contains(out.String(), "the cwd is gone") {
			t.Errorf("need %v: the reason was not reported: %s", need, out.String())
		}
	}
}

func TestANeedThatWantsARepositoryReportsHavingNone(t *testing.T) {
	for _, need := range []Need{NeedAdopted, NeedAbsent} {
		ran := false
		f, out := frameWith(okWd, noRepo, aCommand(need, &ran))
		if code := Run(f, []string{"probe"}); code != 1 {
			t.Errorf("need %v: exit = %d, want 1", need, code)
		}
		if ran {
			t.Errorf("need %v: the command ran outside a repository", need)
		}
		if !strings.Contains(out.String(), "not inside a git repository") {
			t.Errorf("need %v: the reason was not reported: %s", need, out.String())
		}
	}
}

func TestNeedAnyRunsWithoutARepositoryAtAll(t *testing.T) {
	ran := false
	f, _ := frameWith(okWd, noRepo, aCommand(NeedAny, &ran))
	if code := Run(f, []string{"probe"}); code != 0 {
		t.Errorf("exit = %d, want 0", code)
	}
	if !ran {
		t.Error("a command needing nothing did not run")
	}
}

func TestNeedAdoptedRefusesAnUnadoptedRepository(t *testing.T) {
	ran := false
	f, out := frameWith(okWd, unadoptedRepo, aCommand(NeedAdopted, &ran))
	if code := Run(f, []string{"probe"}); code != 1 {
		t.Errorf("exit = %d, want 1", code)
	}
	if ran {
		t.Error("the command ran in an unadopted repository")
	}
	if !strings.Contains(out.String(), "not an adopted repository") {
		t.Errorf("the refusal was not reported: %s", out.String())
	}
}

func TestNeedAbsentRefusesAnAdoptedRepositoryPointingAtUpdate(t *testing.T) {
	ran := false
	f, out := frameWith(okWd, adoptedRepo, aCommand(NeedAbsent, &ran))
	if code := Run(f, []string{"probe"}); code != 1 {
		t.Errorf("exit = %d, want 1", code)
	}
	if ran {
		t.Error("the command ran in an already-adopted repository")
	}
	if !strings.Contains(out.String(), "writrun update") {
		t.Errorf("the refusal does not point at update: %s", out.String())
	}
}
