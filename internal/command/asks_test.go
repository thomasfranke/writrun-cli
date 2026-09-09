package command_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A command that declares it asks nothing must ask nothing.
//
// The declaration decides whether the screen captures the command's
// output and pages it, and a captured question waits forever on a
// reader who cannot see it. So the claim is checked against the source
// rather than trusted: a package that calls Ask* and also claims to ask
// nothing is named here, at the moment it is written.
func TestNoCommandBothAsksAndDeclaresItDoesNot(t *testing.T) {
	dirs, err := filepath.Glob("*cmd")
	if err != nil || len(dirs) == 0 {
		t.Fatalf("no command packages found: %v", err)
	}
	for _, dir := range dirs {
		files, _ := filepath.Glob(filepath.Join(dir, "*.go"))
		asks, declares := "", ""
		for _, f := range files {
			if strings.HasSuffix(f, "_test.go") {
				continue
			}
			b, err := os.ReadFile(f)
			if err != nil {
				t.Fatalf("reading %s: %v", f, err)
			}
			src := string(b)
			for _, ask := range []string{"AskInput(", "AskConfirm(", "AskSelect("} {
				if strings.Contains(src, ask) {
					asks = f
				}
			}
			if strings.Contains(src, "AsksNothing: true") {
				declares = f
			}
		}
		if asks != "" && declares != "" {
			t.Errorf("%s declares AsksNothing in %s but asks in %s — "+
				"its output would be captured and its question would wait unseen",
				dir, declares, asks)
		}
	}
}
