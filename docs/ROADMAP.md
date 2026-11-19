# Snipbox Improvement Plan

The plan keeps Snipbox focused on its terminal workflow: open the Git command tree directly, find a command, and act on it without a second welcome screen or repeated instructions.

## Completed

1. [x] Open directly in the expanded Git command tree.
2. [x] Keep the welcome dashboard and splash screen out of the normal launch flow.
3. [x] Browse commands through expandable category and subcategory rows.
4. [x] Keep the two-panel layout within the terminal bounds and show a resize hint for small windows.
5. [x] Put copy, run, favorite, edit, and delete actions in a focused command menu.
6. [x] Show single-key action choices in the command menu and preserve confirmation flows for risky actions.

## Next

7. [ ] Make selected-command context and category visible in the action menu without repeating its full description.
8. [ ] Improve empty states for a category with no commands and a search with no matches.
9. [ ] Keep the selected command visible and stable when changing categories or clearing a search.
10. [ ] Make search feedback show the number of matches and the active scope clearly.
11. [ ] Add inline validation and field-specific error messages to command forms.
12. [ ] Prevent saving duplicate commands with the same command text and offer to open the existing entry.
13. [ ] Make favorite state and usage count updates consistent after copy and run.
14. [ ] Add a review step to shell-history import so users can inspect a command before saving it.
15. [ ] Improve command-details scrolling and indicate when more content is available.
16. [ ] Keep destructive-action dialogs explicit about the selected command and the effect of confirmation.
17. [ ] Make status messages consistent, brief, and visible for the same amount of time.
18. [ ] Add reliable export and restore instructions for the local TOML command library.
19. [ ] Verify clean builds and keyboard behavior on Windows, Linux, and macOS.
20. [ ] Document a release checklist covering the binary, default command library, and first-run experience.

## Quality bar

- English-only interface text and no emoji.
- No repeated welcome or branding screen in the startup path.
- Modals own keyboard input while open and fit within the terminal.
- Existing safe-run, delete-confirmation, and atomic-storage behavior stays intact.
- Every behavior change gets a focused regression test before implementation.
