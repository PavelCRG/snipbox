<div align="center">

# Snipbox

### A personal library for terminal commands you reuse.

Fuzzy-find, customize, and run the commands you use every day — from one focused terminal workspace.

[Getting started](#getting-started) · [Features](#made-for-the-terminal) · [Build](#build-from-source) · [Keyboard shortcuts](#keyboard-shortcuts)

</div>

---

## About

Snipbox is a cross-platform command library for people who work in a terminal and want their dependable commands to be easy to find, understand, and reuse. It brings search, documentation, templates, favorites, and execution into one keyboard-first workspace instead of scattering useful commands across shell history, notes, and old project files.

The interface follows a simple workflow: find a command, inspect its purpose and parameters, fill in any values, then copy or run the result. Snipbox opens directly to the **Git** collection with its subcategories expanded. Use the category tabs to move between topics, or switch to **All** to browse the entire library as a searchable tree. Click a command to inspect it; right-click it or press **M** to open its action menu.

The Rust engine provides fuzzy matching and template parsing; the Go application handles the terminal UI and local integrations. Commands are stored as ordinary TOML on your own machine. Snipbox does not require an account or send your library anywhere; shell-history import and Git sync happen only when you choose them.

Created by **Pavel Slepcov**. Press **Ctrl+O** to open About inside the application.

## Made for the terminal

- **Live fuzzy search** across command names, descriptions, tags, and command text.
- **A two-panel workspace** with category tabs, favorites, usage counts, highlighted matches, and command details.
- **Color-coded command tree** with visible branch connectors, expandable categories, and a contextual action menu.
- **Execution context** showing supported shells, operating systems, and required tools for each starter command.
- **Per-command working directory** with a checked preview before a command starts.
- **Reusable templates** with editable prompts and defaults, such as `{{branch:feature/new-work}}`.
- **Create and edit snippets** in the interface; changes are saved immediately to your local TOML library.
- **Copy or run commands** with captured output and usage tracking. Destructive patterns require an explicit confirmation.
- **Bring in shell history** from Bash, Zsh, or PowerShell without uploading it anywhere.
- **Keep your library portable** with a local Git commit and push to a remote you configure.

## Getting started

Launch the interactive workspace:

```sh
snipbox
```

Snipbox opens directly in the two-panel command browser with **Git** selected. Use **Tab** to cycle collections, type to filter commands, and click a command or move with the arrow keys to inspect its details. Press **M** to open its action menu. Press **/** to search the full library.

Open it with a search already in place, list the library, or locate its data file:

```sh
snipbox docker logs
snipbox --list
snipbox --path
```

On the first launch Snipbox creates `~/.snipbox/commands.toml` with 60 English starter commands: 16 Git, 14 PowerShell, and 10 each for Docker, Linux/Bash, and Network. The collection covers everyday source control, Windows administration, containers, Linux troubleshooting, and network diagnostics. Each entry has a category, subcategory, practical description, searchable tags, and a command you can adapt with template variables. You can edit the file in any text editor, or use **Ctrl+A** and **Ctrl+E** inside Snipbox. Existing libraries are upgraded by adding missing starter entries while preserving custom commands and saved preferences.

### Command templates

Use `{{name}}` for a value with no preset, or `{{name:default}}` to offer a useful starting value:

```sh
git switch -c {{branch:feature/new-work}}
docker logs -f {{container:app}}
git commit -m "{{message:Update files}}"
```

Press **Ctrl+R** and enter values in the prompt. Snipbox then shows the exact command, shell, and working directory; press **Enter** to run it or **Esc** to cancel. A relative working directory is resolved from the folder where Snipbox was started. When the command exits, Snipbox reports its exit code. The engine checks command context for destructive operations such as recursive forced removal, hard Git resets, filesystem formatting, SQL `DROP`, PowerShell removal, and container or cluster deletion. It ignores quoted examples and checks the rendered command after template values are inserted. This heuristic cannot understand every shell or script; confirmation adds a review step, not a sandbox or permission limit.

## Keyboard shortcuts

| Key | Action |
|:---|:---|
| `↑` / `↓` | Move through results |
| Type | Search the library |
| `Enter` | Copy the selected command |
| `M` | Open actions for the selected command |
| Right-click | Open actions for the command under the pointer |
| `Ctrl+R` | Review, then run the selected command |
| `Tab` | Cycle through categories |
| `Ctrl+A` | Create a command |
| `Ctrl+E` | Edit the selected command |
| `Ctrl+H` | Import a command from shell history |
| `Ctrl+F` / `Ctrl+D` | Toggle favorite / request deletion |
| `Left` / `Right` or `Space` | Collapse or expand a tree branch |
| `/` | Open search and browse |
| `Ctrl+O` | Open About |
| `?` | Show keyboard shortcuts |
| `Esc` | Close a dialog or quit |

## Your library, your machine

Snipbox stores commands, tags, favorites, and usage counts in `~/.snipbox/commands.toml`. The file is written through a temporary file and rename, and is readable by ordinary TOML tools. Shell history is read locally when you choose **Ctrl+H**; Snipbox does not transmit it.

Back up the library by copying that TOML file somewhere safe. On Linux or macOS:

```sh
cp ~/.snipbox/commands.toml ~/.snipbox/commands.toml.backup
cp ~/.snipbox/commands.toml.backup ~/.snipbox/commands.toml  # restore
```

On Windows PowerShell:

```powershell
$library = Join-Path $HOME '.snipbox\commands.toml'
Copy-Item $library "$library.backup"
Copy-Item "$library.backup" $library  # restore
```

To sync the library with a Git remote, initialize the Snipbox data directory and add your remote once:

```sh
git -C ~/.snipbox init
git -C ~/.snipbox remote add origin <your-private-or-public-repository>
snipbox --sync
```

`--sync` stages the local library, creates a `sync` commit, and pushes to the configured remote. Your Git credentials and remote access remain under your control.

## Architecture

```text
┌───────────────────────────────────┐       C ABI · UTF-8 JSON       ┌────────────────────────────┐
│ Go application                    │ ─────────────────────────────▶ │ Rust engine                │
│                                   │                                 │                            │
│ Bubble Tea · Lipgloss · Cobra     │ ◀───────────────────────────── │ SkimMatcherV2 · templates  │
│ TOML · clipboard · shell history  │         JSON results           │ dangerous-pattern checks   │
└───────────────────────────────────┘                                 └────────────────────────────┘
             cli/                                                               engine/
```

The Rust crate builds as a static C library. Go calls it through the isolated `internal/bridge` cgo package; Rust returns UTF-8 JSON strings, and Go releases each returned string with `snipbox_free_string`. The Go application owns the terminal interface, local storage, clipboard, history, and Git integration.

## Build from source

Build on the target operating system with its native Go and Rust toolchains. Go's cgo support and a C linker are required.

```sh
make build
```

Launch it from the repository root:

```powershell
# Windows
.\bin\snipbox.exe
.\bin\snipbox.exe git
```

```sh
# Linux / macOS
./bin/snipbox
./bin/snipbox git
```

Inside the interface, press **Ctrl+O** to open About and **Esc** to quit. Run the Rust checks separately with:

```sh
cargo test --manifest-path engine/Cargo.toml
cargo build --release --manifest-path engine/Cargo.toml
```

### Platform notes

| Platform | Native toolchain |
|:---|:---|
| Linux | Go, Rust, GNU Make, and a C compiler. The Makefile requests static external linking. |
| macOS | Go with cgo, Rust, GNU Make, and Xcode Command Line Tools. macOS system libraries remain dynamically linked, as required by the platform. |
| Windows | Go with cgo, GNU Make, MinGW-w64 in `PATH`, and `rustup target add x86_64-pc-windows-gnu`. Snippets run in PowerShell. |

Each target is built natively; the Rust static library and C linker must target the same platform and architecture as Go. A single executable is produced per target. Linux can use static external linking when the host has the required static system libraries; macOS does not support fully static executables.

## Safety

Running a snippet executes it through the current platform shell with the permissions of your user account. The destructive-command check is a warning and confirmation step, not a sandbox; review commands before running them. Shell history import and Git sync happen only on your machine and use the tools and remote you configure.

<div align="center">

**Find it. Understand it. Run it.**

</div>
