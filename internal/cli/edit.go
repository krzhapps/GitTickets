package cli

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/krzhapps/GithubTickets/internal/editor"
)

func newEditCmd(g *Globals) *cobra.Command {
	return &cobra.Command{
		Use:   "edit <slug>",
		Short: "Open a ticket's DESCRIPTION.md in $EDITOR",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := storeFromGlobals(g)
			if err != nil {
				return err
			}

			t, err := s.Find(args[0])
			if err != nil {
				return err
			}

			descPath := filepath.Join(t.Dir, "DESCRIPTION.md")
			if !stdinIsTTY() {
				fmt.Fprintln(cmd.OutOrStdout(), descPath)
				return nil
			}

			return editor.Open(descPath)
		},
	}
}
