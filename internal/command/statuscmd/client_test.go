package statuscmd

import (
	"strings"
	"testing"

	"github.com/thomasfranke/writrun-cli/internal/command"
	"github.com/thomasfranke/writrun-cli/internal/drawing"
)

// The row names the binary that answered. Two machines answering
// differently is the question a version exists to settle, and until
// this row existed `status` could not be read for it (report-0034).
func TestTheClientRowNamesTheProductAndItsVersion(t *testing.T) {
	out, err := ask(t, fixture(), onBranch("main"), nil)
	if err != nil {
		t.Fatalf("run = %v", err)
	}
	wants(t, out, "Client", command.Product+" "+version)
}

// The version is the frame's, which is the string `--version` prints.
// Read anywhere else the two surfaces could answer differently about
// one process (spec-0043, step 2).
func TestTheVersionIsTheOneTheFrameAnswersWith(t *testing.T) {
	out, err := askVersion(t, "v9.9.9-rc1+dirty")
	if err != nil {
		t.Fatalf("run = %v", err)
	}
	wants(t, out, "Client", command.Product+" v9.9.9-rc1+dirty")
}

// A development build says so. `dev` is what a source build reports,
// and a build that is not a release must not be able to look like one.
func TestADevelopmentBuildPrintsWhatItReports(t *testing.T) {
	out, err := askVersion(t, "dev")
	if err != nil {
		t.Fatalf("run = %v", err)
	}
	wants(t, out, "Client", command.Product+" dev")
}

// Client and Kit are two facts, not one: the binary that answered, and
// the tag its scripts came from. The order is the drawing's.
func TestTheClientRowSitsAboveTheKitRow(t *testing.T) {
	out, err := ask(t, fixture(), onBranch("main"), nil)
	if err != nil {
		t.Fatalf("run = %v", err)
	}
	client, kit := strings.Index(out, "Client "), strings.Index(out, "Kit ")
	if client < 0 || kit < 0 {
		t.Fatalf("status =\n%s\nwant both rows", out)
	}
	if client > kit {
		t.Errorf("status =\n%s\nwant the client named above the kit", out)
	}
}

// The answer is the frame the drawing states, row for row
// (docs/product/screens/tasks/status.excalidraw). Where the two
// disagree the binary is what changes.
func TestTheAnswerIsTheFrameTheDrawingStates(t *testing.T) {
	const caption = "writrun status — the rows explained, and the client named"
	want, err := drawing.Frame("../../../docs/product/screens/tasks/status.excalidraw", caption)
	if err != nil {
		t.Fatalf("%v", err)
	}

	// The fixture answers the drawing's own facts: a frame states the
	// client and the kit it was drawn against, and a fixture spelling
	// its own would be a second opinion about what the rows say.
	f := fixture()
	f.Seed(root+"/.writrun/VERSION", []byte(drawnTag(t, want)+"\n"), 0o644)
	f.Seed(root+"/work/reports/report-0001-fixture.md", []byte(report("report-0001", "open")), 0o644)

	out, err := askAll(t, f, onBranch("task/0014-status-command"), drawnVersion(t, want), drawnTag(t, want))
	if err != nil {
		t.Fatalf("run = %v", err)
	}

	got := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(got) != len(want) {
		t.Fatalf("status printed %d rows and the frame draws %d:\n%s", len(got), len(want), out)
	}
	for i := range want {
		// A drawn row stops where its window does, so the frame's row
		// is what the printed one opens with.
		if !strings.HasPrefix(got[i], want[i]) {
			t.Errorf("row %d:\n  binary %q\n  frame  %q", i+1, got[i], want[i])
		}
	}
}

// drawnVersion and drawnTag read the two facts the frame names out of
// the frame itself.
func drawnVersion(t *testing.T, frame []string) string {
	t.Helper()
	return strings.TrimPrefix(drawnRow(t, frame, "Client"), command.Product+" ")
}

func drawnTag(t *testing.T, frame []string) string {
	t.Helper()
	fields := strings.Fields(drawnRow(t, frame, "Kit"))
	if len(fields) < 2 {
		t.Fatalf("the frame's Kit row names no tag: %q", drawnRow(t, frame, "Kit"))
	}
	return fields[1]
}

func drawnRow(t *testing.T, frame []string, label string) string {
	t.Helper()
	for _, l := range frame {
		if strings.HasPrefix(l, label+" ") {
			return strings.TrimSpace(strings.TrimPrefix(l, label))
		}
	}
	t.Fatalf("the frame draws no %s row: %q", label, frame)
	return ""
}
