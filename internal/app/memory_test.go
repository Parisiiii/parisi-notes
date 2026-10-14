package app

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

func TestMemoryBrowserLoadsProjectsPagesAndMarkdown(t *testing.T) {
	requests := []string{}
	oldClient := memoryHTTPClient
	memoryHTTPClient = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		requests = append(requests, r.URL.EscapedPath())
		if r.Header.Get("Authorization") != "Bearer secret" {
			t.Error("missing ai-memory bearer token")
		}
		body := ""
		status := http.StatusOK
		switch r.URL.Path {
		case "/api/v1/projects":
			body = `[{"workspace_name":"default","project_name":"parisi stats","page_count":1}]`
		case "/api/v1/workspaces/default/projects/parisi stats/pages":
			body = `[{"path":"notes/first page.md","title":"First memory","kind":"note"}]`
		case "/api/v1/workspaces/default/projects/parisi stats/pages/notes/first page.md":
			body = `{"path":"notes/first page.md","title":"First memory","kind":"note","body_markdown":"# Decision\n\nKeep the API read-only."}`
		default:
			status = http.StatusNotFound
			body = `{"error":"not found"}`
		}
		return &http.Response{
			StatusCode: status,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(body)),
		}, nil
	})}
	t.Cleanup(func() { memoryHTTPClient = oldClient })
	t.Setenv("AI_MEMORY_SERVER_URL", "http://memory.test")
	t.Setenv("AI_MEMORY_AUTH_TOKEN", "secret")

	m := Model{width: 80, height: 24, tab: memoryTab, memoryPanel: viewport.New(74, 15)}
	m.handleMemoryMsg(fetchMemoryProjects().(memoryMsg))
	if len(m.memoryProjects) != 1 || !strings.Contains(m.memoryProjectsView(), "parisi stats") {
		t.Fatalf("projects were not rendered: %#v", m.memoryProjects)
	}

	next, cmd := m.handleMemoryKey(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(Model)
	m.handleMemoryMsg(cmd().(memoryMsg))
	if m.memoryLevel != memoryPagesLevel || !strings.Contains(m.memoryPagesView(), "First memory") {
		t.Fatal("pages were not rendered")
	}

	next, cmd = m.handleMemoryKey(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(Model)
	m.handleMemoryMsg(cmd().(memoryMsg))
	if m.memoryLevel != memoryPageLevel || !strings.Contains(m.memoryPanel.View(), "Keep the API read-only") {
		t.Fatal("page markdown was not rendered")
	}
	if len(requests) != 3 || !strings.Contains(requests[1], "parisi%20stats") || !strings.Contains(requests[2], "first%20page.md") {
		t.Fatalf("unexpected escaped API paths: %#v", requests)
	}
}

func TestMemoryFailureKeepsPreviousContent(t *testing.T) {
	m := Model{
		memoryLevel:    memoryProjectsLevel,
		memoryProjects: []memoryProject{{ProjectName: "remembered", WorkspaceName: "default"}},
		memoryLoading:  true,
	}
	m.handleMemoryMsg(memoryMsg{err: errTestMemoryOffline{}})
	if m.memoryLoading || len(m.memoryProjects) != 1 || m.memoryError == "" {
		t.Fatalf("previous memory content was not retained: %#v", m)
	}
}

func TestMemoryTabCyclesAndPreservesNoteDraft(t *testing.T) {
	m := Model{mode: modeNote, editor: textarea.New(), memoryLevel: memoryProjectsLevel}
	m.editor.SetValue("unsaved")
	next, cmd := m.handleKey(tea.KeyMsg{Type: tea.KeyShiftTab})
	m = next.(Model)
	if m.tab != memoryTab || !m.memoryLoading || cmd == nil || m.editor.Value() != "unsaved" {
		t.Fatal("Shift+Tab did not open AI Memory while preserving the draft")
	}
	next, _ = m.handleKey(tea.KeyMsg{Type: tea.KeyTab})
	m = next.(Model)
	if m.tab != notesTab || m.editor.Value() != "unsaved" {
		t.Fatal("Tab did not return to Notes with the draft")
	}
}

type errTestMemoryOffline struct{}

func (errTestMemoryOffline) Error() string { return "offline" }

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func TestMemoryRejectsInvalidServerURL(t *testing.T) {
	t.Setenv("AI_MEMORY_SERVER_URL", "file:///tmp/memory")
	msg := fetchMemoryProjects().(memoryMsg)
	if msg.err == nil || !strings.Contains(msg.err.Error(), "http or https") {
		t.Fatalf("invalid server URL accepted: %v", msg.err)
	}
}

func TestMemoryReadsInstalledAuthToken(t *testing.T) {
	dataDir := t.TempDir()
	t.Setenv("AI_MEMORY_AUTH_TOKEN", "")
	t.Setenv("AI_MEMORY_DATA_DIR", dataDir)
	want := "installed-secret"
	if err := os.WriteFile(filepath.Join(dataDir, "auth-token"), []byte(want+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	got, err := memoryAuthToken()
	if err != nil || got != want {
		t.Fatalf("installed token: got %q, err %v", got, err)
	}
	t.Setenv("AI_MEMORY_AUTH_TOKEN", "environment-secret")
	got, err = memoryAuthToken()
	if err != nil || got != "environment-secret" {
		t.Fatalf("environment token did not take precedence: got %q, err %v", got, err)
	}
}

func TestMemoryRejectsExposedAuthToken(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows does not expose Unix permission bits")
	}
	dataDir := t.TempDir()
	t.Setenv("AI_MEMORY_AUTH_TOKEN", "")
	t.Setenv("AI_MEMORY_DATA_DIR", dataDir)
	if err := os.WriteFile(filepath.Join(dataDir, "auth-token"), []byte("secret"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := memoryAuthToken(); err == nil || !strings.Contains(err.Error(), "owner-only") {
		t.Fatalf("exposed token accepted: %v", err)
	}
}

func TestMemoryUsesLinuxDefaultAuthTokenPath(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux data-directory convention")
	}
	dataHome := t.TempDir()
	t.Setenv("AI_MEMORY_AUTH_TOKEN", "")
	t.Setenv("AI_MEMORY_DATA_DIR", "")
	t.Setenv("XDG_DATA_HOME", dataHome)
	path := filepath.Join(dataHome, "ai-memory", "auth-token")
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("default-secret"), 0600); err != nil {
		t.Fatal(err)
	}
	got, err := memoryAuthToken()
	if err != nil || got != "default-secret" {
		t.Fatalf("default token: got %q, err %v", got, err)
	}
}
