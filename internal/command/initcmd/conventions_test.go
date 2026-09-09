package initcmd

import (
	"reflect"
	"strings"
	"testing"

	"github.com/thomasfranke/writrun-cli/internal/gitx"

	"github.com/thomasfranke/writrun-cli/internal/kit"
	"github.com/thomasfranke/writrun-cli/internal/vfs"
)

func TestExtractVocabularyReadsTheHistory(t *testing.T) {
	target := makeTarget(t,
		"feat: begin",
		"fix(api): repair the thing",
		"feat(api): add the thing",
		"feat(cli): add another",
		"not a conventional subject",
	)
	v := extractVocabulary(vfs.OS{}, target, gitx.Run)
	if got, want := v.Types, []string{"feat", "fix"}; !reflect.DeepEqual(got, want) {
		t.Errorf("types = %v, want %v (frequency first)", got, want)
	}
	if got, want := v.Scopes, []string{"api", "cli"}; !reflect.DeepEqual(got, want) {
		t.Errorf("scopes = %v, want %v", got, want)
	}
	if v.Source != "the commit history" {
		t.Errorf("source = %q", v.Source)
	}
}

func TestExtractVocabularyReadsTheContributingGuide(t *testing.T) {
	target := makeTarget(t, "plain subject")
	write(t, target, "CONTRIBUTING.md", "Use `build(deps): bump things` and `test: cover it`.\n")
	v := extractVocabulary(vfs.OS{}, target, gitx.Run)
	if got, want := v.Types, []string{"build", "test"}; !reflect.DeepEqual(got, want) {
		t.Errorf("types = %v, want %v", got, want)
	}
	if v.Source != "the contributing guide" {
		t.Errorf("source = %q", v.Source)
	}
}

func TestExtractVocabularyMergesBothSources(t *testing.T) {
	target := makeTarget(t, "feat: begin")
	write(t, target, "docs/CONTRIBUTING.md", "Subjects look like `fix(core): mend`.\n")
	v := extractVocabulary(vfs.OS{}, target, gitx.Run)
	if len(v.Types) != 2 {
		t.Errorf("types = %v, want feat and fix", v.Types)
	}
	if v.Source != "the commit history and the contributing guide" {
		t.Errorf("source = %q", v.Source)
	}
}

func TestExtractVocabularyWithNeitherSourceIsEmpty(t *testing.T) {
	target := makeTarget(t, "initial import", "more work")
	v := extractVocabulary(vfs.OS{}, target, gitx.Run)
	if len(v.Types) != 0 || v.Source != "" {
		t.Errorf("vocabulary = %+v, want the zero value for shipped defaults", v)
	}
}

// applyTestKit lays the two files applyVocabulary rewrites into a
// bare directory, as the copy step would have.

// applyTestKit seeds the one file the vocabulary now lands in.
func applyTestKit(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	write(t, root, kit.Settings, templateSettings)
	return root
}

func TestApplyVocabularyWritesTheSettings(t *testing.T) {
	root := applyTestKit(t)
	v := vocabulary{Types: []string{"feat", "fix"}, Scopes: []string{"api"}, Source: "the commit history"}
	if err := applyVocabulary(vfs.OS{}, root, v); err != nil {
		t.Fatalf("applyVocabulary = %v", err)
	}
	got := read(t, root, kit.Settings)
	for _, want := range []string{`"commit_types": "feat fix"`, `"commit_scopes": "api"`} {
		if !strings.Contains(got, want) {
			t.Errorf("settings missing %s\n%s", want, got)
		}
	}
}

// The vocabulary has one home from WritRun v0.0.07. A second copy in a
// kit file is what a refresh used to revert, so nothing outside the
// settings may carry it.
func TestApplyVocabularyTouchesNoOtherFile(t *testing.T) {
	root := applyTestKit(t)
	write(t, root, "writrun/conventions/commits.md", templateCommits)
	write(t, root, ".writrun/scripts/stage-2-pull-requests/check_observance.sh", templateObservance)
	v := vocabulary{Types: []string{"feat"}, Source: "the commit history"}
	if err := applyVocabulary(vfs.OS{}, root, v); err != nil {
		t.Fatalf("applyVocabulary = %v", err)
	}
	if got := read(t, root, "writrun/conventions/commits.md"); got != templateCommits {
		t.Error("the convention was rewritten; it explains the vocabulary, it does not carry it")
	}
	if got := read(t, root, ".writrun/scripts/stage-2-pull-requests/check_observance.sh"); got != templateObservance {
		t.Error("a kit file was rewritten; a refresh would revert it")
	}
}

func TestApplyVocabularyKeepsShippedScopesWhenNoneObserved(t *testing.T) {
	root := applyTestKit(t)
	v := vocabulary{Types: []string{"feat"}, Source: "the commit history"}
	if err := applyVocabulary(vfs.OS{}, root, v); err != nil {
		t.Fatalf("applyVocabulary = %v", err)
	}
	got := read(t, root, kit.Settings)
	if !strings.Contains(got, `"commit_scopes": "about product"`) {
		t.Errorf("shipped scopes were not kept\n%s", got)
	}
}

func TestApplyVocabularyEmptyChangesNothing(t *testing.T) {
	root := applyTestKit(t)
	before := read(t, root, kit.Settings)
	if err := applyVocabulary(vfs.OS{}, root, vocabulary{}); err != nil {
		t.Fatalf("applyVocabulary = %v", err)
	}
	if read(t, root, kit.Settings) != before {
		t.Error("an empty vocabulary rewrote the shipped defaults")
	}
}

// A miss is loud: a silent no-op would report a vocabulary the file
// does not record.
func TestApplyVocabularyRefusesSettingsWithoutTheKey(t *testing.T) {
	root := t.TempDir()
	write(t, root, kit.Settings, "{\n  \"stage\": 1\n}\n")
	err := applyVocabulary(vfs.OS{}, root, vocabulary{Types: []string{"feat"}})
	if err == nil {
		t.Fatal("applyVocabulary accepted settings carrying no commit_types key")
	}
	if !strings.Contains(err.Error(), "commit_types") {
		t.Errorf("the refusal does not name the key: %v", err)
	}
}
