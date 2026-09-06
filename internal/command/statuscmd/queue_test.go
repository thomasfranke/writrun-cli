package statuscmd

import (
	"reflect"
	"testing"
)

func TestIdNumberReadsWhatTheQueueSpells(t *testing.T) {
	for _, tc := range []struct {
		name, prefix string
		want         int
		ok           bool
	}{
		{"task-0014-status-command", "task-", 14, true},
		{"task-14", "task-", 14, true},
		{"spec-0013-status", "spec-", 13, true},
		{"README", "task-", 0, false},
		{"task-", "task-", 0, false},
		{"task-abc", "task-", 0, false},
		{"spec-0013", "task-", 0, false},
	} {
		got, ok := idNumber(tc.name, tc.prefix)
		if ok != tc.ok || got != tc.want {
			t.Errorf("idNumber(%q, %q) = %d, %v; want %d, %v", tc.name, tc.prefix, got, ok, tc.want, tc.ok)
		}
	}
}

func TestListReadsAFrontMatterList(t *testing.T) {
	for _, tc := range []struct {
		raw  string
		want []string
	}{
		{"[spec-0013]", []string{"spec-0013"}},
		{"[spec-0013, spec-0021]", []string{"spec-0013", "spec-0021"}},
		{"[]", nil},
		{"null", nil},
		{"", nil},
		{"spec-0013", []string{"spec-0013"}},
	} {
		if got := list(tc.raw); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("list(%q) = %v; want %v", tc.raw, got, tc.want)
		}
	}
}

func TestFrontMatterReadsOnlyTheLeadingBlock(t *testing.T) {
	fm := frontMatter([]byte(taskFile))
	if fm["id"] != "task-0014" || fm["status"] != "in-progress" {
		t.Errorf("front matter = %v", fm)
	}
	if _, there := fm["Some body."]; there {
		t.Error("the body was read as front matter")
	}
	if got := frontMatter([]byte("# A README\n\nstatus: open\n")); len(got) != 0 {
		t.Errorf("a file without front matter yielded %v", got)
	}
	if got := frontMatter([]byte("---\nid: task-0014\n")); got["id"] != "task-0014" {
		t.Errorf("an unclosed block yielded %v", got)
	}
}

func TestHeadingIsTheFilesTitleOrNothing(t *testing.T) {
	if got := heading([]byte(taskFile)); got != "Answer where the work stands from the current branch" {
		t.Errorf("heading = %q", got)
	}
	if got := heading([]byte("---\nid: task-0014\n---\n\nno title\n")); got != "" {
		t.Errorf("heading = %q; want nothing where there is no title", got)
	}
}

func TestValueNamesWhatTheFileDoesNotRecord(t *testing.T) {
	if got := value("approved", "no status"); got != "approved" {
		t.Errorf("value = %q", got)
	}
	if got := value("null", "no status"); got != "no status" {
		t.Errorf("value = %q; want the absence named", got)
	}
	if got := value("  ", "no status"); got != "no status" {
		t.Errorf("value = %q; want the absence named", got)
	}
}
