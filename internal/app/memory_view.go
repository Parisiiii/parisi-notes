package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/x/ansi"
)

func (m Model) memoryView() (body, help, status string) {
	status = "AI Memory"
	if !m.memoryUpdated.IsZero() {
		status += " • Updated " + m.memoryUpdated.Format("15:04:05")
	}
	if m.memoryLoading {
		status = "Loading AI Memory…"
	}
	if m.memoryError != "" {
		status = m.memoryError
	}
	switch m.memoryLevel {
	case memoryPageLevel:
		body = m.memoryPanel.View()
		help = "↑/↓ scroll • Esc pages • r refresh"
	case memoryPagesLevel:
		body = m.memoryPagesView()
		help = "↑/↓ select • Enter open • Esc projects • r refresh"
	default:
		body = m.memoryProjectsView()
		help = "↑/↓ select • Enter pages • Esc Notes • r refresh"
	}
	if m.memoryError != "" && len(m.memoryProjects) == 0 {
		action := "Start ai-memory's HTTP server, then press r to retry."
		if strings.Contains(strings.ToLower(m.memoryError), "unauthorized") {
			action = "Set AI_MEMORY_AUTH_TOKEN or install/refresh ai-memory hooks, then press r to retry."
		}
		body = m.memoryError + "\n\n" + action
	}
	return
}

func (m Model) memoryProjectsView() string {
	if len(m.memoryProjects) == 0 {
		if m.memoryLoading {
			return "Loading projects…"
		}
		return "No ai-memory projects found."
	}
	lines := []string{accent.Render("AI Memory projects"), ""}
	return m.memoryRows(lines, len(m.memoryProjects), func(i int) string {
		p := m.memoryProjects[i]
		return fmt.Sprintf("%s / %s  ·  %d pages", p.WorkspaceName, p.ProjectName, p.PageCount)
	})
}

func (m Model) memoryPagesView() string {
	header := []string{accent.Render(safeMemoryText(m.memoryProject.ProjectName)), muted.Render(safeMemoryText(m.memoryProject.WorkspaceName)), ""}
	if len(m.memoryPages) == 0 {
		if m.memoryLoading {
			return strings.Join(header, "\n") + "Loading pages…"
		}
		return strings.Join(header, "\n") + "No memory pages found."
	}
	return m.memoryRows(header, len(m.memoryPages), func(i int) string {
		p := m.memoryPages[i]
		title := p.Title
		if strings.TrimSpace(title) == "" {
			title = p.Path
		}
		return fmt.Sprintf("%-10s %s", p.Kind, title)
	})
}

func (m Model) memoryRows(header []string, count int, label func(int) string) string {
	height := max(1, m.height-10-len(header))
	start := max(0, m.memoryCursor-height+1)
	lines := append([]string{}, header...)
	for i := start; i < count && i < start+height; i++ {
		row := safeMemoryText(label(i))
		row = ansi.Truncate(row, max(1, m.width-10), "…")
		if i == m.memoryCursor {
			row = accent.Render("┃ " + row)
		} else {
			row = "  " + row
		}
		lines = append(lines, row)
	}
	return strings.Join(lines, "\n")
}
