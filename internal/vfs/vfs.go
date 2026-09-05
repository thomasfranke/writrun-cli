// Package vfs is the filesystem port: the calls the commands make,
// named once, with the real one and a fake behind the same interface.
//
// boundaries.md puts everything leaving the process behind a small
// interface with a fake beside it. Script execution, `gh` and the
// terminal were; this is the fourth, and the one whose absence made a
// failing write untestable without changing a file's mode.
package vfs

import (
	"io/fs"
)

// FS is the filesystem as this binary uses it — nothing wider. Every
// method is one the commands already call; a method nobody calls is a
// method the fake has to answer for no reason.
type FS interface {
	ReadFile(name string) ([]byte, error)
	WriteFile(name string, data []byte, perm fs.FileMode) error
	Stat(name string) (fs.FileInfo, error)
	MkdirAll(path string, perm fs.FileMode) error
	Remove(name string) error
	RemoveAll(path string) error
	// WalkDir walks the tree rooted at root. The port owns the walk
	// because the walk reads the filesystem: one that called
	// filepath.WalkDir directly would step outside the port on every
	// entry, and the fake could not answer it.
	WalkDir(root string, fn fs.WalkDirFunc) error
	// MkdirTemp makes a directory outside the repository — where the
	// kit is fetched, so a failure has written nothing where it
	// matters.
	MkdirTemp(dir, pattern string) (string, error)
}
