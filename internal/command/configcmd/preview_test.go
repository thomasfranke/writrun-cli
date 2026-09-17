package configcmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/thomasfranke/writrun-cli/internal/command"
	"github.com/thomasfranke/writrun-cli/internal/vfs"
)

// stubPreview stands in for doctor: it records what it was asked and
// writes the rows the drawing shows for a stage-1 repository raising to
// stage 2 (docs/product/screens/adoption/config.excalidraw).
type stubPreview struct {
	asked []int
	fail  error
}

func (s *stubPreview) run(_ string, stage int, w io.Writer) (int, int, int, error) {
	s.asked = append(s.asked, stage)
	if s.fail != nil {
		return 0, 0, 0, s.fail
	}
	for _, row := range []struct{ mark, text string }{
		{"✓", "gh on the PATH"},
		{"✓", "gh authenticated"},
		{"✓", "squash merging is on"},
		{"✗", "the recording push can write to main"},
		{"?", "main is governed by a ruleset"},
		{"?", "no rule over main refuses the recording push"},
	} {
		fmt.Fprintf(w, "         %s  %s\n", row.mark, row.text)
	}
	return 3, 1, 2, nil
}

// raise runs `config stage <value>` against a repository declaring
// `from`, with the preview wired.
func raise(t *testing.T, from, to string, p *stubPreview) (string, error) {
	t.Helper()
	root, sc := repo(t)
	sc.values["stage"] = from
	var out, errb bytes.Buffer
	ctx := &command.Ctx{
		Stdout: &out, Stderr: &errb,
		Terminal: &command.FakeTerminal{}, Root: root, Adopted: true, Yes: true,
	}
	err := run(ctx, Deps{Scripts: sc.run, Files: vfs.OS{}, Preview: p.run}, []string{"stage", to})
	return out.String(), err
}

// Raising the stage shows what the new stage would require, in doctor's
// own marks, before the write (spec-0041).
func TestARaiseShowsWhatTheStageWouldRequire(t *testing.T) {
	p := &stubPreview{}
	out, err := raise(t, "1", "2", p)
	if err != nil {
		t.Fatalf("config = %v", err)
	}
	if len(p.asked) != 1 || p.asked[0] != 2 {
		t.Fatalf("doctor was asked for %v, want the target stage alone", p.asked)
	}
	for _, want := range []string{
		" stage 2 — pull requests. doctor previews what the stage would",
		"        require of this repository:",
		"         ✓  gh on the PATH",
		"         ✗  the recording push can write to main",
		"         ?  main is governed by a ruleset",
		"        3 of 6 met, 1 unmet, 2 unread. Declaring it writes",
		"        `stage: 2` and installs nothing; the flows that read the",
		"        forge will stop at the unmet one until it is answered.",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("the preview misses %q:\n%s", want, out)
		}
	}
}

// Nothing new is required of the repository when the stage is lowered or
// unchanged, so nothing is previewed (spec-0041, step 5).
func TestNoPreviewWhereTheStageIsLoweredOrUnchanged(t *testing.T) {
	for _, c := range []struct{ name, from, to string }{
		{"lowered", "3", "1"},
		{"unchanged", "2", "2"},
	} {
		t.Run(c.name, func(t *testing.T) {
			p := &stubPreview{}
			if _, err := raise(t, c.from, c.to, p); err != nil {
				t.Fatalf("config = %v", err)
			}
			if len(p.asked) != 0 {
				t.Errorf("doctor was asked for %v; nothing new is required", p.asked)
			}
		})
	}
}

// The preview is shown for the stage, and for no other key: a conduct
// flag asks nothing of the repository (spec-0041, scope).
func TestNoPreviewForAnyOtherKey(t *testing.T) {
	root, sc := repo(t)
	p := &stubPreview{}
	var out, errb bytes.Buffer
	ctx := &command.Ctx{Stdout: &out, Stderr: &errb,
		Terminal: &command.FakeTerminal{}, Root: root, Adopted: true, Yes: true}
	if err := run(ctx, Deps{Scripts: sc.run, Files: vfs.OS{}, Preview: p.run},
		[]string{"stage_2.pr_title_style", "conventional"}); err != nil {
		t.Fatalf("config = %v", err)
	}
	if len(p.asked) != 0 {
		t.Errorf("doctor was asked for %v on a title-style change", p.asked)
	}
	_ = out
}

// It never refuses. A preview that could not be run says so and the
// write is still offered — the stage is a declaration (spec-0041).
func TestAnUnreadablePreviewDoesNotBlockTheWrite(t *testing.T) {
	p := &stubPreview{fail: errors.New("the forge did not answer")}
	out, err := raise(t, "1", "2", p)
	if err != nil {
		t.Fatalf("config = %v; a preview that could not be run must not refuse a declaration", err)
	}
	if !strings.Contains(out, "the requirements could not be examined") {
		t.Errorf("the reader is not told why there is no preview:\n%s", out)
	}
	if !strings.Contains(out, "stage is 2.") {
		t.Errorf("the stage was not written:\n%s", out)
	}
}

// Without the port the raise asks as any other change does: a wiring
// with no doctor is not a wiring that refuses.
func TestARaiseWithNoPreviewPortStillWrites(t *testing.T) {
	root, sc := repo(t)
	sc.values["stage"] = "1"
	var out, errb bytes.Buffer
	ctx := &command.Ctx{Stdout: &out, Stderr: &errb,
		Terminal: &command.FakeTerminal{}, Root: root, Adopted: true, Yes: true}
	if err := run(ctx, Deps{Scripts: sc.run, Files: vfs.OS{}}, []string{"stage", "2"}); err != nil {
		t.Fatalf("config = %v", err)
	}
	if !strings.Contains(out.String(), "stage is 2.") {
		t.Errorf("the stage was not written:\n%s", out.String())
	}
}

// The frame is the assertion: the preview a stage raise prints is the
// block `config.excalidraw` draws under the `stage 1 → 2` row, line for
// line (docs/product/screens/README.md).
//
// The drawing's frame is a screen and this is a command's output, so
// what is compared is the block the two have in common: from the
// sentence that opens the preview to the one that closes it. The rows
// above it are the settings screen's, and the keys under it are the
// question's, which is `config`'s own footer and not this block's.
func TestThePreviewFrame(t *testing.T) {
	p := &stubPreview{}
	out, err := raise(t, "1", "2", p)
	if err != nil {
		t.Fatalf("config = %v", err)
	}
	drawn := previewBlock(t, frame(t, configDrawing, "writrun config — stage 1 → 2, previewed"))
	// This frame shortens one requirement's name to fit its column. The
	// rows are doctor's answers and the name is doctor's own, drawn in
	// full in `adoption/doctor.excalidraw` — so the shortening is the
	// canvas's, and the full name is what the binary prints.
	for i, l := range drawn {
		drawn[i] = strings.Replace(l, "refuses the push", "refuses the recording push", 1)
	}
	sameLines(t, previewBlock(t, rendered(out)), drawn)
}

// previewBlock is the preview and nothing around it: the lines from the
// one opening "stage 2 —" to the one closing the count's sentence.
func previewBlock(t *testing.T, lines []string) []string {
	t.Helper()
	first, last := -1, -1
	for i, l := range lines {
		if first < 0 && strings.HasPrefix(strings.TrimSpace(l), "stage 2 — pull requests.") {
			first = i
		}
		if strings.HasSuffix(strings.TrimSpace(l), "forge will stop at the unmet one until it is answered.") {
			last = i
		}
	}
	if first < 0 || last < 0 {
		t.Fatalf("no preview block in:\n%s", strings.Join(lines, "\n"))
	}
	block := make([]string, 0, last-first+1)
	for _, l := range lines[first : last+1] {
		// The drawing insets this frame's content by one column more
		// than it insets its own sibling; what the block says is what is
		// compared, not where a canvas put its left edge.
		block = append(block, strings.TrimRight(l, " "))
	}
	return block
}
