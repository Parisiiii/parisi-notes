package app

import (
	"context"
	"errors"
	"os/exec"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type usageTickMsg time.Time

func usageTick() tea.Cmd {
	return tea.Tick(time.Minute, func(t time.Time) tea.Msg { return usageTickMsg(t) })
}
func fetchUsage() tea.Msg {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "ai-usagebar", "usage", "--json")
	cmd.WaitDelay = time.Second
	b, err := cmd.Output()
	if err != nil {
		switch {
		case errors.Is(err, exec.ErrNotFound):
			err = errors.New("ai-usagebar was not found on PATH")
		case ctx.Err() != nil:
			err = errors.New("AI usage request timed out; press r to retry")
		default:
			err = errors.New("ai-usagebar could not load usage; check its configuration and press r")
		}
		return usageMsg{err: err}
	}
	report, err := parseUsage(b)
	return usageMsg{report: report, err: err}
}
