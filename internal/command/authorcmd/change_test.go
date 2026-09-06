package authorcmd

import (
	"strings"
	"testing"
)

func TestNormalizeIsTheBranchAlphabet(t *testing.T) {
	cases := []struct{ in, want string }{
		{"Derived work", "derived-work"},
		{"  The Merge Is The Assenting Act ", "the-merge-is"},
		{"pull_requests/shape", "pull-requests-shape"},
		{"---", ""},
		{"", ""},
		{"already-fine", "already-fine"},
	}
	for _, c := range cases {
		if got := normalize(c.in); got != c.want {
			t.Errorf("normalize(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestSubjectOfReadsTheDocTheRuleWasWrittenInto(t *testing.T) {
	cases := []struct{ in, want string }{
		{"docs/product/pull-requests/author.md", "author"},
		{"docs/product/README.md", "product"},
		{"docs/technical/engineering/README.md", "engineering"},
		{"README.md", "readme"},
	}
	for _, c := range cases {
		if got := subjectOf(c.in); got != c.want {
			t.Errorf("subjectOf(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestBranchNameIsDocsShortName(t *testing.T) {
	cases := []struct {
		name   string
		ch     change
		slug   string
		want   string
		errish string
	}{
		{
			name: "the slug wins over everything",
			ch:   change{branch: "docs/kept", docs: []string{"docs/product/rules.md"}},
			slug: "Derived Work",
			want: "docs/derived-work",
		},
		{
			name: "an authoring branch keeps its own name",
			ch:   change{branch: "docs/the-merge", docs: []string{"docs/product/rules.md"}},
			want: "docs/the-merge",
		},
		{
			name: "otherwise the doc names the subject",
			ch:   change{branch: "scratch", docs: []string{"docs/product/pull-requests/author.md"}},
			want: "docs/author",
		},
		{
			name:   "a slug that is no words at all",
			ch:     change{branch: "scratch", docs: []string{"docs/product/rules.md"}},
			slug:   "///",
			errish: "leaves no subject words",
		},
		{
			name:   "a doc that names nothing",
			ch:     change{branch: "scratch", docs: []string{"docs/---.md"}},
			errish: "name them with --slug",
		},
		{
			name: "a bare docs/ branch is not a name",
			ch:   change{branch: "docs/", docs: []string{"docs/product/rules.md"}},
			want: "docs/rules",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := branchName(c.ch, c.slug)
			if c.errish != "" {
				if err == nil || !strings.Contains(err.Error(), c.errish) {
					t.Fatalf("err = %v, want it to say %q", err, c.errish)
				}
				return
			}
			if err != nil {
				t.Fatalf("branchName: %v", err)
			}
			if got != c.want {
				t.Errorf("branchName = %q, want %q", got, c.want)
			}
		})
	}
}

func TestReadChangeReadsTheDiffAndItsBranch(t *testing.T) {
	g := &fakeGit{
		branch: "docs/authoring",
		refs:   map[string]bool{"refs/remotes/origin/main": true},
		files:  []string{"docs/product/rules.md", newTaskPath, "work/reports/report-0009-a-thing.md"},
	}
	ch, err := readChange(g.run, root, "")
	if err != nil {
		t.Fatalf("readChange: %v", err)
	}
	if ch.rng != "origin/main...HEAD" {
		t.Errorf("range = %q, want origin/main...HEAD", ch.rng)
	}
	if len(ch.docs) != 1 || ch.docs[0] != "docs/product/rules.md" {
		t.Errorf("docs = %v, want the one rule", ch.docs)
	}
	if len(ch.files) != 3 {
		t.Errorf("files = %v, want all three", ch.files)
	}
}

// A report rides whatever change is already open, so `work/reports/` is
// not a stray (AGENTS.md, recording what you noticed).
func TestAReportRidesTheAuthoringChange(t *testing.T) {
	g := &fakeGit{
		branch: "docs/authoring",
		refs:   map[string]bool{"refs/heads/main": true},
		files:  []string{"docs/product/rules.md", "work/reports/report-0009-a-thing.md"},
	}
	ch, err := readChange(g.run, root, "")
	if err != nil {
		t.Fatalf("readChange: %v", err)
	}
	if ch.rng != "main...HEAD" {
		t.Errorf("range = %q, want the local main", ch.rng)
	}
}

func TestAGivenRangeIsUsedAsGiven(t *testing.T) {
	g := &fakeGit{branch: "docs/authoring", files: []string{"docs/product/rules.md"}}
	ch, err := readChange(g.run, root, "upstream/main...HEAD")
	if err != nil {
		t.Fatalf("readChange: %v", err)
	}
	if ch.rng != "upstream/main...HEAD" {
		t.Errorf("range = %q, want the one it was given", ch.rng)
	}
	for _, c := range g.calls {
		if strings.HasPrefix(c, "rev-parse --verify") {
			t.Errorf("a base was resolved although one was given: %v", g.calls)
		}
	}
}

func TestTrimmedLinesDropsTheBlanks(t *testing.T) {
	got := trimmedLines("a\n\n  b  \n\n")
	if strings.Join(got, ",") != "a,b" {
		t.Errorf("trimmedLines = %v", got)
	}
}

func TestHeadCaps(t *testing.T) {
	if got := head([]string{"a", "b", "c"}, 2); len(got) != 2 {
		t.Errorf("head = %v", got)
	}
	if got := head([]string{"a"}, 2); len(got) != 1 {
		t.Errorf("head = %v", got)
	}
}
