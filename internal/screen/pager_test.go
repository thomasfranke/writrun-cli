package screen

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func longOutput(n int) string {
	var b strings.Builder
	for i := 1; i <= n; i++ {
		fmt.Fprintf(&b, "line %d\n", i)
	}
	return b.String()
}

// sized is a pager over n lines in a window of h.
func sized(t *testing.T, n, h int) pager {
	t.Helper()
	p := newPager("status", longOutput(n))
	out, _ := p.Update(tea.WindowSizeMsg{Width: 80, Height: h})
	return out.(pager)
}

// The whole answer is reachable, which is the point: an alternate
// screen keeps no scrollback, so output taller than the window used to
// scroll off with no way back to it.
func TestEveryLineIsReachable(t *testing.T) {
	p := sized(t, 50, 14) // ten rows of window
	if !strings.Contains(p.View(), "line 1") {
		t.Fatal("the answer does not start at its first line")
	}
	out, _ := p.Update(tea.KeyMsg{Type: tea.KeyEnd})
	p = out.(pager)
	if !strings.Contains(p.View(), "line 50") {
		t.Errorf("the last line is out of reach:\n%s", p.View())
	}
	out, _ = p.Update(tea.KeyMsg{Type: tea.KeyHome})
	if !strings.Contains(out.(pager).View(), "line 1") {
		t.Error("the first line is out of reach again")
	}
}

// Scrolling stops at the ends rather than running past them.
func TestScrollingStopsAtBothEnds(t *testing.T) {
	p := sized(t, 50, 14)
	for i := 0; i < 200; i++ {
		out, _ := p.Update(tea.KeyMsg{Type: tea.KeyDown})
		p = out.(pager)
	}
	if p.top != len(p.lines)-p.height {
		t.Errorf("top = %d, want the last window at %d", p.top, len(p.lines)-p.height)
	}
	for i := 0; i < 200; i++ {
		out, _ := p.Update(tea.KeyMsg{Type: tea.KeyUp})
		p = out.(pager)
	}
	if p.top != 0 {
		t.Errorf("top = %d, want the first line", p.top)
	}
}

// An answer that fits says so by offering no scrolling: a footer that
// names a key the screen does not answer is a footer that lies.
func TestAShortAnswerOffersNoScrolling(t *testing.T) {
	short := sized(t, 3, 24)
	if strings.Contains(short.View(), "scroll") {
		t.Errorf("a short answer offered scrolling:\n%s", short.View())
	}
	long := sized(t, 90, 24)
	if !strings.Contains(long.View(), "scroll") {
		t.Error("a long answer did not offer scrolling")
	}
}

// esc and enter go back; q leaves. The pager decides neither — it says
// which, and the session acts.
func TestThePagerNamesWhatWasAskedOfIt(t *testing.T) {
	for _, k := range []tea.KeyMsg{
		{Type: tea.KeyEsc},
		{Type: tea.KeyEnter},
	} {
		out, _ := sized(t, 5, 24).Update(k)
		if !out.(pager).back {
			t.Errorf("%v did not ask to go back", k)
		}
	}
	out, _ := sized(t, 5, 24).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if !out.(pager).left {
		t.Error("q did not ask to leave")
	}
}

// A command that wrote nothing still says so, rather than showing a
// blank screen the reader has to interpret.
func TestSilenceIsSaidRatherThanShown(t *testing.T) {
	if !strings.Contains(newPager("doctor", "").View(), "said nothing") {
		t.Error("an empty answer rendered as an empty screen")
	}
}

// The heading names the command, because the answer alone does not.
func TestThePagerNamesTheCommand(t *testing.T) {
	if !strings.Contains(newPager("doctor", "all clear\n").View(), "doctor") {
		t.Error("the answer does not say what it answers for")
	}
}
