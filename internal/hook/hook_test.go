package hook

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInspectRecognisesOnlyTheKitsOwnHook(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "commit-msg")

	if state, err := Inspect(path); err != nil || state != Absent {
		t.Errorf("nothing installed: state = %v, err = %v; want Absent", state, err)
	}

	if err := Install(path); err != nil {
		t.Fatalf("Install: %v", err)
	}
	if state, err := Inspect(path); err != nil || state != Ours {
		t.Errorf("the hook Install wrote: state = %v, err = %v; want Ours", state, err)
	}

	// One byte's difference is a hook a project owns.
	if err := os.WriteFile(path, []byte(Script+"# edited\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if state, err := Inspect(path); err != nil || state != Foreign {
		t.Errorf("an edited hook: state = %v, err = %v; want Foreign", state, err)
	}

	if err := os.WriteFile(path, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if state, err := Inspect(path); err != nil || state != Foreign {
		t.Errorf("somebody else's hook: state = %v, err = %v; want Foreign", state, err)
	}
}

func TestInstallCreatesTheDirectoryAndTheModeBit(t *testing.T) {
	path := filepath.Join(t.TempDir(), "hooks", "commit-msg")
	if err := Install(path); err != nil {
		t.Fatalf("Install: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("the hook was not written: %v", err)
	}
	if info.Mode().Perm()&0o100 == 0 {
		t.Errorf("the hook is not executable: %v", info.Mode())
	}
}

func TestRefuseForeignNamesThePath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "commit-msg")
	if err := RefuseForeign(path); err != nil {
		t.Fatalf("an absent hook was refused: %v", err)
	}
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	err := RefuseForeign(path)
	if err == nil {
		t.Fatal("an existing hook was not refused")
	}
	if !strings.Contains(err.Error(), path) {
		t.Errorf("the refusal does not name %q: %v", path, err)
	}
}
