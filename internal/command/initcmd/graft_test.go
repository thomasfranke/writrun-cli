package initcmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestGraftSectionCutsHeadingThroughEndMarker(t *testing.T) {
	section, err := graftSection([]byte(templateAgents))
	if err != nil {
		t.Fatalf("graftSection = %v", err)
	}
	s := string(section)
	if !strings.HasPrefix(s, sectionHeading) {
		t.Errorf("section starts %q, want the heading", s[:40])
	}
	if !strings.HasSuffix(s, markerEnd) {
		t.Errorf("section ends %q, want the end marker", s[len(s)-40:])
	}
	if strings.Contains(s, "TODO: one paragraph") {
		t.Error("the section carried text from outside the fence")
	}
}

func TestGraftSectionRefusesATemplateWithoutMarkers(t *testing.T) {
	if _, err := graftSection([]byte("# AGENTS.md\n\nno fence here\n")); err == nil {
		t.Fatal("graftSection accepted a template with no fenced section")
	}
}

func TestGraftKeepsEveryExistingByte(t *testing.T) {
	cases := []struct {
		name     string
		existing string
	}{
		{"trailing newline", "# Mine\n\nMy rules.\n"},
		{"no trailing newline", "# Mine\n\nMy rules."},
	}
	section := []byte("## WritRun\n<!-- writrun:begin -->\nflow\n" + markerEnd)
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := graft([]byte(tc.existing), section)
			if !bytes.HasPrefix(out, []byte(tc.existing)) {
				t.Error("bytes before the graft changed")
			}
			if !bytes.Contains(out, section) {
				t.Error("the section is not in the grafted result")
			}
			if !bytes.HasSuffix(out, []byte("\n")) {
				t.Error("the grafted file lost its final newline")
			}
		})
	}
}
