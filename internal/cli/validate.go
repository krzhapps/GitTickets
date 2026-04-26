package cli

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/krzhapps/GitTickets/internal/store"
	"github.com/krzhapps/GitTickets/internal/ticket"
)

func newValidateCmd(g *Globals) *cobra.Command {
	return &cobra.Command{
		Use:   "validate [<slug>]",
		Short: "Validate ticket frontmatter, structure, and dependency graph",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := storeFromGlobals(g)
			if err != nil {
				return err
			}

			only := ""
			if len(args) == 1 {
				only = args[0]
			}

			issues, err := validateAll(s, only)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			if len(issues) == 0 {
				fmt.Fprintln(out, "OK")
				return nil
			}

			for _, iss := range issues {
				fmt.Fprintln(cmd.ErrOrStderr(), iss)
			}

			return exitErr(2, "%d validation issue(s)", len(issues))
		},
	}
}

// validateAll walks every bucket, parses every DESCRIPTION.md it finds,
// and aggregates issues. A malformed file produces an issue rather than
// halting the walk so users see every problem at once.
func validateAll(s *store.Store, only string) ([]string, error) {
	buckets := []string{"to-do", "in-progress", "done", "archived"}

	type loaded struct {
		bucket string
		t      ticket.Ticket
	}
	var ok []loaded
	var issues []string

	for _, b := range buckets {
		entries, err := os.ReadDir(filepath.Join(s.Root, b))
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, err
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			slug := e.Name()
			if only != "" && slug != only {
				continue
			}
			prefix := fmt.Sprintf("%s/%s", b, slug)
			if err := ticket.ValidateSlug(slug); err != nil {
				issues = append(issues, fmt.Sprintf("%s: %v", prefix, err))
				continue
			}
			descPath := filepath.Join(s.Root, b, slug, "DESCRIPTION.md")
			data, err := os.ReadFile(descPath)
			if err != nil {
				issues = append(issues, fmt.Sprintf("%s: %v", prefix, err))
				continue
			}
			t, err := ticket.Parse(slug, data)
			if err != nil {
				issues = append(issues, fmt.Sprintf("%s: %v", prefix, err))
				continue
			}
			issues = append(issues, validateTicket(*t, b)...)
			ok = append(ok, loaded{bucket: b, t: *t})
		}
	}

	if only == "" {
		all := make([]ticket.Ticket, 0, len(ok))
		for _, l := range ok {
			all = append(all, l.t)
		}
		issues = append(issues, validateDeps(all)...)
	}
	return issues, nil
}

func validateTicket(t ticket.Ticket, bucket string) []string {
	var issues []string
	prefix := fmt.Sprintf("%s/%s:", bucket, t.Slug)
	if t.Title == "" {
		issues = append(issues, prefix+" title is required")
	}
	if !t.Status.Valid() {
		issues = append(issues, fmt.Sprintf("%s status %q is not a known value", prefix, t.Status))
	}
	if !t.Priority.Valid() {
		issues = append(issues, fmt.Sprintf("%s priority %q is not low|medium|high", prefix, t.Priority))
	}
	if t.Created == "" {
		issues = append(issues, prefix+" created is required")
	} else if _, err := time.Parse("2006-01-02", t.Created); err != nil {
		issues = append(issues, fmt.Sprintf("%s created %q must be YYYY-MM-DD", prefix, t.Created))
	}
	if t.Status.Valid() && !store.StatusFitsDir(t.Status, bucket) {
		issues = append(issues, fmt.Sprintf("%s status %q does not match bucket %s",
			prefix, t.Status, bucket))
	}
	return issues
}

// validateDeps checks that every Dependencies entry resolves to a real
// ticket and that the graph is acyclic. Cycles are reported once per
// participating edge.
func validateDeps(all []ticket.Ticket) []string {
	bySlug := make(map[string]ticket.Ticket, len(all))
	for _, t := range all {
		bySlug[t.Slug] = t
	}
	var issues []string
	for _, t := range all {
		for _, d := range t.Dependencies {
			if _, ok := bySlug[d.Slug]; !ok {
				issues = append(issues, fmt.Sprintf("%s: dependency %q not found", t.Slug, d.Slug))
			}
		}
	}

	const (
		unseen  = 0
		onStack = 1
		done    = 2
	)
	color := make(map[string]int, len(all))
	var dfs func(slug string)
	dfs = func(slug string) {
		color[slug] = onStack
		for _, d := range bySlug[slug].Dependencies {
			switch color[d.Slug] {
			case unseen:
				if _, ok := bySlug[d.Slug]; ok {
					dfs(d.Slug)
				}
			case onStack:
				issues = append(issues, fmt.Sprintf("dependency cycle: %s -> %s", slug, d.Slug))
			}
		}
		color[slug] = done
	}
	for slug := range bySlug {
		if color[slug] == unseen {
			dfs(slug)
		}
	}
	return issues
}
