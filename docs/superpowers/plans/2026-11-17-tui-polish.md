# Snipbox TUI Polish Implementation Plan

> **For agentic workers:** Implement each task in order with test-first changes and verify each slice before continuing.

**Goal:** Make Snipbox's current TUI layout and key interactions consistent, compact, and responsive.

**Architecture:** Keep the existing Model/View split. Correct width accounting at the Lipgloss panel boundary and fix state changes in the existing model; add regression tests in the UI package.

**Tech Stack:** Go, Bubble Tea, Bubbles, Lipgloss, Rust static engine (unchanged).

**Spec:** `docs/superpowers/specs/2026-11-17-tui-polish-design.md`

## Global Constraints

- Keep the existing Go/Rust architecture and dependencies.
- Keep the interface English-only and emoji-free.
- Retain the existing Tokyo Night theme and two-panel composition.
- Preserve atomic command storage and existing confirmation behavior.

## Review Focus

- Terminals at 70, 80, 96, and 120 columns: borders meet expected edges and no row exceeds its panel.
- Long result lists: selected rows remain visible without changing panel height.
- Search strings containing `j` and `k`: characters are typed instead of consumed by navigation.
- Filtered tree views: category counts reflect visible matches.
- History import success or duplicate: one saved entry at most, a closed modal, and an accurate status.
- Minimum-size welcome and form screens: all choices and primary actions remain visible.

### Task 1: Correct responsive panel geometry

**Files:** `cli/internal/ui/view.go`, `cli/internal/ui/model_test.go`

- [x] Add tests asserting rendered panel widths plus the gutter equal terminal width at 70, 80, 96, and 120 columns.
- [x] Run the tests and confirm the current panel arithmetic fails them.
- [x] Fix the shared panel width calculation and the Lipgloss content width passed by `commandList` and `detailsPanel`.
- [x] Run the UI tests and verify exact width and height bounds.

### Task 2: Make tree search and action feedback consistent

**Files:** `cli/internal/ui/model.go`, `cli/internal/ui/view.go`, `cli/internal/ui/model_test.go`

- [x] Add tests for typing `j`/`k` while the tree is active, filtered category counts, and history import success, failure, and duplicate handling.
- [x] Run those tests and confirm each fails for the intended behavior.
- [x] Remove the `j`/`k` navigation aliases, count tree results from active matches, and close the history modal only after a successful save.
- [x] Run UI tests and verify save failure keeps the error visible without claiming success, while an existing command is not duplicated.

### Final verification

- [x] Run `go test ./...` and `go vet ./...` from `cli/`.
- [x] Run `cargo test --manifest-path engine/Cargo.toml`.
- [x] Run `make build` from the repository root and launch `bin/snipbox.exe`.
- [x] Check `git diff --check` and confirm the working tree contains only this task's changes.
