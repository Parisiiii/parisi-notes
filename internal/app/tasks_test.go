package app

import (
	"strings"
	"testing"
)

func TestStaleSaveKeepsDraftAndNewerDiskData(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	m, err := New()
	if err != nil {
		t.Fatal(err)
	}
	m.addTask("Original")
	m.focusGroup(m.workspace.File)
	m.cursor++
	m.mode = modeNote
	m.editor.SetValue("Unsaved draft")
	newer := []Task{{Title: "Original", Note: "Edited elsewhere"}}
	if err := saveTasks(m.workspace.File, newer); err != nil {
		t.Fatal(err)
	}
	m.updateTask(m.selected(), Task{Title: "Original", Note: m.editor.Value()})
	saved, err := readTasks(m.workspace.File)
	if err != nil || saved[0].Note != "Edited elsewhere" {
		t.Fatal("overwrote newer data", err)
	}
	if m.mode != modeNote || m.editor.Value() != "Unsaved draft" || !strings.Contains(m.status, "changed on disk") {
		t.Fatal("draft or conflict message lost")
	}
}
func TestSaveDoesNotReloadUnrelatedBrokenWorkspace(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	m, err := New()
	if err != nil {
		t.Fatal(err)
	}
	m.addTask("First")
	// A separate workspace's failure should not turn a successful save into a stale UI.
	m.choices[0].File = "/dev/null/tasks.json"
	m.focusGroup(m.workspace.File)
	m.addTask("Second")
	if len(m.ownedTasks(m.workspace.File)) != 2 || m.mode != modeList {
		t.Fatal("successful edit not applied")
	}
}
