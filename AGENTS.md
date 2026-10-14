# AGENTS.md

## Project

Parisi is a Go 1.24 terminal application built with Bubble Tea. It provides
task lists, multiline notes, directory-scoped workspaces, and an optional AI
usage panel.

## Layout

- `cmd/parisi/main.go` starts the Bubble Tea program.
- `internal/app/` contains the application model, update and view logic,
  persistence, workspace discovery, and tests.
- `README.md` defines the user-facing behavior and keyboard controls.
- `Makefile` contains the supported build, test, and run commands.

Keep application behavior and its tests together in `internal/app`. Preserve
the `main` package as a thin startup wrapper.

## Commands

Run these from the repository root:

```sh
make test
make build
make run
```

Use `go test ./...` when a narrower Make target is not sufficient. Run the
relevant tests after every behavioral change; run `make test` before handing
off a completed change.

## Implementation guidance

- Format modified Go files with `gofmt`.
- Follow existing Go naming and package conventions; do not add dependencies
  unless the requested feature requires one.
- Model state transitions through Bubble Tea `Update` messages and keep view
  rendering side-effect free.
- Preserve keyboard behavior documented in `README.md`. Update it when user
  interactions change.
- Task data is stored outside the repository. Keep writes atomic through the
  existing storage helpers and do not add project-local runtime data.
- A workspace is identified by its canonical root. Preserve worktree
  isolation, symlink handling, and ownership when changing grouped task
  operations.
- Notes save only on explicit save; cancelling an edit must retain the stored
  note and discard the draft.
- AI-usage failures must leave task and note workflows usable and retain any
  previous report when applicable.

## Testing expectations

- Add or update focused tests in `internal/app` for changed behavior.
- Use temporary directories and `t.Setenv` for storage or environment-related
  tests so developer data is never touched.
- Cover persistence and workspace ownership carefully when modifying task,
  note, delete, filter, or grouping flows.
