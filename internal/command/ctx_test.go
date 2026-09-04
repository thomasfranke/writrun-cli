package command

import (
	"strings"
	"testing"
)

func TestAskConfirm(t *testing.T) {
	cases := []struct {
		name    string
		yes     bool
		tty     bool
		answer  bool
		want    bool
		wantErr string
		asked   int
	}{
		{"--yes answers without asking", true, false, false, true, "", 0},
		{"terminal asks and yes", false, true, true, true, "", 1},
		{"terminal asks and no", false, true, false, false, "", 1},
		{"no terminal and no flag aborts naming the flag", false, false, false, false, "--yes", 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ft := &FakeTerminal{In: tc.tty, ConfirmAnswer: tc.answer}
			c := &Ctx{Terminal: ft, Yes: tc.yes}
			got, err := c.AskConfirm("push the branch?")
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("err = %v; want it to name %s", err, tc.wantErr)
				}
			} else if err != nil {
				t.Fatalf("err = %v", err)
			} else if got != tc.want {
				t.Fatalf("answer = %v, want %v", got, tc.want)
			}
			if len(ft.Asked) != tc.asked {
				t.Fatalf("asked %d questions, want %d", len(ft.Asked), tc.asked)
			}
		})
	}
}

func TestAskSelect(t *testing.T) {
	options := []string{"one", "two", "three"}
	cases := []struct {
		name    string
		preset  string
		tty     bool
		index   int
		want    int
		wantErr string
	}{
		{"preset answers without asking", "two", false, 0, 1, ""},
		{"preset not an option", "four", true, 0, -1, "not one of the options"},
		{"terminal selects", "", true, 2, 2, ""},
		{"no terminal and no flag aborts naming the flag", "", false, 0, -1, "--stage"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ft := &FakeTerminal{In: tc.tty, SelectIndex: tc.index}
			c := &Ctx{Terminal: ft}
			got, err := c.AskSelect("pick a stage", options, tc.preset, "--stage")
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("err = %v; want it to name %q", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("err = %v", err)
			}
			if got != tc.want {
				t.Fatalf("index = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestFakeTerminalSpinRunsTheWork(t *testing.T) {
	ft := &FakeTerminal{}
	ran := false
	if err := ft.Spin("waiting", func() error { ran = true; return nil }); err != nil || !ran {
		t.Fatalf("Spin ran=%v err=%v", ran, err)
	}
}
