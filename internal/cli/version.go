package cli

import (
	"fmt"
	"runtime/debug"

	"github.com/spf13/cobra"
)

// versionString resolves the CLI version, preferring a value injected at
// build time via -ldflags. When that is absent (e.g. a plain `go install`
// without the Makefile), it falls back to the module version Go embeds in
// the binary, then to a short VCS revision.
func versionString() string {
	if version != "0.0.1-dev" {
		return version
	}

	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return version
	}

	// `go install module@vX.Y.Z` records the tag here.
	if v := bi.Main.Version; v != "" && v != "(devel)" {
		return v
	}

	// Otherwise surface the commit Go embeds from the local VCS.
	for _, s := range bi.Settings {
		if s.Key == "vcs.revision" {
			rev := s.Value
			if len(rev) > 12 {
				rev = rev[:12]
			}
			return "dev-" + rev
		}
	}

	return version
}

func newVersionCmd(g *Globals) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the tickets CLI version",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := fmt.Fprintln(cmd.OutOrStdout(), versionString())
			return err
		},
	}
}
