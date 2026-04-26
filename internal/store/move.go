package store

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/krzhapps/GithubTickets/internal/ticket"
)

// Create writes a brand-new ticket to disk in the bucket dictated by its
// Status. Fails if a ticket with the same slug already exists in any
// bucket. On success the ticket's Dir is set to the new directory.
func (s *Store) Create(t *ticket.Ticket) error {
	if err := ticket.ValidateSlug(t.Slug); err != nil {
		return err
	}

	if !t.Status.Valid() {
		return fmt.Errorf("invalid status %q", t.Status)
	}

	if existing, err := s.Find(t.Slug); err == nil {
		return fmt.Errorf("ticket %q already exists at %s", t.Slug, existing.Dir)
	} else if !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	bucket := DirForStatus(t.Status)
	dir := filepath.Join(s.Root, bucket, t.Slug)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	t.Dir = dir
	return s.Save(t)
}

// Save writes the ticket back to disk at its current Dir. Re-renders
// the canonical Markdown each time, so an in-memory edit is enough to
// rewrite the file.
func (s *Store) Save(t *ticket.Ticket) error {
	if t.Dir == "" {
		return fmt.Errorf("ticket %q has no Dir set", t.Slug)
	}

	data, err := t.Render()
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(t.Dir, "DESCRIPTION.md"), data, 0o644)
}

// Move transitions a ticket to a new status. If the new status maps to a
// different bucket than the current location, the ticket directory is
// renamed via the Store's RenameFunc (or os.Rename when nil). The
// frontmatter status is then updated and persisted.
//
// Move is intentionally idempotent: calling Move(slug, current) just
// rewrites the frontmatter — useful when an earlier Move's Save half
// failed and left stale frontmatter in the new location.
func (s *Store) Move(slug string, target ticket.Status) (*ticket.Ticket, error) {
	if !target.Valid() {
		return nil, fmt.Errorf("invalid status %q", target)
	}

	t, err := s.Find(slug)
	if err != nil {
		return nil, err
	}

	targetBucket := DirForStatus(target)
	currentBucket := filepath.Base(filepath.Dir(t.Dir))

	if currentBucket != targetBucket {
		newParent := filepath.Join(s.Root, targetBucket)
		if err := os.MkdirAll(newParent, 0o755); err != nil {
			return nil, err
		}

		newDir := filepath.Join(newParent, slug)
		rename := s.Rename
		if rename == nil {
			rename = os.Rename
		}

		if err := rename(t.Dir, newDir); err != nil {
			return nil, fmt.Errorf("rename %s -> %s: %w", t.Dir, newDir, err)
		}

		t.Dir = newDir
	}

	t.Status = target
	if err := s.Save(t); err != nil {
		return nil, err
	}

	return t, nil
}
