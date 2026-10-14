package app

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

type usageReport struct {
	Entries []usageEntry `json:"entries"`
}
type usageEntry struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	DisplayName string         `json:"display_name"`
	Plan        string         `json:"plan"`
	Error       string         `json:"error"`
	Stale       bool           `json:"stale"`
	Sections    []usageSection `json:"sections"`
}
type usageSection struct {
	Type     string   `json:"type"`
	Label    string   `json:"label"`
	Percent  *float64 `json:"percent"`
	Value    string   `json:"value"`
	Detail   string   `json:"detail"`
	Headline string   `json:"headline"`
	Severity string   `json:"severity"`
	ResetAt  string   `json:"reset_at"`
	Body     []string `json:"body"`
}

func parseUsage(b []byte) (*usageReport, error) {
	var report usageReport
	if err := json.Unmarshal(b, &report); err != nil {
		return nil, fmt.Errorf("ai-usagebar returned invalid JSON; update it and retry")
	}
	if report.Entries == nil {
		return nil, fmt.Errorf("ai-usagebar report has no entries; update it and retry")
	}
	return &report, nil
}
func clean(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return ' '
		}
		return r
	}, ansi.Strip(s))
}
func usageColor(severity string) lipgloss.Color {
	switch severity {
	case "critical":
		return lipgloss.Color("203")
	case "warning", "warn":
		return lipgloss.Color("221")
	default:
		return lipgloss.Color("117")
	}
}
func usageCard(e usageEntry, width int) string {
	inner := max(12, width-6)
	wrap := func(s string) string { return lipgloss.NewStyle().Width(inner).Render(clean(s)) }
	name := e.DisplayName
	if name == "" {
		name = e.Name
	}
	if name == "" {
		name = e.ID
	}
	lines := []string{accent.Render(wrap(name))}
	if e.Plan != "" {
		lines = append(lines, muted.Render(wrap(e.Plan)))
	}
	if e.Stale {
		lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("221")).Render("● Cached · may be outdated"))
	}
	if e.Error != "" {
		lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Render(wrap(e.Error)))
	}
	for _, s := range e.Sections {
		switch s.Type {
		case "metric":
			lines = append(lines, "")
			value := s.Value
			if s.Headline != "value" && s.Percent != nil {
				value = fmt.Sprintf("%.0f%% used", *s.Percent)
			}
			lines = append(lines, wrap(s.Label+"  ·  "+value))
			if s.Percent != nil {
				n := int(max(0, min(100, *s.Percent)) * float64(inner) / 100)
				bar := lipgloss.NewStyle().Foreground(usageColor(s.Severity)).Render(strings.Repeat("━", n)) + muted.Render(strings.Repeat("─", inner-n))
				lines = append(lines, bar)
			}
			if s.Headline != "value" && s.Value != "" {
				lines = append(lines, muted.Render(wrap(s.Value)))
			}
			if s.Detail != "" {
				lines = append(lines, muted.Render(wrap(s.Detail)))
			}
			if t, err := time.Parse(time.RFC3339, s.ResetAt); err == nil {
				lines = append(lines, muted.Render(wrap("Resets "+t.Local().Format("Jan 2 · 15:04 MST"))))
			}
		case "text":
			lines = append(lines, wrap(s.Label+"  "+s.Value))
		case "block":
			lines = append(lines, wrap(s.Label))
			for _, line := range s.Body {
				lines = append(lines, muted.Render(wrap(line)))
			}
		case "spacer":
			lines = append(lines, "")
		}
	}
	if len(e.Sections) == 0 && e.Error == "" {
		lines = append(lines, muted.Render("No metrics reported"))
	}
	return lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("60")).Padding(1, 2).Width(width - 2).Render(strings.Join(lines, "\n"))
}
func (m *Model) renderUsage() {
	if m.report == nil {
		m.panel.SetContent("Open AI Usage to load your providers.")
		return
	}
	if len(m.report.Entries) == 0 {
		m.panel.SetContent("No providers configured yet.\nEnable providers in ai-usagebar, then press r.")
		return
	}
	width := max(24, m.panel.Width)
	columns := 1
	if width >= 88 {
		columns = 2
	}
	cardWidth := (width - (columns-1)*2) / columns
	rows := []string{accent.Render("Your AI usage"), muted.Render(fmt.Sprintf("%d providers / accounts · reported usage", len(m.report.Entries))), ""}
	for i := 0; i < len(m.report.Entries); i += columns {
		cards := []string{}
		for j := i; j < min(i+columns, len(m.report.Entries)); j++ {
			cards = append(cards, usageCard(m.report.Entries[j], cardWidth))
		}
		rows = append(rows, cards[0])
		if len(cards) == 2 {
			rows[len(rows)-1] = lipgloss.JoinHorizontal(lipgloss.Top, cards[0], "  ", cards[1])
		}
		rows = append(rows, "")
	}
	m.panel.SetContent(strings.Join(rows, "\n"))
}
