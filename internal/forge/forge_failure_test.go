package forge

import (
	"strings"
	"testing"
)

// Bin is the seam: pointing it at bash is how a case shapes exactly
// what gh's two streams would carry on a failure.
//
// The markers are computed by the shell so that neither appears in the
// command line itself — the invocation is echoed into the error, and a
// literal marker there would be found before the output's own.
const (
	onStdout   = "42"
	onStderr   = "100"
	twoStreams = "echo $((20+22)); echo $((50+50)) >&2; exit 3"
)

func TestAFailureJoinsBothStreams(t *testing.T) {
	_, err := Client{Bin: "bash"}.Run("-c", twoStreams)
	if err == nil {
		t.Fatal("a failing invocation succeeded")
	}
	got := err.Error()
	if !strings.Contains(got, onStderr) {
		t.Errorf("stderr did not reach the error: %v", err)
	}
	if !strings.Contains(got, onStdout) {
		t.Errorf("stdout did not reach the error: %v", err)
	}
	// stderr first, then stdout — some subcommands put the API's own
	// detail on stdout, and it reads as the elaboration it is.
	if strings.Index(got, onStderr) > strings.Index(got, onStdout) {
		t.Errorf("stdout came before stderr: %v", err)
	}
}

func TestAFailureWithOnlyStdoutCarriesIt(t *testing.T) {
	_, err := Client{Bin: "bash"}.Run("-c", "echo $((20+22)); exit 3")
	if err == nil {
		t.Fatal("a failing invocation succeeded")
	}
	if !strings.Contains(err.Error(), onStdout) {
		t.Errorf("stdout did not reach the error: %v", err)
	}
}

func TestAFailureThatSaidNothingStillNamesItself(t *testing.T) {
	_, err := Client{Bin: "bash"}.Run("-c", "exit 3")
	if err == nil {
		t.Fatal("a failing invocation succeeded")
	}
	if !strings.Contains(err.Error(), "gh -c exit 3") {
		t.Errorf("the error does not name the invocation: %v", err)
	}
}

func TestTheZeroValueReachesForGhOnPath(t *testing.T) {
	// Whatever the machine answers, what this proves is that an empty
	// Bin resolves to "gh" rather than to the empty string.
	_, err := Client{}.Run("--no-such-flag-anywhere")
	if err == nil {
		t.Skip("this machine's gh accepted an unknown flag")
	}
	if !strings.Contains(err.Error(), "gh --no-such-flag-anywhere") {
		t.Errorf("the zero value did not resolve to gh: %v", err)
	}
}

func TestStdoutIsReturnedOnSuccess(t *testing.T) {
	out, err := Client{Bin: "bash"}.Run("-c", "echo answered")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if strings.TrimSpace(out) != "answered" {
		t.Errorf("Run = %q, want answered", out)
	}
}
