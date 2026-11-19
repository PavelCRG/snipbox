# Snipbox Improvement Plan

The plan keeps Snipbox focused on its terminal workflow: open the Git command tree directly, find a command, and act on it without a second welcome screen or repeated instructions.

## Completed

1. [x] Open directly in the expanded Git command tree.
2. [x] Keep the welcome dashboard and splash screen out of the normal launch flow.
3. [x] Browse commands through expandable category and subcategory rows.
4. [x] Keep the two-panel layout within the terminal bounds and show a resize hint for small windows.
5. [x] Put copy, run, favorite, edit, and delete actions in a focused command menu.
6. [x] Show single-key action choices in the command menu and preserve confirmation flows for risky actions.

7. [x] Show the selected command and category in its action menu without repeating the description.
8. [x] Explain empty categories, empty libraries, and searches with no matches.
9. [x] Keep the selected command when it remains available after changing tabs or clearing search.
10. [x] Show matching count, active category scope, and category total in the header.
11. [x] Validate command fields inline and keep the form open until required values are fixed.
12. [x] Detect duplicate command text and offer to open the existing entry.
13. [x] Use one persistent counter for successful copies and run attempts; restore the old value if saving fails.
14. [x] Preview shell-history commands before importing them.
15. [x] Indicate when command details can scroll up or down.
16. [x] Name the command and permanent effect in deletion confirmation.
17. [x] Show feedback messages for a consistent two seconds; validation remains visible beside invalid fields.
18. [x] Document backup and restore steps for the TOML library.
19. [x] Add a Windows, Linux, and macOS CI matrix for builds, Go checks, and Rust tests.
20. [x] Add a release checklist for builds, first-run behavior, packaging, and data safety.

## Quality bar

- English-only interface text and no emoji.
- No repeated welcome or branding screen in the startup path.
- Modals own keyboard input while open and fit within the terminal.
- Existing safe-run, delete-confirmation, and atomic-storage behavior stays intact.
- Every behavior change gets a focused regression test before implementation.

The cross-platform workflow runs after it is pushed to GitHub. Local verification has been performed on Windows; remote Linux and macOS results are reported by Actions.
