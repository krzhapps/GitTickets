package store

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/krzhapps/GithubTickets/internal/ticket"
)

// Load returns every ticket present in any bucket directory. The Dir
// field on each returned Ticket is set to the absolute path of its
// <slug>/ directory.
//
// Bucket directories that don't exist yet are skipped (so Load works on
// a freshly-initialized or partially-populated tree). A directory that
// exists but lacks a DESCRIPTION.md is silently ignored. A malformed
// DESCRIPTION.md returns an error so callers can surface it.
func (s *Store) Load() ([]ticket.Ticket, error) {
	var tickets []ticket.Ticket
	for _, b := range allBuckets {
		bucket := filepath.Join(s.Root, b)
		entries, err := os.ReadDir(bucket)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}
			return nil, err
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			t, err := readTicket(bucket, e.Name())
			if err != nil {
				if errors.Is(err, fs.ErrNotExist) {
					continue // dir without DESCRIPTION.md — skip
				}
				return nil, err
			}
			tickets = append(tickets, *t)
		}
	}
	return tickets, nil
}

// Find returns the ticket with the given slug, searching every bucket.
// Returns an error wrapping fs.ErrNotExist if no such ticket exists, so
// callers can use errors.Is(err, fs.ErrNotExist).
func (s *Store) Find(slug string) (*ticket.Ticket, error) {
	if err := ticket.ValidateSlug(slug); err != nil {
		return nil, err
	}
	for _, b := range allBuckets {
		bucket := filepath.Join(s.Root, b)
		t, err := readTicket(bucket, slug)
		if err == nil {
			return t, nil
		}
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		return nil, err
	}
	return nil, fmt.Errorf("ticket %q: %w", slug, fs.ErrNotExist)
}

// readTicket reads bucket/slug/DESCRIPTION.md and returns the parsed
// Ticket with Dir populated. fs.ErrNotExist if the file is absent.
func readTicket(bucket, slug string) (*ticket.Ticket, error) {
	dir := filepath.Join(bucket, slug)
	descPath := filepath.Join(dir, "DESCRIPTION.md")
	data, err := os.ReadFile(descPath)
	if err != nil {
		return nil, err
	}
	t, err := ticket.Parse(slug, data)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", descPath, err)
	}
	t.Dir = dir
	return t, nil
}
