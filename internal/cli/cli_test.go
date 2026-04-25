package cli

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// runCLI builds a fresh root command and runs it with args, capturing
// stdout and stderr. Each call gets a clean command tree.
func runCLI(t *testing.T, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	cmd := NewRootCmd()
	var so, se bytes.Buffer
	cmd.SetOut(&so)
	cmd.SetErr(&se)
	cmd.SetArgs(args)
	err = cmd.Execute()
	return so.String(), se.String(), err
}

// withRoot prepends --root <tmp>/tickets to a command.
func withRoot(root string, args ...string) []string {
	return append([]string{"--root", root}, args...)
}

func setupTree(t *testing.T) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), "tickets")
	if _, _, err := runCLI(t, withRoot(root, "init")...); err != nil {
		t.Fatalf("init: %v", err)
	}
	return root
}

func TestLifecycle_HappyPath(t *testing.T) {
	root := setupTree(t)

	// `new` with --no-edit so we don't touch $EDITOR.
	stdout, _, err := runCLI(t, withRoot(root,
		"new", "auth-google-oauth-errors",
		"--title", "Better Google OAuth errors",
		"--priority", "high",
		"--label", "auth", "--label", "oauth",
		"--no-edit",
	)...)
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	descPath := filepath.Join(root, "to-do", "auth-google-oauth-errors", "DESCRIPTION.md")
	if !strings.Contains(stdout, descPath) {
		t.Errorf("new stdout = %q, want it to contain %q", stdout, descPath)
	}
	if _, err := os.Stat(descPath); err != nil {
		t.Fatalf("DESCRIPTION.md not written: %v", err)
	}

	// `list` shows the ticket.
	stdout, _, err = runCLI(t, withRoot(root, "list")...)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if !strings.Contains(stdout, "auth-google-oauth-errors") {
		t.Errorf("list missing slug: %q", stdout)
	}
	if !strings.Contains(stdout, "Better Google OAuth errors") {
		t.Errorf("list missing title: %q", stdout)
	}

	// `show` prints the file contents (which must include the title line).
	stdout, _, err = runCLI(t, withRoot(root, "show", "auth-google-oauth-errors")...)
	if err != nil {
		t.Fatalf("show: %v", err)
	}
	if !strings.Contains(stdout, "title: Better Google OAuth errors") {
		t.Errorf("show output missing title: %q", stdout)
	}

	// `start` moves it to in-progress.
	if _, _, err := runCLI(t, withRoot(root, "start", "auth-google-oauth-errors")...); err != nil {
		t.Fatalf("start: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "in-progress", "auth-google-oauth-errors", "DESCRIPTION.md")); err != nil {
		t.Errorf("ticket not in in-progress/: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "to-do", "auth-google-oauth-errors")); !os.IsNotExist(err) {
		t.Errorf("old to-do/ dir still present: %v", err)
	}

	// `done` moves it to done/.
	if _, _, err := runCLI(t, withRoot(root, "done", "auth-google-oauth-errors")...); err != nil {
		t.Fatalf("done: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "done", "auth-google-oauth-errors", "DESCRIPTION.md")); err != nil {
		t.Errorf("ticket not in done/: %v", err)
	}

	// `validate` is OK on a clean tree.
	stdout, _, err = runCLI(t, withRoot(root, "validate")...)
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if !strings.Contains(stdout, "OK") {
		t.Errorf("validate stdout = %q, want OK", stdout)
	}

	// `search` finds the ticket via title substring (case-insensitive).
	stdout, _, err = runCLI(t, withRoot(root, "search", "OAUTH")...)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if !strings.Contains(stdout, "auth-google-oauth-errors") {
		t.Errorf("search missing hit: %q", stdout)
	}
}

func TestNew_RejectsDuplicateSlug(t *testing.T) {
	root := setupTree(t)
	if _, _, err := runCLI(t, withRoot(root, "new", "dup", "--no-edit")...); err != nil {
		t.Fatalf("new: %v", err)
	}
	if _, _, err := runCLI(t, withRoot(root, "new", "dup", "--no-edit")...); err == nil {
		t.Errorf("expected duplicate-slug error")
	}
}

func TestNew_RejectsBadSlug(t *testing.T) {
	root := setupTree(t)
	if _, _, err := runCLI(t, withRoot(root, "new", "Bad Slug", "--no-edit")...); err == nil {
		t.Errorf("expected slug validation error")
	}
}

func TestMove_RejectsInvalidStatus(t *testing.T) {
	root := setupTree(t)
	if _, _, err := runCLI(t, withRoot(root, "new", "x", "--no-edit")...); err != nil {
		t.Fatal(err)
	}
	if _, _, err := runCLI(t, withRoot(root, "move", "x", "bogus")...); err == nil {
		t.Errorf("expected invalid-status error")
	}
}

func TestValidate_DetectsBucketStatusMismatch(t *testing.T) {
	root := setupTree(t)
	if _, _, err := runCLI(t, withRoot(root, "new", "wrong", "--no-edit")...); err != nil {
		t.Fatal(err)
	}
	// Manually relocate the ticket to the wrong bucket so frontmatter
	// (status: pending) no longer matches the directory (done/).
	src := filepath.Join(root, "to-do", "wrong")
	dst := filepath.Join(root, "done", "wrong")
	if err := os.Rename(src, dst); err != nil {
		t.Fatal(err)
	}

	_, stderr, err := runCLI(t, withRoot(root, "validate")...)
	if err == nil {
		t.Fatalf("expected validation failure")
	}
	var ex *ExitError
	if !errors.As(err, &ex) || ex.Code != 2 {
		t.Errorf("expected ExitError code 2, got %v", err)
	}
	if !strings.Contains(stderr, "does not match bucket") {
		t.Errorf("stderr missing mismatch message: %q", stderr)
	}
}

func TestValidate_DetectsMissingDependency(t *testing.T) {
	root := setupTree(t)
	if _, _, err := runCLI(t, withRoot(root, "new", "needy", "--no-edit")...); err != nil {
		t.Fatal(err)
	}
	// Append a dependency on a slug that doesn't exist.
	desc := filepath.Join(root, "to-do", "needy", "DESCRIPTION.md")
	data, err := os.ReadFile(desc)
	if err != nil {
		t.Fatal(err)
	}
	patched := string(data) + "\n## Dependencies\n\n- `ghost` — does not exist\n"
	if err := os.WriteFile(desc, []byte(patched), 0o644); err != nil {
		t.Fatal(err)
	}

	_, stderr, err := runCLI(t, withRoot(root, "validate")...)
	if err == nil {
		t.Fatalf("expected validation failure")
	}
	if !strings.Contains(stderr, `dependency "ghost" not found`) {
		t.Errorf("stderr missing missing-dep message: %q", stderr)
	}
}

func TestList_JSONOutput(t *testing.T) {
	root := setupTree(t)
	if _, _, err := runCLI(t, withRoot(root, "new", "j1", "--title", "first", "--no-edit")...); err != nil {
		t.Fatal(err)
	}
	stdout, _, err := runCLI(t, withRoot(root, "list", "--json")...)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout, `"slug": "j1"`) {
		t.Errorf("json missing slug field: %q", stdout)
	}
	if !strings.Contains(stdout, `"title": "first"`) {
		t.Errorf("json missing title: %q", stdout)
	}
}

func TestDeps_TreeOutput(t *testing.T) {
	root := setupTree(t)
	if _, _, err := runCLI(t, withRoot(root, "new", "child", "--no-edit")...); err != nil {
		t.Fatal(err)
	}
	if _, _, err := runCLI(t, withRoot(root, "new", "parent", "--no-edit")...); err != nil {
		t.Fatal(err)
	}
	parentDesc := filepath.Join(root, "to-do", "parent", "DESCRIPTION.md")
	data, err := os.ReadFile(parentDesc)
	if err != nil {
		t.Fatal(err)
	}
	patched := string(data) + "\n## Dependencies\n\n- `child` — needed first\n"
	if err := os.WriteFile(parentDesc, []byte(patched), 0o644); err != nil {
		t.Fatal(err)
	}

	stdout, _, err := runCLI(t, withRoot(root, "deps", "parent")...)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout, "parent") || !strings.Contains(stdout, "child") {
		t.Errorf("tree missing nodes: %q", stdout)
	}
}
