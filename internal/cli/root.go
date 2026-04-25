// Package cli wires up the tickets CLI command tree.
//
// Cobra lives only in this package; everything beneath internal/ stays
// framework-free so the domain layers can be tested without a CLI runner.
package cli

import (
	"github.com/spf13/cobra"
)

// version is overridden at build time via -ldflags "-X .../cli.version=..."
var version = "0.0.1-dev"

// Globals carries flags shared by every subcommand. It is built once per
// invocation by NewRootCmd and threaded into subcommands so we avoid
// package-level mutable state and keep tests independent.
type Globals struct {
	Root    string // override for the tickets/ root directory (else: discovered)
	Verbose bool
	NoColor bool
}

// NewRootCmd builds a fresh command tree. Tests construct one per case so
// flag state never leaks between runs.
func NewRootCmd() *cobra.Command {
	g := &Globals{}

	cmd := &cobra.Command{
		Use:           "tickets",
		Short:         "A git-backed, file-based ticketing system",
		Long:          rootLongHelp,
		Version:       version,
		SilenceUsage:  true, // don't dump usage on runtime errors
		SilenceErrors: true, // main.go prints the error itself
	}

	cmd.PersistentFlags().StringVar(&g.Root, "root", "", "tickets directory (default: discovered from CWD)")
	cmd.PersistentFlags().BoolVarP(&g.Verbose, "verbose", "v", false, "verbose output")
	cmd.PersistentFlags().BoolVar(&g.NoColor, "no-color", false, "disable color output")

	cmd.AddCommand(
		newInitCmd(g),
		newNewCmd(g),
		newListCmd(g),
		newShowCmd(g),
		newEditCmd(g),
		newMoveCmd(g),
		newStartCmd(g),
		newDoneCmd(g),
		newArchiveCmd(g),
		newValidateCmd(g),
		newSearchCmd(g),
		newDepsCmd(g),
		newBranchCmd(g),
	)

	return cmd
}

// Execute is the entry point invoked from main.
func Execute() error {
	return NewRootCmd().Execute()
}

const rootLongHelp = `tickets manages a directory of Markdown tickets tracked in git.

Tickets live under tickets/{to-do,in-progress,done,archived}/<slug>/DESCRIPTION.md
with YAML frontmatter. The CLI handles creation, listing, lifecycle moves
(preserving history via 'git mv'), validation, search, dependency graphs,
and per-ticket branches.`
