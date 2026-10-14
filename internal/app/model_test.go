package app

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestPersistenceAndCorruptFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data", "tasks.json")
	want := []Task{{Title: "Organizar amanhã", Note: "First line\n第二行", Done: true}}
	if err := saveTasks(path, want); err != nil {
		t.Fatal(err)
	}
	got, err := readTasks(path)
	if err != nil || len(got) != 1 || got[0] != want[0] {
		t.Fatalf("round trip: %#v %v", got, err)
	}
	if err = os.WriteFile(path, []byte("broken"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = readTasks(path); err == nil {
		t.Fatal("corruption must not be treated as an empty list")
	}
}
func TestFilteredCompletionUpdatesCorrectTask(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	m, err := New()
	if err != nil {
		t.Fatal(err)
	}
	m.tasks = []Task{{Title: "First"}, {Title: "Second", Note: "match"}}
	if err := saveTasks(m.workspace.File, m.tasks); err != nil {
		t.Fatal(err)
	}
	m.origins = []workspace{m.workspace, m.workspace}
	m.cursor = 1
	m.query = "match"
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = updated.(Model)
	if m.tasks[0].Done || !m.tasks[1].Done {
		t.Fatal("completed the wrong task")
	}
	saved, err := readTasks(m.workspace.File)
	if err != nil || !saved[1].Done {
		t.Fatal("completion not persisted", err)
	}
}
func TestNoteCancelAndSave(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	m, _ := New()
	m.tasks = []Task{{Title: "Task", Note: "Original"}}
	if err := saveTasks(m.workspace.File, m.tasks); err != nil {
		t.Fatal(err)
	}
	m.origins = []workspace{m.workspace}
	m.focusGroup(m.workspace.File)
	m.cursor++
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)
	m.editor.SetValue("Discard this")
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)
	if m.tasks[0].Note != "Original" {
		t.Fatal("cancel changed saved note")
	}
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)
	m.editor.SetValue("Keep this")
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	m = updated.(Model)
	saved, err := readTasks(m.workspace.File)
	if err != nil || saved[0].Note != "Keep this" {
		t.Fatal("note not saved", err)
	}
}
