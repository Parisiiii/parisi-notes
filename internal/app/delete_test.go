package app

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func key(m Model, k string) Model {
	n, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)})
	return n.(Model)
}
func TestDeleteFilteredTaskPreservesOtherOwners(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	m, err := New()
	if err != nil {
		t.Fatal(err)
	}
	local := m.workspace.File
	personal := filepath.Join(m.base, "tasks.json")
	if err = saveTasks(personal, []Task{{Title: "Personal"}}); err != nil {
		t.Fatal(err)
	}
	if err = saveTasks(local, []Task{{Title: "Keep"}, {Title: "Remove", Note: "Private note"}, {Title: "Keep too"}}); err != nil {
		t.Fatal(err)
	}
	if err = m.showAll(); err != nil {
		t.Fatal(err)
	}
	m.query = "Remove"
	m.cursor = 1
	m = key(m, "d")
	if m.mode != "delete" {
		t.Fatal("confirmation missing")
	}
	m = key(m, "n")
	if len(m.tasks) != 4 {
		t.Fatal("cancel deleted task")
	}
	m = key(m, "d")
	m = key(m, "y")
	saved, err := readTasks(local)
	if err != nil || len(saved) != 2 || saved[0].Title != "Keep" || saved[1].Title != "Keep too" {
		t.Fatal("wrong deletion", saved, err)
	}
	other, err := readTasks(personal)
	if err != nil || len(other) != 1 {
		t.Fatal("other owner changed", err)
	}
	if m.mode != "" || m.cursor != 0 {
		t.Fatal("bad selection after removing last search match")
	}
}
func TestDeleteLastTaskKeepsHeadingAndFailureKeepsTask(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	m, _ := New()
	file := m.workspace.File
	if err := saveTasks(file, []Task{{Title: "Only"}}); err != nil {
		t.Fatal(err)
	}
	if err := m.showAll(); err != nil {
		t.Fatal(err)
	}
	for i, r := range m.groupRows() {
		if r.task >= 0 {
			m.cursor = i
		}
	}
	m = key(m, "d")
	m = key(m, "y")
	saved, err := readTasks(file)
	if err != nil || len(saved) != 0 {
		t.Fatal("last task not deleted", err)
	}
	m.focusGroup(file)
	m = key(m, "d")
	if m.mode == "delete" {
		t.Fatal("heading deletable")
	}
	m.tasks = []Task{{Title: "Must survive"}}
	m.cursor = 0
	blocked := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(blocked, []byte("block"), 0600); err != nil {
		t.Fatal(err)
	}
	m.workspace.File = filepath.Join(blocked, "tasks.json")
	m.choices = []workspace{m.workspace}
	m.origins = []workspace{m.workspace}
	m.cursor = 1
	m = key(m, "d")
	m = key(m, "y")
	if len(m.tasks) != 1 || m.mode != "delete" {
		t.Fatal("failed write changed state")
	}
}
