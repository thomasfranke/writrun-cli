package workcmd

import (
	"reflect"
	"strings"
	"testing"
)

func TestTheConfiguredCommandIsSplitIntoTheProgramAndItsArguments(t *testing.T) {
	cases := []struct {
		value string
		name  string
		args  []string
	}{
		{"claude", "claude", []string{}},
		{"  claude  ", "claude", []string{}},
		{"claude --print", "claude", []string{"--print"}},
		{"claude\t--print\n-v", "claude", []string{"--print", "-v"}},
		{`claude --append-system-prompt "be terse"`, "claude", []string{"--append-system-prompt", "be terse"}},
		{`claude --prompt 'two words'`, "claude", []string{"--prompt", "two words"}},
		{`/opt/my agent/run --flag`, "/opt/my", []string{"agent/run", "--flag"}},
		{`"/opt/my agent/run" --flag`, "/opt/my agent/run", []string{"--flag"}},
		{`codex exec ""`, "codex", []string{"exec", ""}},
	}
	for _, c := range cases {
		name, args, err := words(c.value)
		if err != nil {
			t.Errorf("words(%q) = %v", c.value, err)
			continue
		}
		if name != c.name || !reflect.DeepEqual(args, c.args) {
			t.Errorf("words(%q) = %q %v; want %q %v", c.value, name, args, c.name, c.args)
		}
	}
}

func TestAnUnclosedQuoteIsRefusedRatherThanGuessedAt(t *testing.T) {
	for _, value := range []string{`claude --prompt "unclosed`, `claude --prompt 'unclosed`} {
		if _, _, err := words(value); err == nil || !strings.Contains(err.Error(), "unclosed") {
			t.Errorf("words(%q) = %v; want the quote named", value, err)
		}
	}
}

func TestAValueNamingNoCommandIsRefused(t *testing.T) {
	for _, value := range []string{"", `""`, `'' --flag`} {
		if _, _, err := words(value); err == nil {
			t.Errorf("words(%q) named a command to launch", value)
		}
	}
}

func TestTheAgentIsReadFromTheRepositorysGitConfig(t *testing.T) {
	var asked []string
	git := func(dir string, args ...string) (string, error) {
		if dir != root {
			t.Errorf("dir = %q; want the repository root", dir)
		}
		asked = append(asked, strings.Join(args, " "))
		return "claude\n", nil
	}
	got, err := configured(git, root)
	if err != nil {
		t.Fatalf("configured = %v", err)
	}
	if got != "claude" {
		t.Errorf("agent = %q; want the configured command", got)
	}
	if want := []string{"config --get writrun.agent"}; !reflect.DeepEqual(asked, want) {
		t.Errorf("asked %v; want %v", asked, want)
	}
}
