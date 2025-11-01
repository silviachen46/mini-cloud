package storage

import (
	"io"
	"os"
	"path/filepath"
)

type FS struct{ root string }

func NewFS(root string) *FS { return &FS{root: root} }

func (f *FS) Init() error { return os.MkdirAll(f.root, 0o755) }

func (f *FS) path(key string) string { return filepath.Join(f.root, filepath.Clean(key)) }

func (f *FS) Put(key string, r io.Reader) (int64, error) {
	if err := os.MkdirAll(filepath.Dir(f.path(key)), 0o755); err != nil { return 0, err }
	out, err := os.Create(f.path(key))
	if err != nil { return 0, err }
	defer out.Close()
	return io.Copy(out, r)
}

func (f *FS) Get(key string, w io.Writer) (int64, error) {
	in, err := os.Open(f.path(key))
	if err != nil { return 0, err }
	defer in.Close()
	return io.Copy(w, in)
}

func (f *FS) Stat(key string) (os.FileInfo, error) {
	return os.Stat(f.path(key))
}

func (f *FS) Delete(key string) error {
	return os.Remove(f.path(key))
}
