package store

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/krzhapps/GithubTickets/internal/ticket"
)

// newStore returns a Store rooted at <tempdir>/tickets, already Init'd.
func newStore(t *testing.T) *Store {
	t.Helper()
	root := filepath.Join(t.TempDir(), "tickets")
	s, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}
	return s
}

func newTicket(slug, title string, status ticket.Status) *ticket.Ticket {
	return &ticket.Ticket{
		Title:    title,
		Status:   status,
		Priority: ticket.PriorityMedium,
		Created:  "2026-04-25",
		Slug:     slug,
	}
}

func TestInit_Idempotent(t *testing.T) {
	t.Parallel()
	s := newStore(t)

	for _, b := range allBuckets {
		fi, err := os.Stat(filepath.Join(s.Root, b))
		if err != nil {
			t.Fatalf("bucket %s missing: %v", b, err)
		}
		if !fi.IsDir() {
			t.Errorf("bucket %s is not a directory", b)
		}
		if _, err := os.Stat(filepath.Join(s.Root, b, ".gitkeep")); err != nil {
			t.Errorf("bucket %s missing .gitkeep: %v", b, err)
		}
	}

	// Second Init must not fail.
	if err := s.Init(); err != nil {
		t.Errorf("second Init failed: %v", err)
	}
}

func TestCreate_RejectsDuplicateSlug(t *testing.T) {
	t.Parallel()
	s := newStore(t)
	a := newTicket("dupe", "first", ticket.StatusPending)
	if err := s.Create(a); err != nil {
		t.Fatal(err)
	}
	b := newTicket("dupe", "second", ticket.StatusInProgress)
	if err := s.Create(b); err == nil {
		t.Fatal("expected error creating duplicate slug, got nil")
	}
}

func TestCreate_InvalidSlug(t *testing.T) {
	t.Parallel()
	s := newStore(t)
	if err := s.Create(newTicket("Bad Slug!", "x", ticket.StatusPending)); err == nil {
		t.Fatal("expected slug validation error")
	}
}

func TestFind_NotExistIsWrapped(t *testing.T) {
	t.Parallel()
	s := newStore(t)
	_, err := s.Find("ghost")
	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("Find ghost: got %v, want fs.ErrNotExist", err)
	}
}

func TestLoad_FindsAcrossBuckets(t *testing.T) {
	t.Parallel()
	s := newStore(t)

	for _, c := range []struct {
		slug   string
		status ticket.Status
	}{
		{"todo-one", ticket.StatusPending},
		{"wip-one", ticket.StatusInProgress},
		{"blocked-one", ticket.StatusBlocked},
		{"done-one", ticket.StatusDone},
		{"archived-one", ticket.StatusArchived},
	} {
		if err := s.Create(newTicket(c.slug, c.slug, c.status)); err != nil {
			t.Fatalf("create %s: %v", c.slug, err)
		}
	}

	loaded, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(loaded), 5; got != want {
		t.Errorf("Load returned %d tickets, want %d", got, want)
	}
	bySlug := map[string]ticket.Ticket{}
	for _, t := range loaded {
		bySlug[t.Slug] = t
	}
	if bySlug["wip-one"].Status != ticket.StatusInProgress {
		t.Errorf("wip-one status = %q", bySlug["wip-one"].Status)
	}
	if bySlug["blocked-one"].Status != ticket.StatusBlocked {
		t.Errorf("blocked-one status = %q", bySlug["blocked-one"].Status)
	}
}

func TestMove_AcrossBuckets(t *testing.T) {
	t.Parallel()
	s := newStore(t)
	if err := s.Create(newTicket("hop", "Hop", ticket.StatusPending)); err != nil {
		t.Fatal(err)
	}

	moved, err := s.Move("hop", ticket.StatusInProgress)
	if err != nil {
		t.Fatal(err)
	}
	if moved.Status != ticket.StatusInProgress {
		t.Errorf("status = %q, want in-progress", moved.Status)
	}
	if got, want := filepath.Base(filepath.Dir(moved.Dir)), bucketInProgress; got != want {
		t.Errorf("ticket bucket = %q, want %q", got, want)
	}
	// The old location must be gone.
	if _, err := os.Stat(filepath.Join(s.Root, bucketToDo, "hop")); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("expected old dir to be gone, stat err = %v", err)
	}
}

func TestMove_BlockedStaysInInProgressBucket(t *testing.T) {
	t.Parallel()
	s := newStore(t)
	if err := s.Create(newTicket("wip", "Wip", ticket.StatusInProgress)); err != nil {
		t.Fatal(err)
	}
	moved, err := s.Move("wip", ticket.StatusBlocked)
	if err != nil {
		t.Fatal(err)
	}
	if moved.Status != ticket.StatusBlocked {
		t.Errorf("status = %q", moved.Status)
	}
	if got := filepath.Base(filepath.Dir(moved.Dir)); got != bucketInProgress {
		t.Errorf("bucket = %q, want %q", got, bucketInProgress)
	}
}

func TestMove_UsesInjectedRenameFunc(t *testing.T) {
	t.Parallel()
	s := newStore(t)
	if err := s.Create(newTicket("plug", "P", ticket.StatusPending)); err != nil {
		t.Fatal(err)
	}

	var called bool
	s.Rename = func(oldPath, newPath string) error {
		called = true
		return os.Rename(oldPath, newPath)
	}
	if _, err := s.Move("plug", ticket.StatusDone); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Errorf("injected RenameFunc was not invoked")
	}
}

func TestMove_PropagatesRenameError(t *testing.T) {
	t.Parallel()
	s := newStore(t)
	if err := s.Create(newTicket("breaks", "B", ticket.StatusPending)); err != nil {
		t.Fatal(err)
	}
	wantErr := errors.New("rename refused")
	s.Rename = func(_, _ string) error { return wantErr }

	_, err := s.Move("breaks", ticket.StatusDone)
	if !errors.Is(err, wantErr) {
		t.Errorf("got %v, want wrapped %v", err, wantErr)
	}
}

func TestDiscover_FindsTicketsDir(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmp, "tickets"), 0o755); err != nil {
		t.Fatal(err)
	}
	nested := filepath.Join(tmp, "src", "deep")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}

	s, err := Discover(nested)
	if err != nil {
		t.Fatal(err)
	}
	want, _ := filepath.Abs(filepath.Join(tmp, "tickets"))
	if s.Root != want {
		t.Errorf("Root = %s, want %s", s.Root, want)
	}
}

func TestDiscover_FindsViaGitDir(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmp, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	nested := filepath.Join(tmp, "x")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	s, err := Discover(nested)
	if err != nil {
		t.Fatal(err)
	}
	want, _ := filepath.Abs(filepath.Join(tmp, "tickets"))
	if s.Root != want {
		t.Errorf("Root = %s, want %s", s.Root, want)
	}
}

func TestDirForStatus(t *testing.T) {
	t.Parallel()
	cases := map[ticket.Status]string{
		ticket.StatusPending:    bucketToDo,
		ticket.StatusInProgress: bucketInProgress,
		ticket.StatusBlocked:    bucketInProgress,
		ticket.StatusDone:       bucketDone,
		ticket.StatusArchived:   bucketArchived,
		"unknown":               "",
	}
	for s, want := range cases {
		if got := DirForStatus(s); got != want {
			t.Errorf("DirForStatus(%q) = %q, want %q", s, got, want)
		}
	}
}
