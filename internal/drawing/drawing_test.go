package drawing

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// at is one drawn text, placed the way a terminal frame places it: a
// row and a column inside the window.
type at struct {
	row, col int
	text     string
}

// drawn writes one drawing — a window, its title bar with the caption,
// an optional divider at a row, and the texts — and returns its path.
func drawn(t *testing.T, caption string, divider int, cells ...at) string {
	t.Helper()
	els := []element{
		{Type: "rectangle", X: 0, Y: 0, Width: 600, Height: 500, Background: windowFill},
		{Type: "rectangle", X: 0, Y: 0, Width: 600, Height: 30, Background: barFill},
		{Type: "text", X: 78, Y: 6, Text: caption},
	}
	if divider > 0 {
		els = append(els, element{
			Type: "rectangle", X: 8, Y: y(divider), Width: 590, Height: 1, Background: dividerFill,
		})
	}
	for _, c := range cells {
		els = append(els, element{Type: "text", X: x(c.col), Y: y(c.row), OriginalText: c.text})
	}
	b, err := json.Marshal(file{Elements: els})
	if err != nil {
		t.Fatalf("marshal = %v", err)
	}
	path := filepath.Join(t.TempDir(), "case.excalidraw")
	if err := os.WriteFile(path, b, 0o644); err != nil {
		t.Fatalf("write = %v", err)
	}
	return path
}

// y and x are the canvas positions of a row and a column inside the
// window, the title bar's 30 pixels included.
func y(row int) float64 { return 30 + 6 + float64(row)*lineHeight }
func x(col int) float64 { return 8 + float64(col)*charWidth }

func frame(t *testing.T, path, caption string) []string {
	t.Helper()
	lines, err := Frame(path, caption)
	if err != nil {
		t.Fatalf("Frame = %v", err)
	}
	return lines
}

func TestTheFramesLinesAreWhatTheBinaryWouldPrint(t *testing.T) {
	// The prompt is the reader typing and is dropped; every output line
	// carries one column of frame padding and loses it; a gap between
	// two drawn lines is the blank line between them; and a line drawn
	// in two texts is one line.
	path := drawn(t, "writrun status — a frame", 0,
		at{0, 0, "$ writrun status"},
		at{1, 1, "Branch"},
		at{1, 10, "main"},
		at{3, 1, "Kit      WritRun v0.0.09"},
	)
	want := []string{
		"Branch   main",
		"",
		"Kit      WritRun v0.0.09",
	}
	if got := frame(t, path, "writrun status — a frame"); !reflect.DeepEqual(got, want) {
		t.Fatalf("frame =\n%q\nwant\n%q", got, want)
	}
}

func TestWhatIsDrawnBelowTheDividerIsNotThisRunsOutput(t *testing.T) {
	// A frame's divider separates the run from the annotation beside it
	// or from a second scene; neither is what the binary printed here.
	path := drawn(t, "a frame", 5,
		at{0, 0, "$ writrun --version"},
		at{1, 1, "writrun-cli v0.0.2 (pins WritRun v0.0.09)"},
		at{6, 1, "Two facts, not one: the binary that answered."},
	)
	want := []string{"writrun-cli v0.0.2 (pins WritRun v0.0.09)"}
	if got := frame(t, path, "a frame"); !reflect.DeepEqual(got, want) {
		t.Fatalf("frame =\n%q\nwant\n%q", got, want)
	}
}

func TestAnAuthorsOwnLineBreaksAreLinesOfTheirOwn(t *testing.T) {
	path := drawn(t, "a frame", 0, at{0, 1, "first\nsecond"})
	want := []string{"first", "second"}
	if got := frame(t, path, "a frame"); !reflect.DeepEqual(got, want) {
		t.Fatalf("frame =\n%q\nwant\n%q", got, want)
	}
}

func TestACaptionNoFrameCarriesIsNamedRatherThanAnsweredEmpty(t *testing.T) {
	path := drawn(t, "a frame", 0, at{0, 1, "something"})
	_, err := Frame(path, "another frame")
	if err == nil {
		t.Fatal("Frame = nil error; want the missing caption named")
	}
	if !strings.Contains(err.Error(), "another frame") {
		t.Fatalf("err = %v; want the caption named", err)
	}
}

func TestADrawingThatCannotBeReadIsAnError(t *testing.T) {
	if _, err := Frame(filepath.Join(t.TempDir(), "absent.excalidraw"), "a frame"); err == nil {
		t.Fatal("Frame = nil error; want the read failure")
	}
	path := filepath.Join(t.TempDir(), "broken.excalidraw")
	if err := os.WriteFile(path, []byte("{not json"), 0o644); err != nil {
		t.Fatalf("write = %v", err)
	}
	if _, err := Frame(path, "a frame"); err == nil {
		t.Fatal("Frame = nil error; want the parse failure")
	}
}
