// Package store is the filesystem layer for the tickets/ tree.
//
// It owns the bucket-directory layout (to-do, in-progress, done, archived)
// and the operations that read, create, save, and move ticket directories.
// It does NOT know about git: rename operations are pluggable via the
// RenameFunc field so the CLI can swap in `git mv` while keeping store
// trivially testable against a tempdir.
package store

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/krzhapps/GithubTickets/internal/ticket"
)

// RenameFunc renames a directory (or file). Same shape as os.Rename so a
// nil value can default to it transparently.
type RenameFunc func(oldPath, newPath string) error

// Store wraps a tickets/ root directory.
type Store struct {
	// Root is the absolute path to the tickets/ directory.
	Root string

	// Rename, if non-nil, is used for ticket directory moves. The git
	// package wires in a `git mv`-backed implementation; tests and
	// non-git use cases leave this nil and get os.Rename.
	Rename RenameFunc
}

// Bucket directory names — the on-disk grouping. Kept package-private:
// callers that care about status should go through DirForStatus.
const (
	bucketToDo       = "to-do"
	bucketInProgress = "in-progress"
	bucketDone       = "done"
	bucketArchived   = "archived"
)

// allBuckets is the iteration order for Load — pending work surfaces first.
var allBuckets = []string{bucketToDo, bucketInProgress, bucketDone, bucketArchived}

// DirForStatus returns the bucket directory a ticket with the given status
// belongs in. Both in-progress and blocked share the in-progress/ bucket.
// Returns "" for an unknown status.
func DirForStatus(s ticket.Status) string {
	switch s {
	case ticket.StatusPending:
		return bucketToDo
	case ticket.StatusInProgress, ticket.StatusBlocked:
		return bucketInProgress
	case ticket.StatusDone:
		return bucketDone
	case ticket.StatusArchived:
		return bucketArchived
	}
	return ""
}

// StatusFitsDir reports whether a ticket with the given status is allowed
// to live in the given bucket directory. Used by `tickets validate`.
func StatusFitsDir(s ticket.Status, dir string) bool {
	return DirForStatus(s) == dir
}

// Open returns a Store rooted at the given path. The directory does not
// have to exist yet — Init can create it.
func Open(root string) (*Store, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	return &Store{Root: abs}, nil
}

// Discover walks up from startDir looking for a directory that contains
// either a tickets/ subdirectory or a .git/ subdirectory, and returns a
// Store rooted at <found>/tickets. The tickets/ directory itself need
// not exist yet — Init can populate it.
func Discover(startDir string) (*Store, error) {
	abs, err := filepath.Abs(startDir)
	if err != nil {
		return nil, err
	}
	dir := abs
	for {
		if isDir(filepath.Join(dir, "tickets")) || isDir(filepath.Join(dir, ".git")) {
			return Open(filepath.Join(dir, "tickets"))
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return nil, fmt.Errorf("no tickets/ or .git directory found from %s", startDir)
		}
		dir = parent
	}
}

// Init creates the four bucket directories with empty .gitkeep files so
// they're tracked by git even when empty. Idempotent.
func (s *Store) Init() error {
	for _, b := range allBuckets {
		bucket := filepath.Join(s.Root, b)
		if err := os.MkdirAll(bucket, 0o755); err != nil {
			return err
		}
		gk := filepath.Join(bucket, ".gitkeep")
		if _, err := os.Stat(gk); errors.Is(err, fs.ErrNotExist) {
			if err := os.WriteFile(gk, nil, 0o644); err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
	}
	return nil
}

// isDir reports whether p exists and is a directory.
func isDir(p string) bool {
	fi, err := os.Stat(p)
	return err == nil && fi.IsDir()
}
