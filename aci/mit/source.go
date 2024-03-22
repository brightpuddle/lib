package mit

import (
	"os"
	"path/filepath"
	"strings"
)

// Entry is an individual source entry
type Entry struct {
	Class string
	Read  func() ([]byte, error)
}

// Source is a source for the DB data
type Source interface {
	Entries() ([]*Entry, error)
}

// MemSource is an in-memory source.
//
// Used to pass collection results directly without touching disk.
type MemSource struct {
	entries []*Entry
}

// NewMemSource creates a new memory source for database entries.
func NewMemSource() *MemSource {
	return &MemSource{entries: []*Entry{}}
}

// Entries fulfills the Source interface.
func (src *MemSource) Entries() ([]*Entry, error) {
	return src.entries, nil
}

// FolderSource is a temp folder source.
//
// Used to read collection results from a temp folder, e.g. a zip archive.
type FolderSource struct {
	Path string
}

// NewFolderSource creates a new source for folder entries
func NewFolderSource(path string) *FolderSource {
	return &FolderSource{Path: path}
}

// Entries fulfills the Source interface.
func (src *FolderSource) Entries() (entries []*Entry, err error) {
	err = filepath.WalkDir(src.Path, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".json") {
			return nil
		}
		class := strings.TrimSuffix(d.Name(), ".json")
		entries = append(entries, &Entry{
			Class: class,
			Read: func() ([]byte, error) {
				return os.ReadFile(path)
			},
		})
		return nil
	})
	return
}
