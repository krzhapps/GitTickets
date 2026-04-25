package cli

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/krzhapps/GithubTickets/internal/ticket"
)

func newDepsCmd(g *Globals) *cobra.Command {
	var asGraph bool
	cmd := &cobra.Command{
		Use:   "deps [<slug>]",
		Short: "Print the dependency tree for one ticket, or every ticket→dep edge",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := storeFromGlobals(g)
			if err != nil {
				return err
			}
			ts, err := s.Load()
			if err != nil {
				return err
			}
			bySlug := make(map[string]ticket.Ticket, len(ts))
			for _, t := range ts {
				bySlug[t.Slug] = t
			}
			out := cmd.OutOrStdout()

			if asGraph {
				return writeDOT(out, ts)
			}
			if len(args) == 1 {
				return writeTree(out, bySlug, args[0])
			}
			return writeFlat(out, ts)
		},
	}
	cmd.Flags().BoolVar(&asGraph, "graph", false, "emit Graphviz DOT")
	return cmd
}

func writeFlat(out io.Writer, ts []ticket.Ticket) error {
	sorted := append([]ticket.Ticket(nil), ts...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Slug < sorted[j].Slug })
	for _, t := range sorted {
		if len(t.Dependencies) == 0 {
			continue
		}
		var slugs []string
		for _, d := range t.Dependencies {
			slugs = append(slugs, d.Slug)
		}
		fmt.Fprintf(out, "%s -> %s\n", t.Slug, strings.Join(slugs, ", "))
	}
	return nil
}

func writeTree(out io.Writer, bySlug map[string]ticket.Ticket, root string) error {
	if _, ok := bySlug[root]; !ok {
		return fmt.Errorf("ticket %q not found", root)
	}
	visited := map[string]bool{}
	var walk func(slug string, prefix string, last bool, depth int)
	walk = func(slug string, prefix string, last bool, depth int) {
		connector := "├── "
		nextPrefix := prefix + "│   "
		if last {
			connector = "└── "
			nextPrefix = prefix + "    "
		}
		if depth == 0 {
			fmt.Fprintln(out, slug)
		} else {
			fmt.Fprintf(out, "%s%s%s\n", prefix, connector, slug)
		}
		if visited[slug] {
			return
		}
		visited[slug] = true
		t, ok := bySlug[slug]
		if !ok {
			return
		}
		for i, d := range t.Dependencies {
			walk(d.Slug, nextPrefix, i == len(t.Dependencies)-1, depth+1)
		}
	}
	walk(root, "", true, 0)
	return nil
}

func writeDOT(out io.Writer, ts []ticket.Ticket) error {
	fmt.Fprintln(out, "digraph tickets {")
	fmt.Fprintln(out, "  rankdir=LR;")
	for _, t := range ts {
		for _, d := range t.Dependencies {
			fmt.Fprintf(out, "  %q -> %q;\n", t.Slug, d.Slug)
		}
	}
	fmt.Fprintln(out, "}")
	return nil
}
