package editor

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func TestResolve_Precedence(t *testing.T) {
	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", "")
	if got := Resolve(); got != "vi" {
		t.Errorf("default = %q, want vi", got)
	}

	t.Setenv("EDITOR", "nano")
	if got := Resolve(); got != "nano" {
		t.Errorf("EDITOR set: got %q, want nano", got)
	}

	t.Setenv("VISUAL", "code --wait")
	if got := Resolve(); got != "code --wait" {
		t.Errorf("VISUAL set: got %q, want 'code --wait'", got)
	}
}

func TestResolve_TrimsWhitespace(t *testing.T) {
	t.Setenv("VISUAL", "   ")
	t.Setenv("EDITOR", "  micro  ")
	if got := Resolve(); got != "micro" {
		t.Errorf("got %q, want micro", got)
	}
}

func TestOpen_NonZeroExitIsError(t *testing.T) {
	if _, err := exec.LookPath("false"); err != nil {
		t.Skip("no `false` binary")
	}
	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", "false")
	if err := Open("/dev/null"); err == nil {
		t.Errorf("expected error from non-zero editor exit")
	}
}

func TestOpen_ZeroExitNoError(t *testing.T) {
	if _, err := exec.LookPath("true"); err != nil {
		t.Skip("no `true` binary")
	}
	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", "true")
	if err := Open("/dev/null"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// Verifies that the path argument is actually passed to the editor as
// argv[1] and that whitespace-separated args from the env are forwarded
// before it. Skipped on non-Unix where /bin/sh isn't guaranteed.
func TestOpen_PassesPathAsLastArg(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("requires /bin/sh")
	}
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("no sh on PATH")
	}

	tmp := t.TempDir()
	out := filepath.Join(tmp, "captured")
	script := filepath.Join(tmp, "fake-editor.sh")

	// Script writes its argv (one per line) to $out then exits.
	body := "#!/bin/sh\nfor a in \"$@\"; do echo \"$a\" >> " + out + "; done\n"
	if err := os.WriteFile(script, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", script+" --flag1 --flag2")

	target := "/path/to/ticket.md"
	if err := Open(target); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	want := "--flag1\n--flag2\n" + target + "\n"
	if string(got) != want {
		t.Errorf("argv:\n got: %q\nwant: %q", got, want)
	}
}
