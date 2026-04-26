package ticket

import (
	"bytes"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// Render serializes the Ticket to canonical Markdown.
//
// Layout:
//
//	---
//	<frontmatter, fields in struct-declared order>
//	---
//
//	## Description
//	...
//
//	## Acceptance Criteria
//	...
//
// Empty body sections are omitted entirely. Dependencies use em-dash as
// the canonical separator. The output ends with exactly one newline.
//
// Round-trip property (enforced by tests):
//
//	Parse(slug, t.Render()) == t  // for serialized fields
func (t *Ticket) Render() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString("---\n")

	fm, err := yaml.Marshal(t)
	if err != nil {
		return nil, fmt.Errorf("marshal frontmatter: %w", err)
	}

	buf.Write(fm)
	buf.WriteString("---\n")

	sections := t.bodySections()
	for _, s := range sections {
		buf.WriteByte('\n')
		buf.WriteString(s)
		buf.WriteByte('\n')
	}

	return buf.Bytes(), nil
}

func (t *Ticket) bodySections() []string {
	var out []string

	if s := strings.TrimSpace(t.Description); s != "" {
		out = append(out, "## Description\n\n"+s)
	}

	if len(t.AcceptanceCriteria) > 0 {
		var sb strings.Builder
		sb.WriteString("## Acceptance Criteria\n\n")
		for i, ac := range t.AcceptanceCriteria {
			if i > 0 {
				sb.WriteByte('\n')
			}
			mark := " "
			if ac.Done {
				mark = "x"
			}
			fmt.Fprintf(&sb, "- [%s] %s", mark, ac.Text)
		}

		out = append(out, sb.String())
	}

	if s := strings.TrimSpace(t.Notes); s != "" {
		out = append(out, "## Notes\n\n"+s)
	}

	if len(t.Dependencies) > 0 {
		var sb strings.Builder
		sb.WriteString("## Dependencies\n\n")
		for i, d := range t.Dependencies {
			if i > 0 {
				sb.WriteByte('\n')
			}
			fmt.Fprintf(&sb, "- `%s` — %s", d.Slug, d.Reason)
		}

		out = append(out, sb.String())
	}

	return out
}
