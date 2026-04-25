// Package editor launches the user's preferred text editor on a path.
//
// Resolution order matches the convention used by git, crontab, etc.:
// $VISUAL, then $EDITOR, then "vi". The env value may include arguments
// ("code --wait", "emacsclient -nw") which are split on whitespace.
package editor

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Resolve returns the editor command string from the environment, or
// "vi" if neither $VISUAL nor $EDITOR is set.
func Resolve() string {
	for _, name := range []string{"VISUAL", "EDITOR"} {
		if v := strings.TrimSpace(os.Getenv(name)); v != "" {
			return v
		}
	}
	return "vi"
}

// Open launches the resolved editor on path and blocks until it exits.
// Stdin/Stdout/Stderr are inherited so the editor can take over the
// terminal. Returns the editor's error (wrapped) on non-zero exit or if
// the editor binary is not on PATH.
func Open(path string) error {
	return openWith(Resolve(), path)
}

func openWith(cmdLine, path string) error {
	parts := strings.Fields(cmdLine)
	if len(parts) == 0 {
		return errors.New("editor is empty")
	}
	bin, args := parts[0], append(parts[1:], path)

	c := exec.Command(bin, args...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		return fmt.Errorf("editor %q: %w", cmdLine, err)
	}
	return nil
}
