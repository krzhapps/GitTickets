package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/krzhapps/GithubTickets/internal/render"
)

func newShowCmd(g *Globals) *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "show <slug>",
		Short: "Print a single ticket",
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
			if asJSON {
				return render.TicketJSON(cmd.OutOrStdout(), *t)
			}
			data, err := os.ReadFile(filepath.Join(t.Dir, "DESCRIPTION.md"))
			if err != nil {
				return err
			}
			fmt.Fprint(cmd.OutOrStdout(), string(data))
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "output JSON")
	return cmd
}
