package app

import "slices"

// ownedTasks returns a copy, so a failed save never changes the visible tasks.
func (m Model) ownedTasks(file string) []Task {
	tasks := []Task{}
	for i, t := range m.tasks {
		if m.origins[i].File == file {
			tasks = append(tasks, t)
		}
	}
	return tasks
}
func (m *Model) addTask(title string) {
	row, ok := m.activeRow()
	if !ok {
		return
	}
	tasks := append(m.ownedTasks(row.workspace.File), Task{Title: title})
	if m.persistTasks(row.workspace, tasks) {
		m.collapsed[row.workspace.File] = false
	}
}
func (m *Model) updateTask(index int, task Task) {
	owner := m.origins[index]
	tasks := []Task{}
	for i, t := range m.tasks {
		if m.origins[i].File == owner.File {
			if i == index {
				t = task
			}
			tasks = append(tasks, t)
		}
	}
	m.persistTasks(owner, tasks)
}
func (m *Model) persistTasks(owner workspace, tasks []Task) bool {
	disk, err := readTasks(owner.File)
	if err != nil {
		m.status = "Save failed: " + err.Error()
		return false
	}
	if !slices.Equal(disk, m.ownedTasks(owner.File)) {
		m.status = "Notes changed on disk. Copy your draft, cancel, then press a to reload."
		return false
	}
	if err := saveTasks(owner.File, tasks); err != nil {
		m.status = "Save failed: " + err.Error()
		return false
	}
	// Apply only this owner's saved state. Do not reload unrelated workspaces
	// after a successful write: their read failures must not leave this edit stale.
	var all []Task
	var origins []workspace
	for _, w := range m.choices {
		collection := tasks
		if w.File != owner.File {
			collection = m.ownedTasks(w.File)
		}
		for _, task := range collection {
			all = append(all, task)
			origins = append(origins, w)
		}
	}
	m.tasks = all
	m.origins = origins
	m.mode = modeList
	m.cursor = min(m.cursor, max(0, len(m.groupRows())-1))
	m.status = "Saved in " + owner.Name
	return true
}
