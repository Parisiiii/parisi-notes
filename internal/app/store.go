package app

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type Task struct {
	Title string `json:"title"`
	Note  string `json:"note"`
	Done  bool   `json:"done"`
}

func readTasks(path string) ([]Task, error) {
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return []Task{}, nil
	}
	if err != nil {
		return nil, err
	}
	var tasks []Task
	err = json.Unmarshal(b, &tasks)
	return tasks, err
}

// Replace the file atomically so interrupted writes cannot truncate saved notes.
func saveTasks(path string, tasks []Task) error {
	b, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return err
	}
	return atomicWrite(path, b)
}

func atomicWrite(path string, b []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".tasks-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err = f.Write(b); err != nil {
		f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), path)
}
