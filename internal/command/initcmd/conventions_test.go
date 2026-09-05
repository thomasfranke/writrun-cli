package initcmd

import (
	"reflect"
	"strings"
	"testing"

	"github.com/thomasfranke/writrun-cli/internal/gitx"
)

func TestExtractVocabularyReadsTheHistory(t *testing.T) {
	target := makeTarget(t,
		"feat: begin",
		"fix(api): repair the thing",
		"feat(api): add the thing",
		"feat(cli): add another",
		"not a conventional subject",
	)
	v := extractVocabulary(target, gitx.Run)
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
	v := extractVocabulary(target, gitx.Run)
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
	v := extractVocabulary(target, gitx.Run)
	if len(v.Types) != 2 {
		t.Errorf("types = %v, want feat and fix", v.Types)
	}
	if v.Source != "the commit history and the contributing guide" {
		t.Errorf("source = %q", v.Source)
	}
}

func TestExtractVocabularyWithNeitherSourceIsEmpty(t *testing.T) {
	target := makeTarget(t, "initial import", "more work")
	v := extractVocabulary(target, gitx.Run)
	if len(v.Types) != 0 || v.Source != "" {
		t.Errorf("vocabulary = %+v, want the zero value for shipped defaults", v)
	}
}

// applyTestKit lays the two files applyVocabulary rewrites into a
// bare directory, as the copy step would have.
func applyTestKit(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	write(t, root, ".writrun/conventions/commits.md", templateCommits)
	write(t, root, ".writrun/scripts/stage-2-pull-requests/check_observance.sh", templateObservance)
	return root
}

func TestApplyVocabularyRewritesBothHalves(t *testing.T) {
	root := applyTestKit(t)
	v := vocabulary{Types: []string{"feat", "fix"}, Scopes: []string{"api"}, Source: "the commit history"}
	if err := applyVocabulary(root, v); err != nil {
		t.Fatalf("applyVocabulary = %v", err)
	}
	commits := read(t, root, ".writrun/conventions/commits.md")
	if !strings.Contains(commits, "- **Types**: `feat`, `fix`.") {
		t.Errorf("commits.md types not rewritten:\n%s", commits)
	}
	if !strings.Contains(commits, "`api`.") || strings.Contains(commits, "`about`") {
		t.Errorf("commits.md scopes not rewritten:\n%s", commits)
	}
	observance := read(t, root, ".writrun/scripts/stage-2-pull-requests/check_observance.sh")
	if !strings.Contains(observance, `TYPES="feat fix"`) || !strings.Contains(observance, `SCOPES="api"`) {
		t.Errorf("check_observance.sh not rewritten:\n%s", observance)
	}
}

func TestApplyVocabularyKeepsShippedScopesWhenNoneObserved(t *testing.T) {
	root := applyTestKit(t)
	if err := applyVocabulary(root, vocabulary{Types: []string{"fix"}}); err != nil {
		t.Fatalf("applyVocabulary = %v", err)
	}
	observance := read(t, root, ".writrun/scripts/stage-2-pull-requests/check_observance.sh")
	if !strings.Contains(observance, `SCOPES="about product technical"`) {
		t.Errorf("shipped scopes did not survive:\n%s", observance)
	}
	commits := read(t, root, ".writrun/conventions/commits.md")
	if !strings.Contains(commits, "`about`, `product`, `technical`") {
		t.Errorf("shipped scopes bullet did not survive:\n%s", commits)
	}
}

func TestApplyVocabularyEmptyChangesNothing(t *testing.T) {
	root := applyTestKit(t)
	before := read(t, root, ".writrun/conventions/commits.md")
	if err := applyVocabulary(root, vocabulary{}); err != nil {
		t.Fatalf("applyVocabulary = %v", err)
	}
	if read(t, root, ".writrun/conventions/commits.md") != before {
		t.Error("an empty vocabulary rewrote the shipped defaults")
	}
}

func TestReplaceBulletSwallowsContinuationLines(t *testing.T) {
	in := "- **Scopes** (optional): `a`,\n  `b`, `c`.\n- Example: `x: y`.\n"
	out := replaceBullet(in, "- **Scopes**", "- **Scopes**: `z`.")
	want := "- **Scopes**: `z`.\n- Example: `x: y`.\n"
	if out != want {
		t.Errorf("replaceBullet = %q, want %q", out, want)
	}
}

func TestApplyVocabularyRespellsTheExample(t *testing.T) {
	root := applyTestKit(t)
	v := vocabulary{Types: []string{"feat", "fix"}, Scopes: []string{"api"}, Source: "the commit history"}
	if err := applyVocabulary(root, v); err != nil {
		t.Fatalf("applyVocabulary = %v", err)
	}
	commits := read(t, root, ".writrun/conventions/commits.md")
	// The summary is the kit's prose and stays; the vocabulary it
	// spells is the project's, and a shipped one would be a subject the
	// hook installed by the same run refuses.
	if !strings.Contains(commits, "- Example: `feat(api): add a chapter`.") {
		t.Errorf("the example was not respelled:\n%s", commits)
	}
}

func TestApplyVocabularyExampleKeepsTheShippedScope(t *testing.T) {
	root := applyTestKit(t)
	if err := applyVocabulary(root, vocabulary{Types: []string{"fix"}}); err != nil {
		t.Fatalf("applyVocabulary = %v", err)
	}
	commits := read(t, root, ".writrun/conventions/commits.md")
	if !strings.Contains(commits, "- Example: `fix(product): add a chapter`.") {
		t.Errorf("the example dropped the shipped scope:\n%s", commits)
	}
}
