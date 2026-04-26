package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/krzhapps/GithubTickets/internal/git"
)

func newBranchCmd(g *Globals) *cobra.Command {
	var checkout bool

	cmd := &cobra.Command{
		Use:   "branch <slug>",
		Short: "Create a git branch named after a ticket slug",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			slug := args[0]
			s, err := storeFromGlobals(g)
			if err != nil {
				return err
			}

			if _, err := s.Find(slug); err != nil {
				return err
			}

			gw := git.New(repoRoot(s.Root))
			if !gw.IsRepo() {
				return fmt.Errorf("not a git repository: %s", repoRoot(s.Root))
			}

			if err := gw.Branch(slug, checkout); err != nil {
				return err
			}

			verb := "created branch"
			if checkout {
				verb = "switched to branch"
			}

			fmt.Fprintf(cmd.OutOrStdout(), "%s %s\n", verb, slug)
			return nil
		},
	}

	cmd.Flags().BoolVar(&checkout, "checkout", false, "checkout the branch after creating it")
	return cmd
}
