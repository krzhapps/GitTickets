package ticket

import (
	"flag"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// -update regenerates the golden Markdown files in testdata/tickets/.
// Run: `go test ./internal/ticket -update` after intentional format changes,
// then review the diff before committing.
var update = flag.Bool("update", false, "update golden testdata files")

// goldenCases pairs a slug+golden-file with the Ticket it must parse to
// and render from. Each case enforces the round-trip property.
func goldenCases() []struct {
	name string
	file string
	slug string
	want Ticket
} {
	return []struct {
		name string
		file string
		slug string
		want Ticket
	}{
		{
			name: "full",
			file: "testdata/tickets/full.md",
			slug: "auth-google-oauth-errors",
			want: Ticket{
				Title:    "Better Google OAuth errors",
				Status:   StatusPending,
				Priority: PriorityHigh,
				Created:  "2026-04-25",
				Labels:   []string{"auth", "oauth"},
				Assignee: "alice",
				Description: "Surface friendlier error messages when Google OAuth fails so users " +
					"can self-correct instead of contacting support.",
				AcceptanceCriteria: []ACItem{
					{Done: false, Text: "Map common Google error codes to user-facing strings"},
					{Done: true, Text: "Add structured logging for OAuth failures"},
				},
				Notes: "Reference: https://developers.google.com/identity/protocols/oauth2",
				Dependencies: []Dependency{
					{Slug: "auth-session-store", Reason: "needs new error types"},
				},
				Slug: "auth-google-oauth-errors",
			},
		},
		{
			name: "minimal",
			file: "testdata/tickets/minimal.md",
			slug: "tests-scraper-coverage",
			want: Ticket{
				Title:    "Increase scraper test coverage",
				Status:   StatusPending,
				Priority: PriorityLow,
				Created:  "2026-04-25",
				Slug:     "tests-scraper-coverage",
			},
		},
	}
}

func TestRoundTripGolden(t *testing.T) {
	for _, c := range goldenCases() {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			// Render the expected ticket to bytes — this is the source of
			// truth for what's on disk. With -update we write it; otherwise
			// we compare it to the committed golden.
			rendered, err := c.want.Render()
			if err != nil {
				t.Fatalf("Render: %v", err)
			}

			if *update {
				if err := os.MkdirAll(filepath.Dir(c.file), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(c.file, rendered, 0o644); err != nil {
					t.Fatal(err)
				}
			}

			onDisk, err := os.ReadFile(c.file)
			if err != nil {
				t.Fatalf("read golden: %v (run `go test -update` to seed)", err)
			}
			if string(rendered) != string(onDisk) {
				t.Errorf("Render() does not match golden %s\n--- got ---\n%s\n--- want ---\n%s",
					c.file, rendered, onDisk)
			}

			// Parse the on-disk bytes and assert every serialized field
			// equals the expected struct.
			parsed, err := Parse(c.slug, onDisk)
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}
			parsed.Dir = "" // not set by Parse; keep comparison clean
			if !reflect.DeepEqual(*parsed, c.want) {
				t.Errorf("Parse(%s) mismatch:\n got:  %+v\n want: %+v", c.file, *parsed, c.want)
			}

			// Round-trip property: rendered → parsed must equal the
			// original Ticket. (Above already covers this, but assert
			// explicitly so a regression is unambiguous.)
			rt, err := Parse(c.slug, rendered)
			if err != nil {
				t.Fatalf("Parse(rendered): %v", err)
			}
			rt.Dir = ""
			if !reflect.DeepEqual(*rt, c.want) {
				t.Errorf("round-trip mismatch:\n got:  %+v\n want: %+v", *rt, c.want)
			}
		})
	}
}

func TestParse_Errors(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		in   string
	}{
		{"empty", ""},
		{"no opening delim", "title: x\n---\n"},
		{"no closing delim", "---\ntitle: x\n"},
		{"bad yaml", "---\ntitle: [unterminated\n---\n"},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if _, err := Parse("any", []byte(c.in)); err == nil {
				t.Errorf("expected error for %q, got nil", c.name)
			}
		})
	}
}

func TestParse_LenientDependencyDashes(t *testing.T) {
	t.Parallel()
	// All three should parse to the same dependency. Render emits em-dash;
	// hand-edited tickets often use ASCII --, so we accept both.
	inputs := []string{
		"---\ntitle: t\nstatus: pending\npriority: low\ncreated: \"2026-04-25\"\n---\n\n## Dependencies\n\n- `other` — reason\n",
		"---\ntitle: t\nstatus: pending\npriority: low\ncreated: \"2026-04-25\"\n---\n\n## Dependencies\n\n- `other` -- reason\n",
		"---\ntitle: t\nstatus: pending\npriority: low\ncreated: \"2026-04-25\"\n---\n\n## Dependencies\n\n- `other` - reason\n",
	}
	for i, in := range inputs {
		got, err := Parse("x", []byte(in))
		if err != nil {
			t.Fatalf("input %d: %v", i, err)
		}
		want := []Dependency{{Slug: "other", Reason: "reason"}}
		if !reflect.DeepEqual(got.Dependencies, want) {
			t.Errorf("input %d: got %+v, want %+v", i, got.Dependencies, want)
		}
	}
}

func TestParse_LenientCheckboxCase(t *testing.T) {
	t.Parallel()
	in := "---\ntitle: t\nstatus: pending\npriority: low\ncreated: \"2026-04-25\"\n---\n\n" +
		"## Acceptance Criteria\n\n- [X] upper\n- [x] lower\n- [ ] open\n"
	got, err := Parse("x", []byte(in))
	if err != nil {
		t.Fatal(err)
	}
	want := []ACItem{
		{Done: true, Text: "upper"},
		{Done: true, Text: "lower"},
		{Done: false, Text: "open"},
	}
	if !reflect.DeepEqual(got.AcceptanceCriteria, want) {
		t.Errorf("got %+v, want %+v", got.AcceptanceCriteria, want)
	}
}
