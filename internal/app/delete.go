package app

func (m *Model) deleteSelected() {
	selected := m.selected()
	if selected < 0 {
		m.mode = modeList
		return
	}
	owner := m.origins[selected]
	next := []Task{}
	for i, t := range m.tasks {
		if i != selected && m.origins[i].File == owner.File {
			next = append(next, t)
		}
	}
	if m.persistTasks(owner, next) {
		m.status = "Deleted from " + owner.Name
	}
}
func (m *Model) askDelete() {
	if m.selected() < 0 {
		m.status = "Select a task to delete; workspace headings are kept"
		return
	}
	m.mode = modeDelete
}
