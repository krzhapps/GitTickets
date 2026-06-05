package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
)

func TestWorktreeAdd(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name         string
		createBranch bool
		want         [][]string
	}{
		{"new branch", true, [][]string{{"worktree", "add", "-b", "ticket/x", "/wt/x"}}},
		{"existing branch", false, [][]string{{"worktree", "add", "/wt/x", "ticket/x"}}},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			fr := &fakeRunner{responses: map[string]fakeResp{"worktree": {}}}
			g := &Git{Runner: fr}
			if err := g.WorktreeAdd("/wt/x", "ticket/x", c.createBranch); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(fr.calls, c.want) {
				t.Errorf("calls = %v, want %v", fr.calls, c.want)
			}
		})
	}
}

func TestWorktreeRemove(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		force bool
		want  [][]string
	}{
		{"plain", false, [][]string{{"worktree", "remove", "/wt/x"}}},
		{"force", true, [][]string{{"worktree", "remove", "--force", "/wt/x"}}},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			fr := &fakeRunner{responses: map[string]fakeResp{"worktree": {}}}
			g := &Git{Runner: fr}
			if err := g.WorktreeRemove("/wt/x", c.force); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(fr.calls, c.want) {
				t.Errorf("calls = %v, want %v", fr.calls, c.want)
			}
		})
	}
}

func TestParseWorktreeList(t *testing.T) {
	t.Parallel()
	out := []byte("worktree /repo\n" +
		"HEAD abc123\n" +
		"branch refs/heads/main\n" +
		"\n" +
		"worktree /repo-worktrees/foo\n" +
		"HEAD def456\n" +
		"branch refs/heads/ticket/foo\n" +
		"\n" +
		"worktree /repo-worktrees/detached\n" +
		"HEAD 789aaa\n" +
		"detached\n")

	got := parseWorktreeList(out)
	want := []Worktree{
		{Path: "/repo", Head: "abc123", Branch: "main"},
		{Path: "/repo-worktrees/foo", Head: "def456", Branch: "ticket/foo"},
		{Path: "/repo-worktrees/detached", Head: "789aaa", Branch: ""},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("parseWorktreeList = %#v, want %#v", got, want)
	}
}

func TestWorktreeExists(t *testing.T) {
	t.Parallel()
	porcelain := []byte("worktree /repo\nHEAD abc\nbranch refs/heads/main\n\n" +
		"worktree /repo-worktrees/foo\nHEAD def\nbranch refs/heads/ticket/foo\n")
	fr := &fakeRunner{responses: map[string]fakeResp{"worktree": {out: porcelain}}}
	g := &Git{Runner: fr}

	if !g.WorktreeExists("/repo-worktrees/foo") {
		t.Error("WorktreeExists = false for present path, want true")
	}
	if g.WorktreeExists("/repo-worktrees/bar") {
		t.Error("WorktreeExists = true for absent path, want false")
	}
}

// TestWorktree_Integration exercises add/list/remove against a real git repo.
// Skipped when the git binary isn't available (CI sandboxes without git).
func TestWorktree_Integration(t *testing.T) {
	t.Parallel()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git binary not available")
	}
	tmp := t.TempDir()
	repo := filepath.Join(tmp, "repo")
	if err := os.Mkdir(repo, 0o755); err != nil {
		t.Fatal(err)
	}

	runGit := func(dir string, args ...string) {
		t.Helper()
		full := append([]string{
			"-c", "user.email=t@example.com",
			"-c", "user.name=Test",
			"-c", "init.defaultBranch=main",
			"-c", "commit.gpgsign=false",
		}, args...)
		cmd := exec.Command("git", full...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"HOME="+tmp,
			"GIT_CONFIG_GLOBAL=/dev/null",
			"GIT_CONFIG_SYSTEM=/dev/null",
		)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}

	runGit(repo, "init")
	if err := os.WriteFile(filepath.Join(repo, "a.txt"), []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(repo, "add", "a.txt")
	runGit(repo, "commit", "-m", "init")

	g := New(repo)
	wt := filepath.Join(tmp, "repo-worktrees", "foo")

	if err := g.WorktreeAdd(wt, "ticket/foo", true); err != nil {
		t.Fatalf("WorktreeAdd: %v", err)
	}
	if _, err := os.Stat(wt); err != nil {
		t.Errorf("worktree dir missing: %v", err)
	}
	if !g.WorktreeExists(wt) {
		t.Error("WorktreeExists = false after add")
	}

	list, err := g.WorktreeList()
	if err != nil {
		t.Fatal(err)
	}
	var foundBranch string
	for _, w := range list {
		if w.Path == wt {
			foundBranch = w.Branch
		}
	}
	if foundBranch != "ticket/foo" {
		t.Errorf("worktree branch = %q, want ticket/foo", foundBranch)
	}

	if err := g.WorktreeRemove(wt, false); err != nil {
		t.Fatalf("WorktreeRemove: %v", err)
	}
	if g.WorktreeExists(wt) {
		t.Error("WorktreeExists = true after remove")
	}
}
