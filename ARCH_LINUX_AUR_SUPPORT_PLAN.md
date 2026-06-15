# Arch Linux + AUR Support Plan

## Goal

Extend Omachy so it supports both macOS and Arch Linux.

- macOS continues using Homebrew.
- Arch Linux uses `pacman` for official repository packages.
- Arch Linux uses an AUR helper, preferably `paru` and then `yay`, for AUR packages.
- Arch Linux replaces AeroSpace/Raycast/macOS defaults with a Hyprland-based desktop setup.

## Current State

The codebase is currently macOS-first.

Important coupling points:

- `internal/brew/` wraps Homebrew.
- `internal/installer/brew.go` installs all packages through Homebrew.
- `internal/manifest/manifest.go` stores Homebrew names, taps, casks, and brew services.
- `internal/preflight/preflight.go` checks macOS, Homebrew, Xcode CLI, Spaces, and AeroSpace.
- `internal/installer/defaults.go` applies macOS defaults, starts brew services, starts AeroSpace, and opens Raycast import.
- `internal/installer/config.go` includes Homebrew-specific zsh plugin paths.
- `cmd/status.go`, `internal/installer/selection.go`, and `internal/uninstaller/uninstaller.go` call Homebrew directly.
- `.goreleaser.yaml` only builds darwin binaries.

## Design Principles

- Keep macOS behavior unchanged unless necessary.
- Add platform-specific behavior behind small abstractions.
- Prefer minimal abstractions over large interface hierarchies.
- Track installed package source in state so uninstall is reliable.
- Do not automatically bootstrap AUR helpers in the first implementation.
- Require an installed AUR helper on Arch.
- Prefer `paru` over `yay` when both are installed.
- Support mixed Arch package lists by routing official packages to `pacman` and everything else to the AUR helper.

## Platform Model

Add a small platform package, likely `internal/platform`.

Suggested types:

```go
type OS string

const (
	OSDarwin OS = "darwin"
	OSArch   OS = "arch"
)

func Current() OS
func IsArchLinux() bool
```

Detection:

- `runtime.GOOS == "darwin"` means macOS.
- `runtime.GOOS == "linux"` plus `/etc/os-release` containing `ID=arch` or `ID_LIKE=arch` means Arch.
- Unsupported platforms should fail preflight with a clear message.

## Manifest Redesign

Replace the Homebrew-specific package shape with platform-aware package metadata.

Suggested structure:

```go
type PackageSource string

const (
	SourceBrew   PackageSource = "brew"
	SourcePacman PackageSource = "pacman"
	SourceAUR    PackageSource = "aur"
	SourceAuto   PackageSource = "auto"
)

type Package struct {
	ID            string
	Name          string
	Label         string
	Description   string
	Platform      platform.OS
	Source        PackageSource
	BrewTap       string
	BrewCask      bool
	Service       string
	SkipIfBinary  string
}
```

Notes:

- `ID` is stable and used for selection/config mapping.
- `Name` is the package-manager-specific package name.
- `SourceAuto` is Arch-only and means try `pacman -Si`, otherwise use AUR.
- `BrewTap` and `BrewCask` are macOS-only fields.
- `Service` should store `colima` on macOS and systemd unit names like `docker.service` on Arch.

Suggested functions:

```go
func Packages(os platform.OS) []Package
func Configs(os platform.OS) []ConfigMapping
func Services(os platform.OS) []Package
```

Keep macOS package IDs stable where possible to reduce churn.

## Arch Package Routing

Implement an Arch package manager package, likely `internal/archpkg`.

It should mirror this zsh helper behavior:

```zsh
if pacman -Si "$pkg"; then
  sudo pacman -S --needed "$pkg"
else
  paru_or_yay -S --needed "$pkg"
fi
```

Suggested functions:

```go
func DetectAURHelper() (string, bool)
func IsOfficialPackage(name string) bool
func IsInstalled(name string) bool
func Install(pkgs []manifest.Package, log func(string)) ([]installer.InstalledPackage, error)
func Uninstall(pkg installer.InstalledPackage, log func(string)) error
```

AUR helper detection:

- Check `pacman -Qi paru`.
- Then check `pacman -Qi yay`.
- Optionally fall back to `shell.Which("paru")` and `shell.Which("yay")`.
- If no helper exists and AUR packages are needed, fail with `No AUR helper found. Install paru or yay first.`

Install behavior:

- Split selected Arch packages into official and AUR groups.
- For `SourcePacman`, always official.
- For `SourceAUR`, always AUR.
- For `SourceAuto`, use `pacman -Si <name>`.
- Install official packages in one batch: `sudo pacman -S --needed <packages...>`.
- Install AUR packages in one batch: `<helper> -S --needed <packages...>`.
- Record packages with the resolved source, not just the requested source.

Uninstall behavior:

- If source is `pacman`, use `sudo pacman -Rns <name>`.
- If source is `aur`, use `<helper> -Rns <name>`.
- If source is unknown legacy data, detect foreign packages with `pacman -Qm <name>`.

## Shell Streaming Fix

`internal/shell.RunStreaming` currently does not attach stdin.

For `sudo pacman` and AUR helpers, add stdin inheritance:

```go
cmd.Stdin = os.Stdin
```

This allows sudo password prompts and AUR helper confirmations to work.

Keep stdout/stderr streaming as-is.

## Package Manager Dispatch

Update the package installation phase so it chooses by platform:

- macOS uses `internal/brew`.
- Arch uses `internal/archpkg`.

Rename `internal/installer/brew.go` to something generic like `packages.go`.

Suggested flow:

```go
func runPackages(p *tea.Program, opts Options) error {
	switch platform.Current() {
	case platform.OSDarwin:
		return runBrewPackages(p, opts)
	case platform.OSArch:
		return runArchPackages(p, opts)
	default:
		return fmt.Errorf("unsupported platform")
	}
}
```

Avoid a large interface unless it makes the implementation smaller.

## State Changes

Current state only stores `Name` and `Cask`.

Extend it:

```go
type InstalledPackage struct {
	ID       string `json:"id,omitempty"`
	Name     string `json:"name"`
	Platform string `json:"platform,omitempty"`
	Source   string `json:"source,omitempty"`
	Cask     bool   `json:"cask,omitempty"`
}
```

Compatibility:

- Existing macOS state with only `name` and `cask` should continue to uninstall through Homebrew.
- New installs should always record platform and source.
- Uninstall should use recorded source when available.

## Arch Package List

Initial pacman or auto packages:

| ID | Arch package | Source |
| --- | --- | --- |
| hyprland | `hyprland` | pacman |
| xdg-desktop-portal-hyprland | `xdg-desktop-portal-hyprland` | pacman |
| waybar | `waybar` | pacman |
| rofi | `rofi-wayland` | pacman |
| mako | `mako` | pacman |
| wl-clipboard | `wl-clipboard` | pacman |
| grim | `grim` | pacman |
| slurp | `slurp` | pacman |
| brightnessctl | `brightnessctl` | pacman |
| playerctl | `playerctl` | pacman |
| polkit-agent | `polkit-kde-agent` | pacman |
| ghostty | `ghostty` | auto |
| neovim | `neovim` | pacman |
| tree-sitter | `tree-sitter` | pacman |
| tmux | `tmux` | pacman |
| starship | `starship` | pacman |
| fzf | `fzf` | pacman |
| docker | `docker` | pacman |
| docker-compose | `docker-compose` | pacman |
| lazygit | `lazygit` | pacman |
| gh | `github-cli` | pacman |
| eza | `eza` | pacman |
| bat | `bat` | pacman |
| yazi | `yazi` | pacman |
| lazydocker | `lazydocker` | auto |
| atuin | `atuin` | pacman |
| zsh-syntax-highlighting | `zsh-syntax-highlighting` | pacman |
| zsh-autosuggestions | `zsh-autosuggestions` | pacman |
| fastfetch | `fastfetch` | pacman |
| mise | `mise` | pacman |
| hack-nerd-font | `ttf-hack-nerd` | pacman |
| jetbrains-mono-nerd-font | `ttf-jetbrains-mono-nerd` | pacman |

Initial AUR packages:

| ID | AUR package | Source |
| --- | --- | --- |
| zen | `zen-browser-bin` | aur |
| 1password | `1password` | aur |
| claude | `claude-desktop-bin` | aur |
| chatgpt | `chatgpt-desktop-bin` | aur |
| opencode | verify package name during implementation | auto or aur |
| zed | verify package name during implementation | auto or aur |
| superhuman | verify availability, otherwise omit | aur or unsupported |

Implementation notes:

- Verify current Arch/AUR package names during implementation.
- If a package is unavailable, omit it from the Arch manifest instead of adding a broken entry.

## Config Changes

Make config mappings platform-specific.

Shared configs:

| Source | Destination |
| --- | --- |
| `tmux/tmux.conf` | `~/.tmux.conf` |
| `starship.toml` | `~/.config/starship.toml` |
| `zed/keymap.json` | `~/.config/zed/keymap.json` |
| `zed/settings.json` | `~/.config/zed/settings.json` |
| `omachy/dev-session.sh` | `~/.config/omachy/dev-session.sh` |
| `yazi/keymap.toml` | `~/.config/yazi/keymap.toml` |
| `yazi/theme.toml` | `~/.config/yazi/theme.toml` |
| `yazi/yazi.toml` | `~/.config/yazi/yazi.toml` |
| `zshrc/.aliases` | `~/.aliases` |

macOS-only configs:

| Source | Destination |
| --- | --- |
| `aerospace/aerospace.toml` | `~/.config/aerospace/aerospace.toml` |
| `Raycast ... .rayconfig` | `~/.config/omachy/raycast.rayconfig` |
| `ghostty/config` | `~/Library/Application Support/com.mitchellh.ghostty/config` |

Arch-only configs to add:

| Source | Destination |
| --- | --- |
| `hypr/hyprland.conf` | `~/.config/hypr/hyprland.conf` |
| `waybar/config` | `~/.config/waybar/config` |
| `waybar/style.css` | `~/.config/waybar/style.css` |
| `mako/config` | `~/.config/mako/config` |
| `rofi/config.rasi` | `~/.config/rofi/config.rasi` |
| `ghostty/config` | `~/.config/ghostty/config` |

## Zsh Integration Changes

`internal/installer/config.go` currently uses Homebrew paths:

```zsh
source $(brew --prefix)/share/zsh-syntax-highlighting/zsh-syntax-highlighting.zsh
source $(brew --prefix)/share/zsh-autosuggestions/zsh-autosuggestions.zsh
```

Add platform-specific shell integration lines.

macOS:

```zsh
source $(brew --prefix)/share/zsh-syntax-highlighting/zsh-syntax-highlighting.zsh
source $(brew --prefix)/share/zsh-autosuggestions/zsh-autosuggestions.zsh
```

Arch:

```zsh
source /usr/share/zsh/plugins/zsh-syntax-highlighting/zsh-syntax-highlighting.zsh
source /usr/share/zsh/plugins/zsh-autosuggestions/zsh-autosuggestions.zsh
```

Keep shared integrations:

```zsh
eval "$(mise activate zsh)"
eval "$(starship init zsh)"
eval "$(fzf --zsh)"
eval "$(atuin init zsh)"
set -o vi
fastfetch
dev() { sh ~/.config/omachy/dev-session.sh "$@"; }
```

## Preflight Changes

Split preflight by platform.

macOS checks remain:

- Architecture
- macOS version
- Homebrew
- Xcode CLI tools
- Separate Spaces
- AeroSpace version

Arch checks:

- Running on Arch Linux
- `pacman` exists
- `sudo` exists
- `git` exists
- `base-devel` installed
- `paru` or `yay` installed
- Hyprland session warning if not currently running
- Optional Docker group warning if Docker is selected and user is not in `docker`

Suggested Arch check commands:

```sh
pacman -Q base-devel
pacman -Qi paru
pacman -Qi yay
```

Failure behavior:

- Missing `pacman`, non-Arch Linux, missing `base-devel`, and missing AUR helper are fatal.
- Not currently running Hyprland should be a warning.
- Docker group membership should be a warning.

## System Phase Changes

Split `runSystem`:

```go
func runSystem(p *tea.Program, opts Options) error {
	switch platform.Current() {
	case platform.OSDarwin:
		return runSystemDarwin(p, opts)
	case platform.OSArch:
		return runSystemArch(p, opts)
	default:
		return fmt.Errorf("unsupported platform")
	}
}
```

macOS keeps current behavior:

- Apply defaults.
- Disable Spotlight shortcuts.
- Restart Dock/SystemUIServer.
- Start brew services.
- Start/reload AeroSpace.
- Open Raycast import.

Arch behavior:

- Do not run macOS defaults.
- Do not open Raycast import.
- Do not start AeroSpace.
- Enable/start systemd services for selected packages, especially `docker.service`.
- Reload Hyprland with `hyprctl reload` if Hyprland is running.
- Log if the user must log out and back in for group changes.

Systemd commands:

```sh
sudo systemctl enable --now docker.service
```

Hyprland reload:

```sh
hyprctl reload
```

Only run `hyprctl reload` if `hyprctl` exists and either `HYPRLAND_INSTANCE_SIGNATURE` is set or `pgrep -x Hyprland` succeeds.

## Package Selection UI

Update `internal/installer/selection.go`.

Current behavior queries Homebrew formulae/casks.

New behavior:

- macOS: keep Homebrew checks.
- Arch: use `pacman -Q <name>` for installed status.
- Arch AUR packages are still visible in the package selector.
- Kind labels should be platform-aware: `Formula`, `Cask`, `Pacman`, `AUR`, or `Auto`.
- App bundle checks in `/Applications` are macOS-only.
- Binary checks can remain cross-platform.

## Status Command

Update `cmd/status.go`.

Current behavior calls `brew.IsInstalled`.

New behavior:

- Load current platform.
- Iterate `manifest.Packages(currentPlatform)`.
- Use appropriate installed check:
- macOS: Homebrew.
- Arch: `pacman -Q`.
- Show package source in output if useful.

## Uninstall Changes

Update `internal/uninstaller/uninstaller.go`.

Services:

- macOS: `brew services stop`.
- Arch: `sudo systemctl disable --now <unit>` for recorded services.

Packages:

- Use state `Source` field.
- macOS legacy packages use Homebrew.
- Arch `pacman` packages use `sudo pacman -Rns`.
- Arch `aur` packages use `paru/yay -Rns`.

Defaults:

- macOS restores defaults.
- Arch skips defaults restore with a log line.

Processes:

- macOS kills `AeroSpace` if Omachy started it.
- Arch should not kill Hyprland.
- Arch can skip process killing initially.

## Command Text and Docs

Update CLI descriptions:

- `cmd/install.go` should not say only macOS system defaults.
- `cmd/uninstall.go` should not say only macOS defaults.
- `internal/tui/splash.go` should render platform-specific text.

README changes:

- Add macOS install section.
- Add Arch install section.
- Document `pacman`, `base-devel`, and AUR helper prerequisites.
- Document Hyprland package/config behavior.
- Document that unsupported packages may require AUR and package names may differ.
- Update architecture section to mention platform-specific package managers.

## Release Changes

Update `.goreleaser.yaml`:

```yaml
goos:
  - darwin
  - linux
goarch:
  - amd64
  - arm64
```

The existing archive `name_template` already includes OS and architecture.

## Test Plan

Add unit tests for package routing:

- `SourceAuto` package found by `pacman -Si` goes to official group.
- `SourceAuto` package not found by `pacman -Si` goes to AUR group.
- `SourcePacman` always uses pacman.
- `SourceAUR` always uses AUR.
- `paru` is preferred over `yay`.
- Missing AUR helper fails when AUR packages are selected.
- Official-only selection does not require AUR helper.

Add package command tests:

- pacman install uses `sudo pacman -S --needed`.
- AUR install uses `paru -S --needed` or `yay -S --needed`.
- pacman uninstall uses `sudo pacman -Rns`.
- AUR uninstall uses helper `-Rns`.

Add platform manifest tests:

- macOS manifest contains AeroSpace and Raycast.
- Arch manifest contains Hyprland and does not contain AeroSpace or Raycast.
- Shared packages appear on both where applicable.
- Arch Ghostty config path is `~/.config/ghostty/config`.
- macOS Ghostty config path remains `~/Library/Application Support/com.mitchellh.ghostty/config`.

Add shell integration tests:

- macOS uses `$(brew --prefix)` zsh plugin paths.
- Arch uses `/usr/share/zsh/plugins/...`.

Add preflight tests:

- Arch detection works from sample `/etc/os-release` content.
- Missing AUR helper is fatal.
- Missing Hyprland session is warning only.

Add state/uninstall tests:

- Legacy macOS state still uninstalls with Homebrew.
- Arch state with source `pacman` uninstalls with pacman.
- Arch state with source `aur` uninstalls with AUR helper.

## Verification Commands

Run formatting:

```sh
gofmt -w <changed-go-files>
gofmt -l .
```

Run checks:

```sh
go vet ./...
go test ./...
```

Build locally:

```sh
go build -ldflags "-X main.version=dev" -o omachy .
```

Safe smoke tests:

```sh
./omachy doctor
./omachy install --dry-run --quiet
```

On Arch, smoke test should include:

```sh
./omachy doctor
./omachy install --dry-run --quiet
```

Do not run plain `./omachy install` unless explicitly approved because it changes host system packages and configs.

## Suggested Implementation Order

1. Add platform detection.
2. Refactor manifest into platform-specific package/config functions while keeping macOS output unchanged.
3. Add `internal/archpkg` with pacman/AUR helper detection and package splitting.
4. Update `shell.RunStreaming` to inherit stdin.
5. Refactor package install phase to dispatch to Homebrew or Arch.
6. Extend state with platform/source and preserve legacy macOS compatibility.
7. Update uninstall/status/package selector to dispatch by platform.
8. Split preflight into macOS and Arch checks.
9. Split system phase into macOS and Arch behavior.
10. Add Arch Hyprland-related configs.
11. Update zsh integration paths per platform.
12. Update splash text, command descriptions, README, and GoReleaser config.
13. Add tests after each refactor step to keep behavior stable.

## Open Decisions

- Confirm whether `paru` should be required or only preferred over `yay`.
- Confirm whether the installer should ever bootstrap an AUR helper automatically.
- Confirm the preferred Arch launcher: `rofi-wayland`, `wofi`, or another tool.
- Confirm whether notification setup should use `mako` or `swaync`.
- Confirm whether Waybar should be part of the default Hyprland setup.
- Verify current AUR names for Zed, opencode, Claude, ChatGPT, Zen Browser, 1Password, and Superhuman before adding them.
