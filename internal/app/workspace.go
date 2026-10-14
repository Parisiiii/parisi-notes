package app

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type workspace struct {
	Name string
	Root string
	File string
}

// The nearest .git marker is the worktree root, not the shared Git directory.
func workspaceRoot(cwd string) (string, error) {
	root, err := filepath.Abs(cwd)
	if err != nil {
		return "", err
	}
	root, err = filepath.EvalSymlinks(root)
	if err != nil {
		return "", err
	}
	for dir := root; ; dir = filepath.Dir(dir) {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir, nil
		} else if !os.IsNotExist(err) {
			return "", err
		}
		if filepath.Dir(dir) == dir {
			break
		}
	}
	return root, nil
}
func registerWorkspace(base, cwd string) (workspace, error) {
	root, err := workspaceRoot(cwd)
	if err != nil {
		return workspace{}, err
	}
	id := fmt.Sprintf("%x", sha256.Sum256([]byte(root)))
	dir := filepath.Join(base, "workspaces", id)
	w := workspace{Name: filepath.Base(root), Root: root, File: filepath.Join(dir, "tasks.json")}
	b, err := json.Marshal(w)
	if err != nil {
		return w, err
	}
	err = atomicWrite(filepath.Join(dir, "workspace.json"), b)
	return w, err
}
func workspaces(base string) ([]workspace, error) {
	list := []workspace{{Name: "Personal", File: filepath.Join(base, "tasks.json")}}
	files, err := filepath.Glob(filepath.Join(base, "workspaces", "*", "workspace.json"))
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		b, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}
		var w workspace
		if err = json.Unmarshal(b, &w); err != nil {
			return nil, fmt.Errorf("read workspace %s: %w", file, err)
		}
		w.File = filepath.Join(filepath.Dir(file), "tasks.json")
		list = append(list, w)
	}
	sort.Slice(list[1:], func(i, j int) bool { return list[i+1].Root < list[j+1].Root })
	return list, nil
}
func (m *Model) showAll() error {
	ws, err := workspaces(m.base)
	if err != nil {
		return err
	}
	var tasks []Task
	var origins []workspace
	for _, w := range ws {
		ts, err := readTasks(w.File)
		if err != nil {
			return err
		}
		for _, t := range ts {
			tasks = append(tasks, t)
			origins = append(origins, w)
		}
	}
	m.choices = ws
	if m.collapsed == nil {
		m.collapsed = map[string]bool{}
	}
	m.tasks = tasks
	m.origins = origins
	m.cursor = 0
	m.query = ""
	m.mode = modeList
	m.status = "All workspaces • Select a heading and press n to add a task"
	return nil
}
