package app

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var accent = lipgloss.NewStyle().Foreground(lipgloss.Color("117")).Bold(true)
var muted = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))

func (m Model) View() string {
	if m.width < 40 || m.height < 12 {
		return "Please resize to at least 40 columns × 12 rows.\nYour work is preserved."
	}
	body, help, status := m.notesView()
	if m.tab == usageTab {
		body, help, status = m.usageView()
	}
	if !strings.Contains(help, "Tab") {
		help += " • Tab switch tab"
	}
	box := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("60")).Width(m.width - 4).Height(max(3, m.height-8)).Render(body)
	return m.tabsView() + "\n\n" + box + "\n" + lipgloss.NewStyle().MaxWidth(m.width).Render(muted.Render(status)) + "\n" + lipgloss.NewStyle().MaxWidth(m.width).Render(help)
}
func (m Model) notesView() (body, help, status string) {
	status = m.status
	switch m.mode {
	case modeDelete:
		if i := m.selected(); i >= 0 {
			body = "Delete this task and its attached note?\n\n" + clean(m.tasks[i].Title) + "\nWorkspace: " + clean(m.origins[i].Name) + "\n\nThis cannot be undone."
		}
		help = "y delete • n / Esc cancel"
	case modeNew, modeSearch:
		body = m.groupedView() + "\n\n" + m.input.View()
		help = "Enter confirm • Esc cancel"
	case modeNote:
		body = m.editor.View()
		help = "Ctrl+S save • Esc discard changes"
	default:
		body = m.groupedView()
		if m.query != "" {
			status += " • Search: " + clean(m.query) + " • Esc clear"
		}
		help = "n new • Enter expand/edit • Space done • d delete • / search\nTab switch tab • w next workspace • u usage • q quit"
	}
	return
}
func (m Model) usageView() (body, help, status string) {
	body = m.panel.View()
	help = "Tab Notes • ↑/↓ scroll • r refresh • Esc back"
	status = "AI usage • refreshes every minute while this tab is open"
	if !m.usageUpdated.IsZero() {
		status = "Updated " + m.usageUpdated.Format("15:04:05") + " • r refresh"
	}
	if m.loading {
		status = "Refreshing AI usage…"
	}
	if m.usageError != "" {
		status = m.usageError
		if !m.usageUpdated.IsZero() {
			status = "Showing previous data • " + m.usageError
		} else {
			body = m.usageError + "\n\nInstall ai-usagebar and configure your providers.\nThen press r to retry."
		}
	}
	return
}

func (m Model) tabsView() string {
	brand := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("147")).Render("PARISI")
	labels := []string{"Notes", "AI Usage"}
	parts := make([]string, len(labels))
	for i, label := range labels {
		style := lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("245"))
		if i == m.tab {
			style = style.Foreground(lipgloss.Color("232")).Background(lipgloss.Color("117")).Bold(true)
			label = "• " + label
		}
		parts[i] = style.Render(label)
	}
	left := " " + brand + muted.Render("  │  ") + strings.Join(parts, " ")
	hint := muted.Render("tab ⇄")
	width := max(0, m.width-2)
	gap := width - lipgloss.Width(left) - lipgloss.Width(hint)
	if gap >= 3 {
		return left + strings.Repeat(" ", gap) + hint
	}
	return left
}
