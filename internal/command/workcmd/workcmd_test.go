package workcmd

import (
	"bytes"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/thomasfranke/writrun-cli/internal/command"
	"github.com/thomasfranke/writrun-cli/internal/kit"
)

func TestNoAgentConfiguredAbortsShowingTheLineThatSetsOneAndLaunchesNothing(t *testing.T) {
	s := scripts()
	got := work(t, noAgent(), s, nil)

	got.wantsError(t, "git config writrun.agent")
	got.wantsNoLaunch(t)
	if len(s.runs) != 0 {
		t.Errorf("ran %v; want nothing selected where nothing can be launched", s.ran())
	}
}

func TestAnEmptyAgentValueIsNoAgentAtAll(t *testing.T) {
	got := work(t, configuredAgent("   "), scripts(), nil)
	got.wantsError(t, "git config writrun.agent")
	got.wantsNoLaunch(t)
}

func TestAGitThatCannotAnswerIsAnErrorNotAnUnsetKey(t *testing.T) {
	boom := errors.New("git is not installed")
	got := work(t, gitFails(boom), scripts(), nil)
	if !errors.Is(got.err, boom) {
		t.Errorf("err = %v; want the cause preserved", got.err)
	}
	got.wantsNoLaunch(t)
}

func TestWithNoArgumentTheFirstAvailableTaskIsTheOneLaunched(t *testing.T) {
	s := scripts()
	got := work(t, configuredAgent("claude"), s, nil)
	if got.err != nil {
		t.Fatalf("run = %v", got.err)
	}
	if !strings.Contains(got.prompted(t), "Work task-0007 in this repository") {
		t.Errorf("the launch did not name the lister's first available task:\n%s", got.prompted(t))
	}
	if want := []string{listScript, briefScript}; !reflect.DeepEqual(s.ran(), want) {
		t.Errorf("ran %v; want %v", s.ran(), want)
	}
}

func TestTheBriefIsAssembledByTheSkillsOwnScriptFromTheRoot(t *testing.T) {
	s := scripts()
	if got := work(t, configuredAgent("claude"), s, nil); got.err != nil {
		t.Fatalf("run = %v", got.err)
	}
	last := s.runs[len(s.runs)-1]
	if last.script != briefScript {
		t.Errorf("script = %q; want the selection skill's brief", last.script)
	}
	if last.dir != root {
		t.Errorf("dir = %q; want the repository root", last.dir)
	}
	if want := []string{"task-0007"}; !reflect.DeepEqual(last.args, want) {
		t.Errorf("args = %v; want %v", last.args, want)
	}
}

func TestTheLaunchedCommandReceivesTheBriefUnedited(t *testing.T) {
	got := work(t, configuredAgent("claude"), scripts(), nil)
	if got.err != nil {
		t.Fatalf("run = %v", got.err)
	}
	prompt := got.prompted(t)
	if !strings.HasSuffix(prompt, brief) {
		t.Errorf("the brief did not reach the launch whole:\n%s", prompt)
	}
	if !strings.Contains(prompt, agentsFile) {
		t.Errorf("the launch does not point at %s:\n%s", agentsFile, prompt)
	}
	if got.launched[0].Dir != root {
		t.Errorf("dir = %q; want the repository root", got.launched[0].Dir)
	}
	if got.launched[0].Name != "claude" {
		t.Errorf("name = %q; want the configured command", got.launched[0].Name)
	}
}

func TestTheConfiguredCommandsOwnArgumentsComeBeforeTheBrief(t *testing.T) {
	got := work(t, configuredAgent(`claude --model opus --append-system-prompt "be terse"`), scripts(), nil)
	if got.err != nil {
		t.Fatalf("run = %v", got.err)
	}
	args := got.launched[0].Args
	if want := []string{"--model", "opus", "--append-system-prompt", "be terse"}; !reflect.DeepEqual(args[:len(args)-1], want) {
		t.Errorf("args = %v; want %v before the brief", args[:len(args)-1], want)
	}
}

func TestAnAgentCommandThatNamesNothingIsRefused(t *testing.T) {
	got := work(t, configuredAgent(`"`), scripts(), nil)
	got.wantsError(t, "unclosed")
	got.wantsNoLaunch(t)
}

func TestAPartialBriefIsShownAndTheLaunchRefused(t *testing.T) {
	s := scripts()
	s.brief = "task-0007  ready\n\n== work/tasks/task-0007-work-command.md ==\n"
	s.briefOut = "Incomplete brief — unresolved: spec-0007\n"
	s.briefErr = exitErr(2)

	got := work(t, configuredAgent("claude"), s, nil)
	got.wantsError(t, "incomplete")
	got.wantsNoLaunch(t)
	if !strings.Contains(got.stdout, "== work/tasks/task-0007-work-command.md ==") {
		t.Errorf("the half that resolved was swallowed:\n%s", got.stdout)
	}
	if !strings.Contains(got.stderr, "Incomplete brief") {
		t.Errorf("stderr = %q; want the script's own words", got.stderr)
	}
}

func TestABriefResolvingNoTaskIsRefusedInTheScriptsOwnWords(t *testing.T) {
	s := scripts()
	s.brief = ""
	s.briefOut = "No task in work/tasks/ resolves 'task-0007'\n"
	s.briefErr = exitErr(1)

	got := work(t, configuredAgent("claude"), s, nil)
	got.wantsError(t, briefScript)
	got.wantsNoLaunch(t)
	if !strings.Contains(got.stderr, "No task in work/tasks/") {
		t.Errorf("stderr = %q; want the script's own refusal", got.stderr)
	}
}

func TestABriefScriptThatCouldNotRunIsSaidSo(t *testing.T) {
	s := scripts()
	s.briefErr = errors.New("bash is not installed")
	got := work(t, configuredAgent("claude"), s, nil)
	got.wantsError(t, "bash is not installed")
	got.wantsNoLaunch(t)
}

func TestTheAgentsOwnExitTravelsUp(t *testing.T) {
	launcher := &FakeLauncher{Err: exitErr(3)}
	got := work(t, configuredAgent("claude"), scripts(), launcher)
	if exitCode(got.err) != 3 {
		t.Errorf("exit = %d; want the agent's own 3", exitCode(got.err))
	}
}

func TestAnAgentThatNeverStartedIsNamed(t *testing.T) {
	boom := errors.New("executable file not found in $PATH")
	launcher := &FakeLauncher{Err: boom}
	got := work(t, configuredAgent("claude"), scripts(), launcher)
	if !errors.Is(got.err, boom) {
		t.Errorf("err = %v; want the cause preserved", got.err)
	}
	got.wantsError(t, "launching claude")
}

func TestTwoTaskIdsAreRefused(t *testing.T) {
	got := work(t, configuredAgent("claude"), scripts(), nil, "task-0007", "task-0011")
	got.wantsError(t, "work takes one")
	got.wantsNoLaunch(t)
}

func TestAnUnknownFlagIsRefused(t *testing.T) {
	got := work(t, configuredAgent("claude"), scripts(), nil, "--all")
	if got.err == nil {
		t.Fatal("an unknown flag was accepted")
	}
	got.wantsNoLaunch(t)
}

func TestNothingButTheTwoWrappedScriptsIsEverRun(t *testing.T) {
	s := scripts()
	if got := work(t, configuredAgent("claude"), s, nil); got.err != nil {
		t.Fatalf("run = %v", got.err)
	}
	for _, r := range s.runs {
		if r.script != listScript && r.script != briefScript {
			t.Errorf("ran %q; work reads the queue and writes nothing to it", r.script)
		}
		if len(r.args) > 0 && r.script == listScript {
			t.Errorf("the lister was given %v; want the script's own defaults", r.args)
		}
	}
}

func TestNewDeclaresTheCommand(t *testing.T) {
	s := scripts()
	launcher := &FakeLauncher{}
	var runner kit.Runner = s.run
	c := New(Deps{Git: configuredAgent("claude"), Scripts: runner, Launch: launcher.Run})
	if c.Name != "work" {
		t.Errorf("name = %q", c.Name)
	}
	if c.Need != command.NeedAdopted {
		t.Errorf("need = %v; want an adopted repository", c.Need)
	}
	if c.Summary == "" {
		t.Error("no summary for --help")
	}
	var stdout bytes.Buffer
	if err := c.Run(&command.Ctx{Stdout: &stdout, Stderr: io.Discard, Root: root}, nil); err != nil {
		t.Fatalf("Run = %v", err)
	}
	if len(launcher.Launched) != 1 {
		t.Error("the wired command launched nothing")
	}
	if !strings.Contains(stdout.String(), "task-0007 — launching claude") {
		t.Errorf("output = %q; want the task and the command named", stdout.String())
	}
}
