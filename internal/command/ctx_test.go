package command

import (
	"errors"
	"strings"
	"testing"
)

func TestAskConfirm(t *testing.T) {
	cases := []struct {
		name    string
		yes     bool
		tty     bool
		answer  bool
		wantErr error
		wantMsg string
		asked   int
	}{
		{"--yes answers without asking", true, false, false, nil, "", 0},
		{"terminal asks and yes", false, true, true, nil, "", 1},
		{"terminal asks and no is ErrDeclined", false, true, false, ErrDeclined, "", 1},
		{"no terminal and no flag aborts naming the flag", false, false, false, nil, "--yes", 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ft := &FakeTerminal{In: tc.tty, ConfirmAnswer: tc.answer}
			c := &Ctx{Terminal: ft, Yes: tc.yes}
			err := c.AskConfirm("push the branch?")
			switch {
			case tc.wantMsg != "":
				if err == nil || !strings.Contains(err.Error(), tc.wantMsg) {
					t.Fatalf("err = %v; want it to name %s", err, tc.wantMsg)
				}
			case tc.wantErr != nil:
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("err = %v, want %v", err, tc.wantErr)
				}
			case err != nil:
				t.Fatalf("err = %v", err)
			}
			if len(ft.Asked) != tc.asked {
				t.Fatalf("asked %d questions, want %d", len(ft.Asked), tc.asked)
			}
		})
	}
}

func TestAskConfirmCarriesTheTerminalError(t *testing.T) {
	boom := errors.New("render failed")
	c := &Ctx{Terminal: &FakeTerminal{In: true, ConfirmErr: boom}}
	if err := c.AskConfirm("push the branch?"); !errors.Is(err, boom) {
		t.Fatalf("err = %v, want the terminal's own error", err)
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
