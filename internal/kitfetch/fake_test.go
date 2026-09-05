package kitfetch

import (
	"errors"
	"strings"
	"testing"
)

func TestTheFakeHandsOverATemplateAndItsCleanup(t *testing.T) {
	f := NewFake("/kit/template")
	got, err := f.Fetch(tag, "the source")
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if got.Template != "/kit/template" {
		t.Errorf("Template = %q, want /kit/template", got.Template)
	}
	if got.Cleanup == nil {
		t.Fatal("the fake handed back no cleanup — the real fetch would leak and no test would see it")
	}
	got.Cleanup()
	if f.Cleaned != 1 {
		t.Errorf("Cleaned = %d, want 1", f.Cleaned)
	}
	if len(f.Asked) != 1 || f.Asked[0] != (Ask{Tag: tag, Source: "the source"}) {
		t.Errorf("Asked = %v, want one ask for %s from the source", f.Asked, tag)
	}
}

func TestTheFakeRefusesTheTagItWasToldFails(t *testing.T) {
	f := NewFake("/kit/template")
	f.Fail(tag, errors.New("repository not found"))
	_, err := f.Fetch(tag, "the source")
	if err == nil {
		t.Fatal("a tag told to fail was fetched")
	}
	for _, want := range []string{tag, "the source", "nothing was written", "repository not found"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal does not name %q: %v", want, err)
		}
	}

	f.Heal(tag)
	if _, err := f.Fetch(tag, "the source"); err != nil {
		t.Errorf("a healed tag still failed: %v", err)
	}
}

func TestTheFakeAnswersARepositoryThatIsNotAWritRunOne(t *testing.T) {
	// The real fetch reaches this refusal by cloning and finding no
	// template/; the fake reaches it by being told, and both say it in
	// the same words.
	f := NewFake("/kit/template")
	f.FailNoTemplate(tag)
	_, err := f.Fetch(tag, "the source")
	if err == nil {
		t.Fatal("a source with no template/ was accepted")
	}
	if err.Error() != errNoTemplate(tag, "the source").Error() {
		t.Errorf("the fake's refusal is not the real one: %v", err)
	}
}

func TestTheFakeOnlyRefusesTheTagItWasGiven(t *testing.T) {
	f := NewFake("/kit/template")
	f.Fail("v0.0.1", errors.New("no"))
	if _, err := f.Fetch(tag, "the source"); err != nil {
		t.Errorf("a tag nobody failed was refused: %v", err)
	}
}
