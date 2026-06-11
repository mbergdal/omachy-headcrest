# AGENTS.md

## Commands

- Use the Go version from `go.mod` (`go 1.26.1`); CI also reads `go-version-file: go.mod`.
- Local build: `make build` or `go build -ldflags "-X main.version=dev" -o omachy .`.
- Safe smoke run after building: `./omachy doctor` or `./omachy install --dry-run --quiet`; avoid running plain `./omachy install` unless the user wants host macOS settings changed.
- CI-equivalent checks: `gofmt -l .`, then `go vet ./...`, then `go test ./...`.
- Format changed Go files with `gofmt -w <files>`; CI fails on any `gofmt -l .` output.
- Focused unit test: `go test ./internal/brew -run TestInstall_CaskPassesAdopt` or `go test ./internal/tui -run TestName`.

## Integration Tests

- `make test-integration` requires Apple Silicon, Tart, sshpass, and a prebuilt `omachy-base` VM from `make test-setup`; first setup downloads a large macOS Tahoe image.
- Treat `make test-integration` as expensive; prefer unit tests unless the user explicitly wants a disposable macOS VM install/uninstall cycle.

## Architecture

- `main.go` is the entrypoint: it injects `cmd.Version` from ldflags and assigns `installer.EmbeddedConfigs = Configs` from `configs_embed.go`.
- CLI commands live in `cmd/`; install/uninstall route through Bubble Tea TUI by default and through `tui.RunQuiet` with `--quiet`.
- Install phases are `Preflight -> Backup -> Packages -> Runtimes -> Configs -> System`; the `Runtimes` phase depends on Homebrew-installed `mise` and configures global mise runtimes.
- The executable source of truth for packages and deployed config paths is `internal/manifest/manifest.go`, not the README.
- Config files under `configs/` are embedded via `//go:embed all:configs`; adding a deployable config also requires a `manifest.Configs()` mapping.
- Installer state is persisted at `~/.omachy/state.json`; tests can redirect it with the package-level `statePathOverride` in `internal/installer`.
- Side-effect wrappers are thin packages (`internal/brew`, `internal/shell`); tests usually fake commands by prepending a temp dir to `PATH`, not by adding interfaces.

## Gotchas

- Homebrew cask installs intentionally pass `--adopt` so pre-existing `.app` bundles can be claimed instead of aborting; keep the brew tests covering this behavior.
- Mise runtime setup lives in `internal/installer/runtimes.go`; npm is updated through `mise exec` after Node, not tracked as a separate mise runtime.
- `manifest.Services()` currently starts the Colima brew service during the System phase; do not add `colima start` unless the user explicitly wants the VM started during install.
- Release builds are darwin-only via GoReleaser; `main.version` is set with `-X main.version={{.Version}}`.
