package app

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	if key == "tab" || key == "shift+tab" {
		return m.switchTab()
	}
	if m.tab == usageTab {
		return m.handleUsageKey(msg)
	}

	var cmd tea.Cmd
	switch m.mode {
	case modeDelete:
		cmd = m.handleDeleteKey(key)
	case modeNote:
		cmd = m.handleNoteKey(msg)
	case modeNew, modeSearch:
		cmd = m.handleInputKey(msg)
	default:
		cmd = m.handleListKey(key)
	}
	return m, cmd
}

func (m Model) switchTab() (tea.Model, tea.Cmd) {
	m.tab = 1 - m.tab
	if m.tab == usageTab && (m.usageUpdated.IsZero() || time.Since(m.usageUpdated) > time.Minute) {
		return m.refreshUsage()
	}
	return m, nil
}

func (m Model) handleUsageKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.tab = notesTab
		return m, nil
	case "q", "ctrl+c":
		return m, tea.Quit
	case "r", "u":
		return m.refreshUsage()
	default:
		var cmd tea.Cmd
		m.panel, cmd = m.panel.Update(msg)
		return m, cmd
	}
}

func (m *Model) handleDeleteKey(key string) tea.Cmd {
	switch key {
	case "y", "Y":
		m.deleteSelected()
	case "n", "N", "esc":
		m.mode = modeList
		m.status = "Deletion cancelled"
	}
	return nil
}

func (m *Model) handleNoteKey(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc":
		m.mode = modeList
	case "ctrl+s":
		if i := m.selected(); i >= 0 {
			task := m.tasks[i]
			task.Note = m.editor.Value()
			m.updateTask(i, task)
		}
	default:
		var cmd tea.Cmd
		m.editor, cmd = m.editor.Update(msg)
		return cmd
	}
	return nil
}

func (m *Model) handleInputKey(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc":
		m.mode = modeList
	case "enter":
		if m.mode == modeSearch {
			m.query = m.input.Value()
			m.cursor = 0
			m.mode = modeList
		} else if title := strings.TrimSpace(m.input.Value()); title != "" {
			m.addTask(title)
		}
	default:
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return cmd
	}
	return nil
}

func (m *Model) handleListKey(key string) tea.Cmd {
	if cmd, handled := m.groupedKey(key); handled {
		return cmd
	}
	switch key {
	case "d", "delete":
		m.askDelete()
	case "q", "ctrl+c":
		return tea.Quit
	case "esc":
		m.query = ""
		m.cursor = 0
	case "/":
		m.mode = modeSearch
		m.input.SetValue("")
		m.input.Placeholder = "Search tasks and notes"
		return m.input.Focus()
	case "u":
		m.tab = usageTab
		return m.startUsageRefresh()
	}
	return nil
}

func (m Model) refreshUsage() (tea.Model, tea.Cmd) {
	return m, m.startUsageRefresh()
}

func (m *Model) startUsageRefresh() tea.Cmd {
	if m.loading {
		return nil
	}
	m.loading = true
	return fetchUsage
}
