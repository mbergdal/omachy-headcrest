# Omachy

**Omarchy for the rest of us.**

https://github.com/user-attachments/assets/ba95b0f2-8d99-4987-b2f5-fa932f45a259

Omachy brings the [Omarchy](https://omakub.org/) experience to macOS — a tiling WM, custom menu bar, terminal emulator, editor, and sane system defaults, all configured in one shot. For people who'd rather be on Linux but can't.

**[Homepage](https://omachy.org)** · **[Usage Guide](https://omachy.org/guide.html)**

## What Gets Installed

| Tool | Type | Description |
|------|------|-------------|
| [AeroSpace](https://github.com/nikitabobko/AeroSpace) | Cask | Tiling window manager |
| [Ghostty](https://ghostty.org/) | Cask | Terminal emulator |
| [Neovim](https://neovim.io/) | Formula | Text editor (LazyVim starter cloned if no config exists) |
| [tree-sitter](https://tree-sitter.github.io/tree-sitter/) | Formula | Parser generator for syntax highlighting |
| [Tmux](https://github.com/tmux/tmux) | Formula | Terminal multiplexer (TPM + plugins) |
| [Zed](https://zed.dev/) | Cask | Graphical code editor |
| [Zen Browser](https://zen-browser.app/) | Cask | Web browser |
| [Raycast](https://www.raycast.com/) | Cask | Launcher and automation |
| [1Password](https://1password.com/) | Cask | Password manager |
| [Starship](https://starship.rs/) | Formula | Cross-shell prompt |
| [fzf](https://github.com/junegunn/fzf) | Formula | Fuzzy finder |
| [Docker](https://www.docker.com/) | Formula | Docker CLI |
| [Docker Compose](https://docs.docker.com/compose/) | Formula | Compose plugin/CLI |
| [Colima](https://github.com/abiosoft/colima) | Formula | Docker runtime on macOS; installed as a brew service |
| [eza](https://github.com/eza-community/eza) | Formula | Modern `ls` replacement |
| [bat](https://github.com/sharkdp/bat) | Formula | `cat` replacement with syntax highlighting |
| [Lazygit](https://github.com/jesseduffield/lazygit) | Formula | Git TUI |
| [GitHub CLI](https://cli.github.com/) | Formula | `gh` command-line client |
| [opencode](https://github.com/sst/opencode) | Formula | AI coding agent CLI |
| [Yazi](https://yazi-rs.github.io/) | Formula | Terminal file manager |
| [Lazydocker](https://github.com/jesseduffield/lazydocker) | Formula | Docker TUI |
| [Atuin](https://atuin.sh/) | Formula | Shell history search & sync |
| [zsh-syntax-highlighting](https://github.com/zsh-users/zsh-syntax-highlighting) | Formula | Fish-like syntax highlighting for Zsh |
| [zsh-autosuggestions](https://github.com/zsh-users/zsh-autosuggestions) | Formula | Fish-like inline suggestions from history |
| [fastfetch](https://github.com/fastfetch-cli/fastfetch) | Formula | System info display |
| [mise](https://mise.jdx.dev/) | Formula | Runtime manager for development languages and tools |
| [Hack Nerd Font](https://www.nerdfonts.com/) | Cask | Nerd Font for terminal/icons |
| [JetBrains Mono](https://www.jetbrains.com/lp/mono/) | Cask | Monospace font for terminal |

**Language runtimes** are installed and set globally through `mise`:

| Runtime | Version |
|---------|---------|
| Node.js | `latest` |
| Python | `latest` |
| Bun | `latest` |
| pnpm | `latest` |
| Erlang | `latest` |
| Elixir | `latest` |

After Node is installed, npm is updated with `mise exec -- npm install -g npm@latest`.

## What Gets Configured

**Config files** are deployed from the embedded `configs/` directory:

| Source | Destination |
|--------|------------|
| `aerospace/aerospace.toml` | `~/.config/aerospace/aerospace.toml` |
| `ghostty/config` | `~/Library/Application Support/com.mitchellh.ghostty/config` |
| `tmux/tmux.conf` | `~/.tmux.conf` *(NeverOverwrite)* |
| `starship.toml` | `~/.config/starship.toml` |
| `zed/keymap.json` | `~/.config/zed/keymap.json` |
| `zed/settings.json` | `~/.config/zed/settings.json` |
| `omachy/dev-session.sh` | `~/.config/omachy/dev-session.sh` |
| `yazi/keymap.toml` | `~/.config/yazi/keymap.toml` |
| `yazi/theme.toml` | `~/.config/yazi/theme.toml` |
| `yazi/yazi.toml` | `~/.config/yazi/yazi.toml` |
| `zshrc/.aliases` | `~/.aliases` |

**These files are never overwritten:**

| File | Behavior |
|------|----------|
| `~/.zshrc` | Never replaced. Omachy only appends a managed block (between clearly marked markers) for shell integrations. Everything else in your `.zshrc` is untouched. |
| `~/.tmux.conf` | Never replaced. If the file already exists, deployment is skipped entirely — your tmux config is preserved as-is. |
| `~/.config/nvim/` | Never replaced. LazyVim starter is only cloned if no Neovim config exists. |

Additionally, the installer:

- Injects shell integrations and `configs/zshrc/extras` (sources `~/.aliases`, Yazi `y` helper, `EDITOR=nvim`) into `~/.zshrc` via a managed block — existing content is preserved
- Clones [LazyVim starter](https://github.com/LazyVim/starter) to `~/.config/nvim/` if no Neovim config exists
- Clones [TPM](https://github.com/tmux-plugins/tpm) to `~/.tmux/plugins/tpm` if not already installed
- Deploys Zed settings/keymaps, Yazi configuration, and shell aliases

**macOS system defaults** are adjusted:

- Auto-hide Dock with zero delay and fast animation
- Set Dock icon size to 48 and hide recent apps
- Disable MRU Spaces reordering (important for tiling WM)
- Disable window open/close animations
- Fastest key repeat rate with shortest initial delay
- Disable press-and-hold for key repeat
- Hide desktop widgets
- Show all file extensions
- Scale minimize effect

## Prerequisites

- macOS 13 (Ventura) or later
- [Homebrew](https://brew.sh)
- Xcode Command Line Tools (`xcode-select --install`) — ensure they're up to date via Software Update

## Installation

```bash
brew tap dough654/omachy
brew install omachy
omachy install
```

### Running a downloaded release

Manually downloaded GitHub release binaries may be blocked by macOS Gatekeeper because they are not Apple-notarized. For your own trusted release binary, remove the quarantine attribute before running:

```bash
xattr -dr com.apple.quarantine ./omachy
chmod +x ./omachy
./omachy install
```

If you extracted the release into a folder, remove quarantine from the whole folder:

```bash
xattr -dr com.apple.quarantine /path/to/omachy-folder
```

### Building from source

```bash
go build -o omachy .
./omachy install
```

## Usage

### `omachy install`

Runs the full installation through an interactive TUI with six phases:

1. **Preflight** — checks architecture, macOS version, Homebrew, Xcode CLI tools, and Spaces settings
2. **Backup** — copies any existing config files to `~/.omachy/backups/<timestamp>/`
3. **Packages** — taps Homebrew repos and installs all packages
4. **Runtimes** — installs latest Node.js, Python, Bun, pnpm, Erlang, and Elixir through `mise`
5. **Configs** — deploys embedded config files to their destinations
6. **System** — applies macOS defaults, starts brew services such as Colima, and prompts for AeroSpace accessibility permissions

**Flags:**

| Flag | Description |
|------|-------------|
| `--dry-run` | Show what would be done without making changes |
| `--force` | Overwrite existing configs without prompting |
| `--skip-backup` | Skip backing up existing configs |
| `--verbose` | Show detailed output |
| `--named-workspaces` | Use named workspaces (D/W/M/E/S) with app-to-workspace rules instead of numbered 1–9 |
| `--quiet` | Run without the TUI (log to stdout) |

### `omachy uninstall`

Reverses the installation — stops started processes, removes configs, removes mise runtimes, uninstalls packages, and restores original macOS defaults from saved state.

**Flags:**

| Flag | Description |
|------|-------------|
| `--dry-run` | Show what would be done without making changes |
| `--keep-configs` | Keep deployed config files |
| `--keep-packages` | Keep installed Homebrew packages |
| `--quiet` | Run without the TUI (log to stdout) |

### `omachy doctor`

Runs preflight checks and reports system readiness without installing anything.

### `omachy status`

Shows current installation status: which packages are installed, whether deployed configs have been modified (drift detection via SHA-256 checksums), and backup location.

### `omachy version`

Prints the version string.

## Architecture

```
.
├── main.go                    # Entry point, wires embedded FS
├── configs_embed.go           # go:embed directive for configs/
├── configs/                   # Embedded config files (deployed at install time)
│   ├── aerospace/
│   ├── ghostty/
│   ├── omachy/
│   └── tmux/
├── cmd/                       # Cobra CLI commands
│   ├── root.go                # Root command + version var
│   ├── install.go             # Install command + flags
│   ├── uninstall.go           # Uninstall command + flags
│   ├── doctor.go              # System readiness checker
│   ├── status.go              # Installation status + drift detection
│   └── version.go             # Version printer
└── internal/
    ├── manifest/              # Package and config declarations (pure data)
    ├── checksum/              # SHA-256 hashing for files and directories
    ├── backup/                # File/directory backup with tilde expansion
    ├── preflight/             # System readiness checks
    ├── installer/             # Install orchestrator + phase implementations
    │   ├── installer.go       # Phase runner
    │   ├── state.go           # JSON state persistence (~/.omachy/state.json)
    │   ├── preflight_phase.go # Preflight phase
    │   ├── brew.go            # Package installation phase
    │   ├── config.go          # Backup + config deployment phases
    │   ├── defaults.go        # macOS defaults + service start phase
    │   └── services.go        # (placeholder)
    ├── uninstaller/           # Uninstall orchestrator (reverse of installer)
    ├── shell/                 # exec.Command wrapper (Run, RunStreaming, Which)
    ├── brew/                  # Homebrew CLI wrapper (tap, install, services)
    └── tui/                   # Bubbletea terminal UI
        ├── app.go             # Root model + state machine
        ├── events.go          # Message types (PhaseStarted, LogLine, etc.)
        ├── splash.go          # Pre-install splash screen
        ├── confirm.go         # User confirmation prompts
        ├── quiet.go           # Non-TUI stdout logging mode
        ├── phases.go          # Left panel phase list with status icons
        ├── header.go          # Top bar with title and progress percentage
        ├── output.go          # Scrollable log viewport
        ├── help.go            # Bottom key-binding hints
        └── styles.go          # Lipgloss color palette and styles
```

### Key Design Decisions

- **Embedded configs** — Config files are compiled into the binary via `go:embed`, making the tool a single self-contained binary with no external file dependencies.
- **State tracking** — `~/.omachy/state.json` tracks installed packages, deployed config checksums, original macOS defaults, backup paths, and running services. This enables clean uninstall and drift detection.
- **Idempotent operations** — Package installs check whether a package is already installed before running `brew install`. Taps are deduplicated.
- **Reversible by default** — Existing configs are backed up before overwriting. Original macOS defaults are read and saved before being changed. The uninstaller restores both.
- **No interfaces or mocks** — The codebase favors simple functions over abstraction layers. Shell calls go through `internal/shell/` as a thin wrapper but aren't abstracted behind interfaces.

### State File

The installer writes `~/.omachy/state.json` with the following structure:

```json
{
  "installed_packages": [
    {"name": "neovim"},
    {"name": "ghostty", "cask": true}
  ],
  "installed_taps": ["nikitabobko/tap", "FelixKratz/formulae"],
  "deployed_configs": {
    "/Users/you/.tmux.conf": "sha256hash",
    "/Users/you/.config/starship.toml": "sha256hash"
  },
  "original_defaults": {
    "com.apple.dock:autohide": "-bool:false",
    "com.apple.dock:tilesize": "-int:64"
  },
  "backup_path": "/Users/you/.omachy/backups/20260314-182007",
  "running_processes": ["AeroSpace"]
}
```

## Testing

```bash
# Run all tests
go test ./...

# Verbose
go test ./... -v
```

Tests cover the pure logic and I/O packages: manifest data, checksum computation, backup file operations, state serialization, TUI components (phases, header, help, splash, app state machine), and preflight logic. Shell-heavy wrappers (brew, system commands) are not unit tested — they're thin wrappers best verified by running the tool.

### CI coverage

- `ci.yml` runs on pull requests and pushes to `master`: `gofmt` check, `go vet ./...`, and `go test ./...`.

## License

Open source. See [LICENSE](LICENSE) for details.
