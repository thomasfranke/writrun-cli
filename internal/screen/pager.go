package screen

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// pager shows a command's output as a screen of its own: scrollable,
// with a way back.
//
// A command that asks nothing has its output captured, so the screen
// never has to give up the terminal for it — which is what buys the
// spinner while it runs and the scrolling afterwards. A command that
// asks keeps the terminal and prints as it always did; there is no
// capturing a question.
type pager struct {
	command string
	chrome  chrome
	lines   []string
	top     int
	height  int
	width   int
	// back and left are read by the session, as the other screens'
	// flags are: this decides nothing about what happens next.
	back bool
	left bool
}

func newPager(identity, command, output string) pager {
	lines := strings.Split(strings.TrimRight(output, "\n"), "\n")
	if len(lines) == 1 && lines[0] == "" {
		lines = []string{"  (it said nothing)"}
	}
	return pager{
		command: command,
		chrome: chrome{
			identity: identity,
			context:  strings.ToUpper(command) + " · run from the screen — reading its output",
		},
		lines: lines,
	}
}

func (p pager) Init() tea.Cmd { return nil }

func (p pager) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Three for the header and the blank under it, two for the
		// footer and the blank over it.
		p.width = msg.Width
		switch {
		case msg.Height == 0:
			p.height = 0
		case msg.Height > pagerChrome:
			p.height = msg.Height - pagerChrome
		default:
			p.height = 1
		}
		p.clamp()
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			p.scroll(-1)
		case "down", "j":
			p.scroll(1)
		case "pgup":
			p.scroll(-p.page())
		case "pgdown", " ":
			p.scroll(p.page())
		case "home", "g":
			p.top = 0
		case "end", "G":
			p.top = len(p.lines)
			p.clamp()
		case keyBack, keyTake:
			p.back = true
		case keyQuit, "ctrl+c":
			p.left = true
		}
	}
	return p, nil
}

func (p pager) page() int {
	if p.height > 1 {
		return p.height - 1
	}
	return 1
}

func (p *pager) scroll(by int) {
	p.top += by
	p.clamp()
}

// clamp keeps the window over the lines. A window taller than the
// output shows all of it and does not scroll at all.
func (p *pager) clamp() {
	last := len(p.lines) - p.height
	if p.height <= 0 || last < 0 {
		last = 0
	}
	if p.top > last {
		p.top = last
	}
	if p.top < 0 {
		p.top = 0
	}
}

// pagerChrome is the lines the pager keeps around the output: the two
// header lines, the blank under them, the blank over the footer, and
// the footer.
const pagerChrome = 5

func (p pager) View() string {
	var b strings.Builder
	for _, line := range p.chrome.lines(contentWidth(p.width)) {
		b.WriteString(line + "\n")
	}
	b.WriteByte('\n')

	end := window(len(p.lines), p.top, p.height)
	for i := p.top; i < end; i++ {
		b.WriteString(p.lines[i])
		b.WriteByte('\n')
	}

	b.WriteByte('\n')
	f := footer{way: backOrQuit}
	if p.height > 0 && len(p.lines) > p.height {
		f.movement = "↑↓ scroll"
	}
	b.WriteString(f.line() + "\n")
	return b.String()
}
