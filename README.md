# Parisi

A terminal notebook for tasks, project notes, and AI usage. Built with Go and
[Bubble Tea](https://github.com/charmbracelet/bubbletea).

- Tasks grouped by directory, with a separate notebook for each Git worktree.
- Notes, search, completion, and deletion with confirmation.
- An AI Usage tab powered by [ai-usagebar](https://github.com/akitaonrails/ai-usagebar).

## Run

Requires Go 1.24+ to build. Developed and tested on Linux.

```sh
make install
parisi-notes
```

`make install` places `parisi-notes` in `~/.local/bin`, which is commonly on
the user `PATH`. Override the destination with `make install INSTALL_DIR=/path`.

To build without installing, run `make build`; the executable is written to
`./bin/parisi-notes`.

Run the binary from any project directory to register and select that workspace.
Inside a Git repository, nested directories use the worktree root. Other folders
use the current directory. All registered workspaces appear in one list.

## Keys

| Key | Action |
| --- | --- |
| Tab / Shift+Tab | Switch Notes and AI Usage; drafts are preserved |
| ↑ / ↓ or j / k | Navigate |
| n | Add a task to the selected workspace |
| Enter / e | Expand a workspace or edit a task's note |
| Space | Toggle completion; expand/collapse a heading |
| Ctrl+S | Save a note |
| d / Delete | Delete a task and its note; y confirms, n / Esc cancels |
| / | Search tasks, notes, and workspace names |
| w / p | Jump to the next workspace / Personal |
| a | Reload workspace data |
| u / r | Open AI Usage / refresh it while visible |
| Esc | Cancel editing, clear search, or return to Notes |
| q | Quit from the list or AI Usage |

Tasks save immediately. Notes save with Ctrl+S; Esc discards the draft.
Deletion is permanent after confirmation.

## AI usage

Install and configure **ai-usagebar** separately and put it on your `PATH`.
Parisi reads `ai-usagebar usage --json` and displays provider cards, usage bars,
balances, and reset times. It refreshes every minute while the tab is open.
Failed refreshes retain the previous report. Notes work without this integration.

## Your data

JSON files live in `$XDG_DATA_HOME/parisi`, falling back to
`~/.local/share/parisi`. Back up that directory to keep your notes safe.
Nothing is stored in your project repositories. Personal notes are shared;
directory notebooks are identified by their canonical absolute paths.
Deleting or moving a worktree leaves its old notebook available in the list.

Use one Parisi instance per data directory. Saves use atomic file replacement;
if data changed since loading, the edit is refused and your draft is preserved.
Copy the draft, cancel editing, and press `a` to reload before trying again.

## Development

```sh
make check  # formatting, vet, tests, and build
```

Tests use temporary data and a mock ai-usagebar; no provider credentials are
needed. Git is required for the worktree tests.

`cmd/parisi-notes` starts the app. `internal/app` contains the UI model, input handlers,
views, workspace storage, and usage integration. Dependencies are pinned in
`go.mod` and `go.sum`.
