package app

import (
	"os"
	"path/filepath"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type viewMode string

const (
	modeList   viewMode = ""
	modeNote   viewMode = "note"
	modeNew    viewMode = "new"
	modeSearch viewMode = "search"
	modeDelete viewMode = "delete"
)
const (
	notesTab = iota
	usageTab
)

// Model owns the UI state. Tasks and origins are loaded together in workspace order.
type Model struct {
	tasks                 []Task
	origins               []workspace
	choices               []workspace
	base                  string
	workspace             workspace
	collapsed             map[string]bool
	cursor, width, height int
	mode                  viewMode
	query, status         string
	input                 textinput.Model
	editor                textarea.Model
	panel                 viewport.Model
	tab                   int
	loading               bool
	usageError            string
	usageUpdated          time.Time
	report                *usageReport
}
type usageMsg struct {
	report *usageReport
	err    error
}

func New() (Model, error) {
	dir := os.Getenv("XDG_DATA_HOME")
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return Model{}, err
		}
		dir = filepath.Join(home, ".local", "share")
	}
	base := filepath.Join(dir, "parisi")
	cwd, err := os.Getwd()
	if err != nil {
		return Model{}, err
	}
	w, err := registerWorkspace(base, cwd)
	if err != nil {
		return Model{}, err
	}
	input := textinput.New()
	input.CharLimit = 200
	editor := textarea.New()
	editor.ShowLineNumbers = false
	editor.CharLimit = 0
	m := Model{base: base, workspace: w, width: 80, height: 24, input: input, editor: editor, panel: viewport.New(74, 15)}
	m.collapsed = map[string]bool{}
	if err := m.showAll(); err != nil {
		return Model{}, err
	}
	m.focusGroup(w.File)
	return m, nil
}
func (m Model) Init() tea.Cmd { return usageTick() }

func (m Model) selected() int {
	row, ok := m.activeRow()
	if !ok {
		return -1
	}
	return row.task
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case usageTickMsg:
		if m.tab == usageTab && !m.loading {
			m.loading = true
			return m, tea.Batch(fetchUsage, usageTick())
		}
		return m, usageTick()
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.editor.SetWidth(max(10, m.width-6))
		m.editor.SetHeight(max(3, m.height-10))
		m.input.Width = max(10, m.width-10)
		m.panel.Width = max(10, m.width-6)
		m.panel.Height = max(3, m.height-9)
		m.renderUsage()
	case usageMsg:
		m.loading = false
		if msg.err != nil {
			m.usageError = msg.err.Error()
		} else {
			m.report = msg.report
			m.usageError = ""
			m.usageUpdated = time.Now()
		}
		m.renderUsage()

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	var cmd tea.Cmd
	switch m.mode {
	case modeNote:
		m.editor, cmd = m.editor.Update(msg)
	case modeNew, modeSearch:
		m.input, cmd = m.input.Update(msg)
	}
	return m, cmd
}
