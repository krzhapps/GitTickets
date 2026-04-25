package git

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// fakeRunner captures every Run call and returns canned responses by the
// command's first arg (the git subcommand name).
type fakeRunner struct {
	calls     [][]string
	responses map[string]fakeResp
}

type fakeResp struct {
	out []byte
	err error
}

func (f *fakeRunner) Run(args ...string) ([]byte, error) {
	f.calls = append(f.calls, append([]string(nil), args...))
	if len(args) == 0 {
		return nil, errors.New("no args")
	}
	r, ok := f.responses[args[0]]
	if !ok {
		return nil, errors.New("unexpected git subcommand: " + args[0])
	}
	return r.out, r.err
}

func TestIsRepo(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		resp fakeResp
		want bool
	}{
		{"true", fakeResp{out: []byte("true\n")}, true},
		{"false", fakeResp{out: []byte("false\n")}, false},
		{"error", fakeResp{err: errors.New("not a git repo")}, false},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			g := &Git{Runner: &fakeRunner{responses: map[string]fakeResp{
				"rev-parse": c.resp,
			}}}
			if got := g.IsRepo(); got != c.want {
				t.Errorf("IsRepo = %v, want %v", got, c.want)
			}
		})
	}
}

func TestMv_FallsBackToOsRenameOutsideRepo(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	src := filepath.Join(tmp, "a.txt")
	dst := filepath.Join(tmp, "b.txt")
	if err := os.WriteFile(src, []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}

	fr := &fakeRunner{responses: map[string]fakeResp{
		"rev-parse": {out: []byte("false\n")}, // not a repo
	}}
	g := &Git{Dir: tmp, Runner: fr}

	if err := g.Mv(src, dst); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(dst); err != nil {
		t.Errorf("dst missing after Mv: %v", err)
	}
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Errorf("src still present after Mv: %v", err)
	}
	// Only rev-parse should have been called — no `git mv`.
	for _, c := range fr.calls {
		if c[0] == "mv" {
			t.Errorf("git mv was called when outside a repo: %v", c)
		}
	}
}

func TestMv_UsesGitMvInsideRepo(t *testing.T) {
	t.Parallel()
	fr := &fakeRunner{responses: map[string]fakeResp{
		"rev-parse": {out: []byte("true\n")},
		"mv":        {out: nil},
	}}
	g := &Git{Runner: fr}

	if err := g.Mv("a", "b"); err != nil {
		t.Fatal(err)
	}
	want := [][]string{
		{"rev-parse", "--is-inside-work-tree"},
		{"mv", "a", "b"},
	}
	if !reflect.DeepEqual(fr.calls, want) {
		t.Errorf("calls = %v, want %v", fr.calls, want)
	}
}

func TestBranch(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		checkout bool
		want     [][]string
	}{
		{"plain", false, [][]string{{"branch", "feat-x"}}},
		{"checkout", true, [][]string{{"checkout", "-b", "feat-x"}}},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			fr := &fakeRunner{responses: map[string]fakeResp{
				"branch":   {},
				"checkout": {},
			}}
			g := &Git{Runner: fr}
			if err := g.Branch("feat-x", c.checkout); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(fr.calls, c.want) {
				t.Errorf("calls = %v, want %v", fr.calls, c.want)
			}
		})
	}
}

// TestMv_Integration spins up a real git repo in a tempdir and asserts
// that Mv records a rename rather than a delete+add. Skipped if the git
// binary isn't on PATH (CI sandboxes without git).
func TestMv_Integration(t *testing.T) {
	t.Parallel()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git binary not available")
	}
	tmp := t.TempDir()

	// Use -c to avoid depending on the user's global git identity.
	runGit := func(args ...string) {
		t.Helper()
		full := append([]string{
			"-c", "user.email=t@example.com",
			"-c", "user.name=Test",
			"-c", "init.defaultBranch=main",
			"-c", "commit.gpgsign=false",
		}, args...)
		cmd := exec.Command("git", full...)
		cmd.Dir = tmp
		// Isolate from any host config.
		cmd.Env = append(os.Environ(),
			"HOME="+tmp,
			"GIT_CONFIG_GLOBAL=/dev/null",
			"GIT_CONFIG_SYSTEM=/dev/null",
		)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}

	runGit("init")
	if err := os.WriteFile(filepath.Join(tmp, "a.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit("add", "a.txt")
	runGit("commit", "-m", "init")

	g := New(tmp)
	if err := g.Mv(filepath.Join(tmp, "a.txt"), filepath.Join(tmp, "b.txt")); err != nil {
		t.Fatal(err)
	}

	// Porcelain output for a staged rename starts with "R " — present
	// only when git records a rename, not a delete + add.
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = tmp
	out, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}
	got := strings.TrimSpace(string(out))
	if !strings.HasPrefix(got, "R ") {
		t.Errorf("expected staged rename, got porcelain: %q", got)
	}
}
