package vfs

import (
	"io/fs"
	"os"
	"path/filepath"
)

// OS is the production filesystem: every method is the `os` call it
// names, so the port costs nothing at runtime and the indirection is
// only what the tests need.
type OS struct{}

func (OS) ReadFile(name string) ([]byte, error) { return os.ReadFile(name) }

func (OS) WriteFile(name string, data []byte, perm fs.FileMode) error {
	return os.WriteFile(name, data, perm)
}

func (OS) Stat(name string) (fs.FileInfo, error) { return os.Stat(name) }

func (OS) MkdirAll(path string, perm fs.FileMode) error { return os.MkdirAll(path, perm) }

func (OS) Remove(name string) error { return os.Remove(name) }

func (OS) RemoveAll(path string) error { return os.RemoveAll(path) }

func (OS) WalkDir(root string, fn fs.WalkDirFunc) error { return filepath.WalkDir(root, fn) }

func (OS) MkdirTemp(dir, pattern string) (string, error) { return os.MkdirTemp(dir, pattern) }
