package ticket

import (
	"bufio"
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// Parse turns DESCRIPTION.md bytes into a Ticket. slug is the directory
// name; it's stamped on the result and not derived from contents.
//
// Parsing is permissive: dependency lines accept em-dash, en-dash, or
// hyphen as the slug/reason separator; AC checkboxes accept "[x]" or
// "[X]". Render emits the canonical forms.
func Parse(slug string, data []byte) (*Ticket, error) {
	fm, body, err := splitFrontmatter(data)
	if err != nil {
		return nil, err
	}

	var t Ticket
	if err := yaml.Unmarshal(fm, &t); err != nil {
		return nil, fmt.Errorf("frontmatter: %w", err)
	}
	t.Slug = slug

	sections := parseSections(body)
	t.Description = sections["Description"]
	t.Notes = sections["Notes"]
	t.AcceptanceCriteria = parseACItems(sections["Acceptance Criteria"])
	t.Dependencies = parseDependencies(sections["Dependencies"])

	return &t, nil
}

// splitFrontmatter splits a file into (frontmatter bytes, body bytes).
// The file must start with a "---" line; the next "---" line ends the
// frontmatter. Both delimiter lines are excluded from the returned slices.
func splitFrontmatter(data []byte) ([]byte, []byte, error) {
	sc := bufio.NewScanner(bytes.NewReader(data))
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	var lines []string
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	if err := sc.Err(); err != nil {
		return nil, nil, err
	}
	if len(lines) == 0 || lines[0] != "---" {
		return nil, nil, fmt.Errorf("missing opening '---' frontmatter delimiter")
	}
	for i := 1; i < len(lines); i++ {
		if lines[i] == "---" {
			fm := []byte(strings.Join(lines[1:i], "\n"))
			body := []byte(strings.Join(lines[i+1:], "\n"))
			return fm, body, nil
		}
	}
	return nil, nil, fmt.Errorf("missing closing '---' frontmatter delimiter")
}

// parseSections splits the body by "## Heading" lines, returning a map
// from heading text to the trimmed contents underneath. Any text before
// the first heading is discarded (the spec requires headings).
func parseSections(body []byte) map[string]string {
	sections := map[string]string{}
	var current string
	var content strings.Builder

	flush := func() {
		if current != "" {
			sections[current] = strings.TrimSpace(content.String())
		}
		content.Reset()
	}

	sc := bufio.NewScanner(bytes.NewReader(body))
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, "## ") {
			flush()
			current = strings.TrimSpace(strings.TrimPrefix(line, "## "))
		} else {
			content.WriteString(line)
			content.WriteByte('\n')
		}
	}
	flush()
	return sections
}

var (
	acItemRe = regexp.MustCompile(`^-\s*\[([ xX])\]\s*(.+)$`)
	// Dep separator accepts em-dash (—), en-dash (–), or one+ ASCII hyphens.
	depRe = regexp.MustCompile("^-\\s+`([^`]+)`\\s*(?:[—–-]+)\\s*(.*)$")
)

func parseACItems(text string) []ACItem {
	if text == "" {
		return nil
	}
	var items []ACItem
	for _, line := range strings.Split(text, "\n") {
		m := acItemRe.FindStringSubmatch(strings.TrimSpace(line))
		if m == nil {
			continue
		}
		done := m[1] == "x" || m[1] == "X"
		items = append(items, ACItem{Done: done, Text: strings.TrimSpace(m[2])})
	}
	return items
}

func parseDependencies(text string) []Dependency {
	if text == "" {
		return nil
	}
	var deps []Dependency
	for _, line := range strings.Split(text, "\n") {
		m := depRe.FindStringSubmatch(strings.TrimSpace(line))
		if m == nil {
			continue
		}
		deps = append(deps, Dependency{
			Slug:   m[1],
			Reason: strings.TrimSpace(m[2]),
		})
	}
	return deps
}
