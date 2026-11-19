# Release Checklist

## Build and checks

- [ ] Run the CI workflow on Windows, Linux, and macOS.
- [ ] Run `go test ./...` and `go vet ./...` from `cli/`.
- [ ] Run `cargo test --manifest-path engine/Cargo.toml`.
- [ ] Run `make build` on the release platform and verify the expected executable in `bin/`.
- [ ] Launch the executable in a clean profile and confirm it opens directly to the Git tree.

## First-run behavior

- [ ] Confirm `~/.snipbox/commands.toml` is created with the expected starter commands.
- [ ] Check that the default Git command is selected and the tree is expanded.
- [ ] Check search, category switching, action-menu shortcuts, copy, and variable entry.
- [ ] Confirm dangerous execution and deletion require explicit confirmation.
- [ ] Confirm forms reject invalid fields and duplicate command text without overwriting saved data.

## Package and publish

- [ ] Include the executable and English README for the target operating system.
- [ ] State the operating system and architecture in the artifact name.
- [ ] Verify no user command library, shell history, credentials, or local config is packaged.
- [ ] Record the commit, version, and SHA-256 checksum with the release.
- [ ] Attach backup and restore instructions for `commands.toml`.
