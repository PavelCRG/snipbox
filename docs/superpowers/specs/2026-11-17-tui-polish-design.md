# Snipbox TUI Polish Design

## Intent

Make the existing terminal UI feel deliberate and reliable: panels should use the available width cleanly, search text must not collide with navigation keys, counts should describe the visible results, and successful actions should leave the correct screen active.

## Goals

- Keep the current Go, Bubble Tea, Bubbles, and Lipgloss architecture and Tokyo Night palette.
- Make the two panels and their rounded borders fit the terminal width at supported sizes without `fit` hiding geometry mistakes.
- Keep ordinary letters available to live search while browsing the tree.
- Close the shell history dialog after a successful import and show result counts consistent with the active search.
- Avoid adding a history command when its exact command text already exists in the library.
- Keep dashboard choices and modal actions visible at the minimum supported terminal size.
- Add behavior tests that fail on the current defects and exercise narrow, standard, and wide terminal widths.

## Non-goals

- No new dependencies, new interface language, or wholesale visual redesign.
- No changes to command data, Rust engine, persistence format, or Git sync.

## Approaches

1. **Correct the current layout and state transitions (chosen).** Reuse the existing components and styles, fix shared calculations and action completion, and pin behavior with model/render tests. This preserves the product's visual identity and limits risk.
2. **Replace the UI composition.** Rebuild the screens and navigation around a new layout. This would be larger and would introduce new behavior without evidence that the existing architecture needs replacement.

## Constraints and acceptance

- All visible text remains English and contains no emoji.
- The UI must not exceed terminal dimensions; the two bordered panels plus gutter must span the available width exactly.
- The filtered result count and tree category counts must describe the same search result set.
- A successful history import persists once, closes its modal, and reports success.
- `go test ./...`, `go vet ./...`, `cargo test`, and `make build` pass.
