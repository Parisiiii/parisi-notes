package app

import (
	"fmt"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type groupRow struct {
	workspace workspace
	task      int
}

func (m Model) groupRows() []groupRow {
	query := strings.ToLower(m.query)
	tasksByWorkspace := make(map[string][]int, len(m.choices))
	for i := range m.tasks {
		owner := m.origins[i]
		tasksByWorkspace[owner.File] = append(tasksByWorkspace[owner.File], i)
	}

	var rows []groupRow
	for _, w := range m.choices {
		var matches []int
		for _, i := range tasksByWorkspace[w.File] {
			task := m.tasks[i]
			searchable := task.Title + " " + task.Note + " " + w.Name
			if strings.Contains(strings.ToLower(searchable), query) {
				matches = append(matches, i)
			}
		}
		if m.query != "" && len(matches) == 0 {
			continue
		}
		rows = append(rows, groupRow{w, -1})
		if !m.collapsed[w.File] || m.query != "" {
			for _, i := range matches {
				rows = append(rows, groupRow{w, i})
			}
		}
	}
	return rows
}
func (m Model) activeRow() (groupRow, bool) {
	rows := m.groupRows()
	if len(rows) == 0 {
		return groupRow{}, false
	}
	return rows[boundedCursor(m.cursor, len(rows))], true
}

func boundedCursor(cursor, rowCount int) int {
	return max(0, min(cursor, max(0, rowCount-1)))
}

func (m *Model) focusGroup(file string) {
	for i, r := range m.groupRows() {
		if r.workspace.File == file {
			m.cursor = i
			return
		}
	}
}
func (m Model) groupedView() string {
	rows := m.groupRows()
	height := max(1, m.height-10)
	if m.mode == modeNew || m.mode == modeSearch {
		height = max(1, height-3)
	}
	start := max(0, m.cursor-height+1)
	lines := []string{}
	for i := start; i < len(rows) && i < start+height; i++ {
		r := rows[i]
		var label string
		if r.task < 0 {
			marker := "▾"
			if m.collapsed[r.workspace.File] && m.query == "" {
				marker = "▸"
			}
			label = marker + "  " + r.workspace.Name
		} else {
			mark := "[ ]"
			if m.tasks[r.task].Done {
				mark = "[x]"
			}
			label = "   " + mark + " " + m.tasks[r.task].Title
		}
		label = lipgloss.NewStyle().MaxWidth(m.width - 8).Render(clean(label))
		if i == m.cursor {
			label = accent.Render("┃ " + label)
		} else if r.task < 0 {
			label = lipgloss.NewStyle().Foreground(lipgloss.Color("147")).Bold(true).Render("  " + label)
		} else {
			label = "  " + label
		}
		lines = append(lines, label)
	}
	if len(lines) == 0 {
		return "No matching tasks. Esc clears search."
	}
	return strings.Join(lines, "\n")
}
func (m *Model) groupedKey(key string) (tea.Cmd, bool) {
	switch key {
	case "up", "k":
		m.cursor = boundedCursor(m.cursor-1, len(m.groupRows()))
	case "down", "j":
		m.cursor = boundedCursor(m.cursor+1, len(m.groupRows()))
	case "w":
		rows := m.groupRows()
		cursor := boundedCursor(m.cursor, len(rows))
		for step := 1; step <= len(rows); step++ {
			i := (cursor + step) % len(rows)
			if rows[i].task < 0 {
				m.cursor = i
				break
			}
		}
	case "p":
		m.query = ""
		m.focusGroup(filepath.Join(m.base, "tasks.json"))
	case "a":
		if err := m.showAll(); err != nil {
			m.status = err.Error()
		}
	case "enter", "e", " ":
		r, ok := m.activeRow()
		if !ok {
			return nil, true
		}
		if r.task < 0 {
			m.collapsed[r.workspace.File] = !m.collapsed[r.workspace.File]
			return nil, true
		}
		if key == " " {
			task := m.tasks[r.task]
			task.Done = !task.Done
			m.updateTask(r.task, task)
		} else {
			m.mode = modeNote
			m.editor.SetValue(m.tasks[r.task].Note)
			return m.editor.Focus(), true
		}
	case "n":
		r, ok := m.activeRow()
		if !ok {
			m.status = "Clear search to choose a workspace"
			return nil, true
		}
		m.mode = modeNew
		m.input.SetValue("")
		m.input.Placeholder = "New task in " + r.workspace.Name
		m.status = fmt.Sprintf("Adding to %s", r.workspace.Name)
		return m.input.Focus(), true
	default:
		return nil, false
	}
	return nil, true
}
