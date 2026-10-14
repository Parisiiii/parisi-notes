package app

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/x/ansi"
)

const dashboardFixture = `{"schema_version":1,"entries":[{"id":"anthropic","display_name":"Claude","plan":"Max","sections":[{"type":"metric","label":"5-hour window","percent":42,"value":"42%","headline":"percent","severity":"normal","detail":"Resets in 2h","reset_at":"2026-10-14T15:00:00Z"}]},{"id":"openrouter","name":"OpenRouter","stale":true,"sections":[{"type":"metric","label":"Credits","percent":125,"value":"-$2.50","headline":"value","severity":"critical"},{"type":"text","label":"Monthly spend","value":"$12.50"}]},{"name":"Codex","error":"Sign in to continue","sections":[]}]}`

func TestDashboardContractAndLayout(t *testing.T) {
	report, err := parseUsage([]byte(dashboardFixture))
	if err != nil {
		t.Fatal(err)
	}
	for _, width := range []int{34, 74, 110} {
		m := Model{report: report, panel: viewport.New(width, 100)}
		m.renderUsage()
		view := m.panel.View()
		for _, want := range []string{"Claude", "42% used", "OpenRouter", "-$2.50", "Codex", "Sign in", "Cached"} {
			if !strings.Contains(view, want) {
				t.Fatalf("missing %q at width %d", want, width)
			}
		}
		for _, line := range strings.Split(view, "\n") {
			if ansi.StringWidth(line) > width {
				t.Fatalf("line overflows %d: %q", width, line)
			}
		}
	}
}
func TestInvalidReportsDoNotBecomeZeroUsage(t *testing.T) {
	for _, input := range []string{"not json", "{}", "null"} {
		if _, err := parseUsage([]byte(input)); err == nil {
			t.Fatalf("accepted %q", input)
		}
	}
	report, err := parseUsage([]byte(`{"entries":[]}`))
	if err != nil {
		t.Fatal(err)
	}
	m := Model{report: report, panel: viewport.New(70, 10)}
	m.renderUsage()
	if !strings.Contains(m.panel.View(), "No providers") {
		t.Fatal("missing empty state")
	}
}
