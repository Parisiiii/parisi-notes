package app

import (
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestGroupedActionsKeepOwnersAndEmptyHeadings(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	m, err := New()
	if err != nil {
		t.Fatal(err)
	}
	local := m.workspace
	personal := filepath.Join(m.base, "tasks.json")
	if !strings.Contains(m.View(), "▾  Personal") || !strings.Contains(m.View(), "▾  "+local.Name) {
		t.Fatal("missing empty workspace headings")
	}
	m.focusGroup(local.File)
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	m = next.(Model)
	m.input.SetValue("Work task")
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(Model)
	for i, r := range m.groupRows() {
		if r.task >= 0 && r.workspace.File == local.File {
			m.cursor = i
		}
	}
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = next.(Model)
	saved, err := readTasks(local.File)
	if err != nil || len(saved) != 1 || !saved[0].Done {
		t.Fatal("completion was not saved to owner", err)
	}
	other, err := readTasks(personal)
	if err != nil || len(other) != 0 {
		t.Fatal("another workspace was changed", err)
	}
	if !strings.Contains(m.View(), "[x] Work task") {
		t.Fatal("lost grouped view")
	}
	m.focusGroup(local.File)
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(Model)
	if strings.Contains(m.View(), "[x] Work task") {
		t.Fatal("collapse did not hide task")
	}
	m.query = "Work task"
	m.cursor = 0
	if !strings.Contains(m.View(), "[x] Work task") {
		t.Fatal("search should reveal collapsed matches")
	}
}

func TestActiveRowClampsInvalidCursor(t *testing.T) {
	m := Model{
		choices:   []workspace{{Name: "Personal", File: "tasks.json"}},
		collapsed: map[string]bool{},
	}

	for _, cursor := range []int{-1, 10} {
		m.cursor = cursor
		row, ok := m.activeRow()
		if !ok || row.workspace.Name != "Personal" || row.task != -1 {
			t.Fatalf("cursor %d selected %#v, %v", cursor, row, ok)
		}
	}

	m.cursor = -10
	m.groupedKey("down")
	if m.cursor != 0 {
		t.Fatalf("down left cursor out of bounds: %d", m.cursor)
	}
}
