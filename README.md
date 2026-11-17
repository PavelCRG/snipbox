<div align="center">

# Snipbox

### Your commands, ready when you are.

Fuzzy-find, customize, and run the commands you use every day — from one focused terminal workspace.

[Getting started](#getting-started) · [Features](#made-for-the-terminal) · [Build](#build-from-source) · [Keyboard shortcuts](#keyboard-shortcuts)

</div>

---

## About

Snipbox is a cross-platform command library with a fast Rust search engine and a polished Go terminal interface. Search snippets as you type, explore a navigable command tree, fill in template values, and copy or run commands from one focused workspace. Your library stays in a plain TOML file on your machine.

The interface is designed around a simple loop: find a command, understand what it does, then copy or run it. A welcome dashboard opens on first launch, with five starter collections: **Git**, **PowerShell**, **Docker**, **Linux/Bash**, and **Network**. Browse the library as a searchable category and subcategory tree. The terminal interface and starter library are in English.

Created by **Pavel Slepcov**. Press **Ctrl+O** to open About inside the application.

## Made for the terminal

- **Live fuzzy search** across command names, descriptions, tags, and command text.
- **A two-panel workspace** with category tabs, favorites, usage counts, highlighted matches, and command details.
- **A welcome dashboard and expandable tree** for browsing categories and subcategories.
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

The welcome dashboard shows command, category, and favorite counts. Select a category with the arrow keys and **Enter**, or use **1–5** to jump directly. Choose **0** to browse the complete command tree. Press **Ctrl+W** in the tree to return to the dashboard when the search is empty.

Open it with a search already in place, list the library, or locate its data file:

```sh
snipbox docker logs
snipbox --list
snipbox --path
```

On the first launch Snipbox creates `~/.snipbox/commands.toml` with 40 English starter commands: 12 Git, 10 PowerShell, and 6 each for Docker, Linux/Bash, and Network. Every command has a category, subcategory, description, tags, and command text. You can edit the file in any text editor, or use **Ctrl+A** and **Ctrl+E** inside Snipbox.

### Command templates

Use `{{name}}` for a value with no preset, or `{{name:default}}` to offer a useful starting value:

```sh
git switch -c {{branch:feature/new-work}}
docker logs -f {{container:app}}
git commit -m "{{message:Update files}}"
```

Press **Ctrl+R** and enter values in the prompt. Snipbox renders the command before running it, so you can review the exact command. Patterns such as `rm -rf`, `reset --hard`, `drop`, `format`, and `mkfs` open an extra confirmation prompt.

## Keyboard shortcuts

| Key | Action |
|:---|:---|
| `↑` / `↓` | Move through results |
| Type | Search the library |
| `Enter` | Copy the selected command |
| `Ctrl+R` | Fill template values and run |
| `Tab` | Cycle through categories |
| `Ctrl+A` | Create a command |
| `Ctrl+E` | Edit the selected command |
| `Ctrl+H` | Import a command from shell history |
| `0`–`5` | Choose a dashboard section (`0` opens the complete tree) |
| `Left` / `Right` or `Space` | Collapse or expand a tree branch |
| `Ctrl+W` | Return to the welcome dashboard when search is empty |
| `/` | Open search and browse |
| `Ctrl+O` | Open About |
| `?` | Show keyboard shortcuts |
| `Esc` | Close a dialog or quit |

## Your library, your machine

Snipbox stores commands, tags, favorites, and usage counts in `~/.snipbox/commands.toml`. The file is written through a temporary file and rename, and is readable by ordinary TOML tools. Shell history is read locally when you choose **Ctrl+H**; Snipbox does not transmit it.

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
