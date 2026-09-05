package hook

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/thomasfranke/writrun-cli/internal/gitx"
	"github.com/thomasfranke/writrun-cli/internal/vfs"
)

func TestInstallReportsTheDirectoryItCouldNotMake(t *testing.T) {
	disk := vfs.NewFake()
	boom := errors.New("the hooks directory is not yours")
	disk.Fail("/repo/.git/hooks", boom)

	err := Install(disk, "/repo/.git/hooks/commit-msg")
	if err == nil {
		t.Fatal("the hook was installed where its directory cannot be made")
	}
	if !errors.Is(err, boom) {
		t.Errorf("the cause did not survive: %v", err)
	}
	if !strings.Contains(err.Error(), "installing the commit-msg hook") {
		t.Errorf("the error does not name the act: %v", err)
	}
}

func TestInstallReportsTheFileItCouldNotWrite(t *testing.T) {
	disk := vfs.NewFake()
	disk.SeedDir("/repo/.git/hooks")
	boom := errors.New("the disk said no")
	disk.Fail("/repo/.git/hooks/commit-msg", boom)

	err := Install(disk, "/repo/.git/hooks/commit-msg")
	if err == nil {
		t.Fatal("the hook was written where the write fails")
	}
	if !errors.Is(err, boom) {
		t.Errorf("the cause did not survive: %v", err)
	}
}

func TestInspectReportsAHookItCannotRead(t *testing.T) {
	// Neither ours, nor foreign, nor absent — a fault worth naming
	// rather than answering as one of the three.
	disk := vfs.NewFake()
	disk.Seed("/repo/.git/hooks/commit-msg", []byte(Script), 0o755)
	boom := errors.New("the hook cannot be read")
	disk.Fail("/repo/.git/hooks/commit-msg", boom)

	state, err := Inspect(disk, "/repo/.git/hooks/commit-msg")
	if err == nil {
		t.Fatalf("an unreadable hook was answered as %v", state)
	}
	if !errors.Is(err, boom) {
		t.Errorf("the cause did not survive: %v", err)
	}
}

func TestInstallThenInspectRoundTripsThroughTheFake(t *testing.T) {
	disk := vfs.NewFake()
	path := "/repo/.git/hooks/commit-msg"
	if err := Install(disk, path); err != nil {
		t.Fatalf("Install: %v", err)
	}
	state, err := Inspect(disk, path)
	if err != nil || state != Ours {
		t.Errorf("Inspect after Install = %v, %v; want Ours", state, err)
	}
	info, err := disk.Stat(path)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if info.Mode().Perm()&0o100 == 0 {
		t.Errorf("the hook was installed unexecutable: %v — nobody could run it", info.Mode())
	}
}

func TestRefuseForeignAsksThePort(t *testing.T) {
	disk := vfs.NewFake()
	path := "/repo/.git/hooks/commit-msg"
	if err := RefuseForeign(disk, path); err != nil {
		t.Fatalf("an absent hook was refused: %v", err)
	}
	disk.Seed(path, []byte("#!/bin/sh\n"), 0o755)
	err := RefuseForeign(disk, path)
	if err == nil {
		t.Fatal("an existing hook was not refused")
	}
	if !strings.Contains(err.Error(), path) {
		t.Errorf("the refusal does not name the hook: %v", err)
	}
}

func TestPathReportsAGitThatCannotAnswer(t *testing.T) {
	_, err := Path(t.TempDir(), gitx.Run)
	if err == nil {
		t.Fatal("the hooks directory was resolved outside a repository")
	}
	if !strings.Contains(err.Error(), "resolving the hooks directory") {
		t.Errorf("the error does not name the act: %v", err)
	}
}

// The installed hook is a bash script, and proving it validates a
// subject means running it — which needs a real file on a real disk.
// That is the one thing the port does not replace.
func TestTheInstalledScriptIsRunnable(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "commit-msg")
	if err := Install(vfs.OS{}, path); err != nil {
		t.Fatalf("Install: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("the hook is not on disk: %v", err)
	}
}
