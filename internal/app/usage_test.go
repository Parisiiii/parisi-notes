package app

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestTabsPreserveDraftAndWorkspaceSelection(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	m, err := New()
	if err != nil {
		t.Fatal(err)
	}
	m.mode = "note"
	m.editor.SetValue("Unsaved draft")
	cursor := m.cursor
	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = next.(Model)
	if m.tab != 1 || cmd == nil || !m.loading {
		t.Fatal("tab did not trigger usage load")
	}
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	m = next.(Model)
	if m.tab != 0 || m.mode != "note" || m.editor.Value() != "Unsaved draft" || m.cursor != cursor {
		t.Fatal("switch lost editing state")
	}
}
func TestUsageFailureKeepsLastReport(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	m, _ := New()
	m.tab = 1
	next, _ := m.Update(usageMsg{report: &usageReport{Entries: []usageEntry{{Name: "Claude 40%"}}}})
	m = next.(Model)
	next, _ = m.Update(usageMsg{err: errors.New("offline")})
	m = next.(Model)
	if !strings.Contains(m.View(), "Claude 40%") || !strings.Contains(m.View(), "previous data") {
		t.Fatal("stale report not preserved")
	}
	m.loading = false
	_, cmd := m.Update(usageTickMsg(time.Now()))
	if cmd == nil {
		t.Fatal("refresh timer not scheduled")
	}
}
func TestUsageCommandAndMissingExecutable(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("PATH", dir)
	if msg := fetchUsage().(usageMsg); msg.err == nil {
		t.Fatal("missing executable not reported")
	}
	script := `#!/bin/sh
[ "$1" = usage ] && [ "$2" = --json ] || exit 2
printf '%s' '{"entries":[{"name":"Claude","sections":[]}]}'
`

	if err := os.WriteFile(filepath.Join(dir, "ai-usagebar"), []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	msg := fetchUsage().(usageMsg)
	if msg.err != nil || msg.report == nil || msg.report.Entries[0].Name != "Claude" {
		t.Fatalf("report: %#v", msg)
	}
}
