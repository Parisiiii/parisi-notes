package app

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestActualWorktreesStaySeparate(t *testing.T) {
	root := t.TempDir()
	repo := filepath.Join(root, "dev")
	other := filepath.Join(root, "dev-2")
	git := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL=/dev/null")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git: %s: %v", out, err)
		}
	}
	git("init", repo)
	git("-C", repo, "-c", "user.name=Test", "-c", "user.email=test@example.com", "-c", "commit.gpgsign=false", "commit", "--allow-empty", "-m", "init")
	git("-C", repo, "worktree", "add", "-b", "second", other)
	sub := filepath.Join(other, "src", "nested")
	if err := os.MkdirAll(sub, 0700); err != nil {
		t.Fatal(err)
	}
	base := filepath.Join(root, "data")
	first, err := registerWorkspace(base, repo)
	if err != nil {
		t.Fatal(err)
	}
	second, err := registerWorkspace(base, sub)
	if err != nil {
		t.Fatal(err)
	}
	if second.Root != other || first.File == second.File {
		t.Fatalf("wrong roots: %#v %#v", first, second)
	}
	if err := saveTasks(first.File, []Task{{Title: "first"}}); err != nil {
		t.Fatal(err)
	}
	if err := saveTasks(second.File, []Task{{Title: "second"}}); err != nil {
		t.Fatal(err)
	}
	alias := filepath.Join(root, "alias")
	if err := os.Symlink(other, alias); err != nil {
		t.Fatal(err)
	}
	aliased, err := registerWorkspace(base, alias)
	if err != nil || aliased.File != second.File {
		t.Fatal("symlink duplicated workspace", err)
	}
	git("-C", repo, "worktree", "remove", "--force", other)
	saved, err := readTasks(second.File)
	if err != nil || len(saved) != 1 || saved[0].Title != "second" {
		t.Fatal("notes did not survive worktree removal", err)
	}
}
func TestAllViewAndPersonalPreserveOwnership(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	m, err := New()
	if err != nil {
		t.Fatal(err)
	}
	personal := workspace{Name: "Personal", File: filepath.Join(m.base, "tasks.json")}
	if err = saveTasks(personal.File, []Task{{Title: "Legacy note"}}); err != nil {
		t.Fatal(err)
	}
	local := m.workspace
	if err = saveTasks(local.File, []Task{{Title: "Project note"}}); err != nil {
		t.Fatal(err)
	}
	if err = m.showAll(); err != nil {
		t.Fatal(err)
	}
	if len(m.tasks) != 2 {
		t.Fatalf("expected both collections: %#v", m.tasks)
	}
	for i, r := range m.groupRows() {
		if r.workspace.File == local.File && r.task >= 0 {
			m.cursor = i
		}
	}
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(Model)
	if m.mode != "note" {
		t.Fatal("did not open owning workspace")
	}
	m.editor.SetValue("Changed locally")
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	m = next.(Model)
	legacy, err := readTasks(personal.File)
	if err != nil || legacy[0].Note != "" {
		t.Fatal("personal notes changed", err)
	}
	m.focusGroup(personal.File)
	if m.tasks[0].Title != "Legacy note" {
		t.Fatal("legacy data unavailable")
	}
}
