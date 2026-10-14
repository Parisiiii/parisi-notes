package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
)

type memoryLevel int

const (
	memoryProjectsLevel memoryLevel = iota
	memoryPagesLevel
	memoryPageLevel
)

type memoryProject struct {
	WorkspaceName string     `json:"workspace_name"`
	ProjectName   string     `json:"project_name"`
	PageCount     int        `json:"page_count"`
	LastUpdated   *time.Time `json:"last_updated"`
}

type memoryPageSummary struct {
	Path      string    `json:"path"`
	Title     string    `json:"title"`
	Kind      string    `json:"kind"`
	Tier      string    `json:"tier"`
	UpdatedAt time.Time `json:"updated_at"`
}

type memoryPage struct {
	Path         string    `json:"path"`
	Title        string    `json:"title"`
	Kind         string    `json:"kind"`
	Tier         string    `json:"tier"`
	UpdatedAt    time.Time `json:"updated_at"`
	BodyMarkdown string    `json:"body_markdown"`
}

type memoryMsg struct {
	level    memoryLevel
	projects []memoryProject
	pages    []memoryPageSummary
	page     *memoryPage
	err      error
}

var memoryHTTPClient = http.DefaultClient

func memoryRequest(path string, dst any) error {
	server := strings.TrimSpace(os.Getenv("AI_MEMORY_SERVER_URL"))
	if server == "" {
		server = "http://127.0.0.1:49374"
	}
	base, err := url.Parse(server)
	if err != nil || (base.Scheme != "http" && base.Scheme != "https") || base.Host == "" {
		return errors.New("AI_MEMORY_SERVER_URL must be an http or https URL")
	}
	base.Path = strings.TrimRight(base.Path, "/") + path
	base.RawQuery = ""
	base.Fragment = ""

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base.String(), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	token, err := memoryAuthToken()
	if err != nil {
		return err
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := memoryHTTPClient.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return errors.New("ai-memory request timed out; press r to retry")
		}
		return errors.New("ai-memory server is unavailable; start it and press r")
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return errors.New("could not read the ai-memory response")
	}
	if resp.StatusCode != http.StatusOK {
		var apiErr struct {
			Error string `json:"error"`
		}
		_ = json.Unmarshal(body, &apiErr)
		if apiErr.Error == "" {
			apiErr.Error = http.StatusText(resp.StatusCode)
		}
		return fmt.Errorf("ai-memory: %s", apiErr.Error)
	}
	if err := json.Unmarshal(body, dst); err != nil {
		return errors.New("ai-memory returned invalid JSON; update it and retry")
	}
	return nil
}

func memoryAuthToken() (string, error) {
	if token := strings.TrimSpace(os.Getenv("AI_MEMORY_AUTH_TOKEN")); token != "" {
		return token, nil
	}
	path, err := memoryAuthTokenPath()
	if err != nil {
		return "", err
	}
	info, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", errors.New("could not inspect ai-memory's auth token")
	}
	if !info.Mode().IsRegular() || (runtime.GOOS != "windows" && info.Mode().Perm()&0077 != 0) {
		return "", errors.New("ai-memory's auth token must be an owner-only regular file")
	}
	if info.Size() > 16<<10 {
		return "", errors.New("ai-memory's auth token file is unexpectedly large")
	}
	body, err := os.ReadFile(path)
	if err != nil {
		return "", errors.New("could not read ai-memory's auth token")
	}
	return strings.TrimSpace(string(body)), nil
}

func memoryAuthTokenPath() (string, error) {
	if dir := strings.TrimSpace(os.Getenv("AI_MEMORY_DATA_DIR")); dir != "" {
		return filepath.Join(dir, "auth-token"), nil
	}
	var dataDir string
	switch runtime.GOOS {
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", errors.New("could not locate ai-memory's data directory")
		}
		dataDir = filepath.Join(home, "Library", "Application Support")
	case "windows":
		dataDir = strings.TrimSpace(os.Getenv("LOCALAPPDATA"))
		if dataDir == "" {
			return "", errors.New("could not locate ai-memory's data directory")
		}
	default:
		dataDir = strings.TrimSpace(os.Getenv("XDG_DATA_HOME"))
		if dataDir == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", errors.New("could not locate ai-memory's data directory")
			}
			dataDir = filepath.Join(home, ".local", "share")
		}
	}
	return filepath.Join(dataDir, "ai-memory", "auth-token"), nil
}

func memoryPath(parts ...string) string {
	segments := make([]string, 0, len(parts))
	for _, part := range parts {
		for _, segment := range strings.Split(part, "/") {
			segments = append(segments, segment)
		}
	}
	return "/" + strings.Join(segments, "/")
}

func fetchMemoryProjects() tea.Msg {
	var projects []memoryProject
	err := memoryRequest("/api/v1/projects", &projects)
	return memoryMsg{level: memoryProjectsLevel, projects: projects, err: err}
}

func fetchMemoryPages(project memoryProject) tea.Cmd {
	return func() tea.Msg {
		var pages []memoryPageSummary
		path := memoryPath("api", "v1", "workspaces", project.WorkspaceName, "projects", project.ProjectName, "pages")
		err := memoryRequest(path, &pages)
		return memoryMsg{level: memoryPagesLevel, pages: pages, err: err}
	}
}

func fetchMemoryPage(project memoryProject, summary memoryPageSummary) tea.Cmd {
	return func() tea.Msg {
		var page memoryPage
		path := memoryPath("api", "v1", "workspaces", project.WorkspaceName, "projects", project.ProjectName, "pages", summary.Path)
		err := memoryRequest(path, &page)
		return memoryMsg{level: memoryPageLevel, page: &page, err: err}
	}
}

func (m *Model) handleMemoryMsg(msg memoryMsg) {
	m.memoryLoading = false
	if msg.err != nil {
		m.memoryError = msg.err.Error()
		return
	}
	m.memoryError = ""
	m.memoryUpdated = time.Now()
	switch msg.level {
	case memoryProjectsLevel:
		m.memoryProjects = msg.projects
		m.memoryLevel = memoryProjectsLevel
		m.memoryCursor = boundedCursor(m.memoryCursor, len(m.memoryProjects))
	case memoryPagesLevel:
		m.memoryPages = msg.pages
		m.memoryLevel = memoryPagesLevel
		m.memoryCursor = 0
	case memoryPageLevel:
		m.memoryPage = msg.page
		m.memoryLevel = memoryPageLevel
		m.memoryPanel.GotoTop()
		m.renderMemoryPage()
	}
}

func (m Model) refreshMemory() (tea.Model, tea.Cmd) {
	return m, m.startMemoryRefresh()
}

func (m *Model) startMemoryRefresh() tea.Cmd {
	if m.memoryLoading {
		return nil
	}
	m.memoryLoading = true
	switch m.memoryLevel {
	case memoryPagesLevel:
		return fetchMemoryPages(m.memoryProject)
	case memoryPageLevel:
		if m.memoryPage != nil {
			return fetchMemoryPage(m.memoryProject, memoryPageSummary{Path: m.memoryPage.Path})
		}
	}
	return fetchMemoryProjects
}

func (m Model) handleMemoryKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "r":
		return m.refreshMemory()
	case "esc":
		switch m.memoryLevel {
		case memoryPageLevel:
			m.memoryLevel = memoryPagesLevel
		case memoryPagesLevel:
			m.memoryLevel = memoryProjectsLevel
			m.memoryCursor = 0
		default:
			m.tab = notesTab
		}
		return m, nil
	}
	if m.memoryLoading {
		return m, nil
	}
	if m.memoryLevel == memoryPageLevel {
		var cmd tea.Cmd
		m.memoryPanel, cmd = m.memoryPanel.Update(msg)
		return m, cmd
	}
	rowCount := len(m.memoryProjects)
	if m.memoryLevel == memoryPagesLevel {
		rowCount = len(m.memoryPages)
	}
	switch key {
	case "up", "k":
		m.memoryCursor = boundedCursor(m.memoryCursor-1, rowCount)
	case "down", "j":
		m.memoryCursor = boundedCursor(m.memoryCursor+1, rowCount)
	case "enter":
		if m.memoryLevel == memoryProjectsLevel && rowCount > 0 {
			m.memoryProject = m.memoryProjects[boundedCursor(m.memoryCursor, rowCount)]
			m.memoryLoading = true
			return m, fetchMemoryPages(m.memoryProject)
		}
		if m.memoryLevel == memoryPagesLevel && rowCount > 0 {
			page := m.memoryPages[boundedCursor(m.memoryCursor, rowCount)]
			m.memoryLoading = true
			return m, fetchMemoryPage(m.memoryProject, page)
		}
	}
	return m, nil
}

func safeMemoryText(s string) string {
	return strings.Map(func(r rune) rune {
		if r == '\n' || r == '\t' {
			return r
		}
		if unicode.IsControl(r) {
			return ' '
		}
		return r
	}, ansi.Strip(s))
}

func (m *Model) renderMemoryPage() {
	if m.memoryPage == nil {
		m.memoryPanel.SetContent("")
		return
	}
	header := accent.Render(safeMemoryText(m.memoryPage.Title))
	meta := muted.Render(safeMemoryText(m.memoryPage.Kind + " · " + m.memoryPage.Path))
	m.memoryPanel.SetContent(header + "\n" + meta + "\n\n" + safeMemoryText(m.memoryPage.BodyMarkdown))
}
