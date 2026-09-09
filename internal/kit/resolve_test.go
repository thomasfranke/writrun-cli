package kit_test

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/thomasfranke/writrun-cli/internal/kit"
)

// runner answers one call, so a case says what the resolver said and
// what it exited with, and nothing else.
type runner struct {
	said, complained string
	verdict          error
	ran              []string
}

func (r *runner) run(_ string, stdout, stderr io.Writer, _ []string, script string, args ...string) error {
	r.ran = append(r.ran, strings.TrimSpace(script+" "+strings.Join(args, " ")))
	fmt.Fprint(stdout, r.said)
	fmt.Fprint(stderr, r.complained)
	return r.verdict
}

func TestResolveAnswersThePathInForce(t *testing.T) {
	r := &runner{said: "writrun/gates.md\n"}
	got, err := kit.Resolve(r.run, "/repo", kit.Gates)
	if err != nil {
		t.Fatalf("Resolve = %v", err)
	}
	if got != "writrun/gates.md" {
		t.Errorf("Resolve = %q, want the project's own file", got)
	}
	if len(r.ran) != 1 || !strings.HasPrefix(r.ran[0], kit.ResolveDoc) {
		t.Errorf("the kit's resolver was not the one asked: %v", r.ran)
	}
	if !strings.HasSuffix(r.ran[0], kit.Gates) {
		t.Errorf("the address was not passed through: %q", r.ran[0])
	}
}

// The address is the project's either way. That the answer came from the
// kit's home is the resolver's business, not the caller's.
func TestResolveAnswersTheDefaultWhereTheFileDefers(t *testing.T) {
	r := &runner{said: ".writrun/defaults/gates.md\n"}
	got, err := kit.Resolve(r.run, "/repo", kit.Gates)
	if err != nil {
		t.Fatalf("Resolve = %v", err)
	}
	if got != ".writrun/defaults/gates.md" {
		t.Errorf("Resolve = %q, want the kit's default", got)
	}
}

// The script's own words reach the user: it knows why it refused and
// this function does not.
func TestResolveCarriesTheScriptsRefusal(t *testing.T) {
	r := &runner{complained: "no default ships for writrun/nope.md\n", verdict: errors.New("exit status 3")}
	_, err := kit.Resolve(r.run, "/repo", "writrun/nope.md")
	if err == nil {
		t.Fatal("Resolve accepted a refusal")
	}
	if !strings.Contains(err.Error(), "no default ships") {
		t.Errorf("the script's own words did not reach the caller: %v", err)
	}
}

// A runner that failed before the script spoke has no words to pass on,
// so the cause travels instead.
func TestResolveCarriesTheRunnersFailure(t *testing.T) {
	cause := errors.New("bash: not found")
	r := &runner{verdict: cause}
	_, err := kit.Resolve(r.run, "/repo", kit.Gates)
	if !errors.Is(err, cause) {
		t.Errorf("the cause did not travel: %v", err)
	}
}

// Exit zero and nothing said is not an answer. Returning "" would have
// the caller read the repository root as a file.
func TestResolveRefusesAnEmptyAnswer(t *testing.T) {
	r := &runner{said: "  \n"}
	_, err := kit.Resolve(r.run, "/repo", kit.Gates)
	if err == nil {
		t.Fatal("Resolve accepted an empty answer")
	}
	if !strings.Contains(err.Error(), "named no file") {
		t.Errorf("the refusal does not say what was wrong: %v", err)
	}
}
