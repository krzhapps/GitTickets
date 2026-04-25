package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newInitCmd(g *Globals) *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Create the tickets/{to-do,in-progress,done,archived} layout",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			s, err := storeFromGlobals(g)
			if err != nil {
				return err
			}
			if err := s.Init(); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "initialized %s\n", s.Root)
			return nil
		},
	}
}
