package installer

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/dough654/Omachy/internal/checksum"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dough654/Omachy/internal/backup"
	"github.com/dough654/Omachy/internal/manifest"
	"github.com/dough654/Omachy/internal/shell"
	"github.com/dough654/Omachy/internal/tui"
)

// EmbeddedConfigs is set by main to provide access to the embedded filesystem.
var EmbeddedConfigs embed.FS

func runBackup(p *tea.Program, opts Options) error {
	log := func(text string) {
		p.Send(tui.LogLine{Text: text})
	}

	if opts.SkipBackup {
		log("==> Skipping backup (--skip-backup)")
		return nil
	}

	state, err := LoadState()
	if err != nil {
		return fmt.Errorf("load state: %w", err)
	}

	// Only back up once — the pre-Omachy state. If a backup already exists
	// in state, the user's originals are already saved.
	if state.BackupPath != "" {
		log(fmt.Sprintf("==> Original backup already exists: %s", state.BackupPath))
		log("    Skipping backup (originals are preserved)")
		return nil
	}

	log("==> Backing up existing configs (pre-Omachy state)")

	var destPaths []string
	for _, cfg := range selectedConfigMappings(opts) {
		destPaths = append(destPaths, cfg.Dest)
	}
	if len(destPaths) == 0 {
		log("==> No selected package configs to back up")
		return nil
	}

	if opts.DryRun {
		for _, d := range destPaths {
			log(fmt.Sprintf("    Would back up %s (if exists)", d))
		}
		return nil
	}

	backupPath, err := backup.Run(destPaths, log)
	if err != nil {
		return err
	}

	if backupPath != "" {
		state.BackupPath = backupPath
		if err := SaveState(state); err != nil {
			return fmt.Errorf("save state: %w", err)
		}
	}

	return nil
}

func runConfigs(p *tea.Program, opts Options) error {
	log := func(text string) {
		p.Send(tui.LogLine{Text: text})
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	state, err := LoadState()
	if err != nil {
		return fmt.Errorf("load state: %w", err)
	}

	configs := selectedConfigMappings(opts)
	if len(configs) == 0 {
		log("==> No selected package configs to deploy")
	}

	// Swap to named workspace configs if requested
	if opts.NamedWorkspaces {
		for i := range configs {
			if configs[i].Source == "aerospace/aerospace.toml" {
				configs[i].Source = "aerospace/aerospace-named.toml"
			}
		}
	}

	for i, cfg := range configs {
		dest := expandHome(cfg.Dest, home)

		// Skip configs marked NeverOverwrite if destination already exists
		// Use Lstat to detect symlinks (even broken ones) as "exists"
		if cfg.NeverOverwrite && !opts.Force {
			if _, err := os.Lstat(dest); err == nil {
				log(fmt.Sprintf("==> Skipping %s (existing config found, use --force to overwrite)", cfg.Dest))
				p.Send(tui.ProgressUpdate{Percent: configProgress(i, len(configs))})
				continue
			}
		}

		if opts.DryRun {
			log(fmt.Sprintf("==> Would deploy %s → %s", cfg.Source, cfg.Dest))
			continue
		}

		log(fmt.Sprintf("==> Deploying %s → %s", cfg.Source, cfg.Dest))

		if cfg.IsDir {
			if err := deployDir(cfg.Source, dest, cfg.Mode); err != nil {
				return fmt.Errorf("deploy %s: %w", cfg.Source, err)
			}
		} else {
			if err := deployFile(cfg.Source, dest, fs.FileMode(cfg.Mode)); err != nil {
				return fmt.Errorf("deploy %s: %w", cfg.Source, err)
			}
		}

		// Compute checksum for state tracking
		hash, _ := checksum.Path(dest)
		state.DeployedConfigs[dest] = hash

		p.Send(tui.ProgressUpdate{Percent: configProgress(i, len(configs))})
	}

	// Clean up legacy AeroSpace config location to avoid ambiguity
	legacyAerospace := filepath.Join(home, ".aerospace.toml")
	if opts.packageSelected("nikitabobko/tap/aerospace") {
		if _, err := os.Stat(legacyAerospace); err == nil {
			if !opts.DryRun {
				log("==> Removing legacy ~/.aerospace.toml (AeroSpace uses ~/.config/aerospace/)")
				os.Remove(legacyAerospace)
				// Also remove from deployed configs if we previously tracked it
				delete(state.DeployedConfigs, legacyAerospace)
			} else {
				log("==> Would remove legacy ~/.aerospace.toml to avoid config ambiguity")
			}
		}
	}

	if opts.packageSelected("neovim") {
		nvimDir := filepath.Join(home, ".config", "nvim")
		if _, err := os.Stat(nvimDir); os.IsNotExist(err) {
			if opts.DryRun {
				log("==> Would install LazyVim.nvim")
			} else {
				log("==> Installing LazyVim.nvim")
				if err := shell.RunStreaming("git", []string{
					"clone", "https://github.com/LazyVim/starter", nvimDir,
				}, log); err != nil {
					log(fmt.Sprintf("    Warning: failed to clone LazyVim.nvim: %v", err))
				}
			}
		} else {
			log("    Neovim config already exists, skipping LazyVim.nvim")
		}
	} else {
		log("==> Skipping LazyVim.nvim (Neovim not selected)")
	}

	// Install TPM and plugins if not already present
	if opts.packageSelected("tmux") {
		tpmDir := filepath.Join(home, ".tmux", "plugins", "tpm")
		if _, err := os.Stat(tpmDir); os.IsNotExist(err) {
			if opts.DryRun {
				log("==> Would install TPM (Tmux Plugin Manager)")
			} else {
				log("==> Installing TPM (Tmux Plugin Manager)")
				if err := shell.RunStreaming("git", []string{
					"clone", "https://github.com/tmux-plugins/tpm", tpmDir,
				}, log); err != nil {
					log(fmt.Sprintf("    Warning: failed to clone TPM: %v", err))
				} else {
					log("==> Installing tmux plugins")
					installScript := filepath.Join(tpmDir, "bin", "install_plugins")
					shell.RunStreaming(installScript, nil, log)
				}
			}
		} else {
			log("    TPM already installed")
		}
	} else {
		log("==> Skipping TPM (Tmux not selected)")
	}

	// Manage shell integrations in .zshrc
	zshrcPath := filepath.Join(home, ".zshrc")
	if !hasSelectedZshrcContent(opts) {
		log("==> Skipping ~/.zshrc update (no selected shell integrations)")
	} else if opts.DryRun {
		log("==> Would update ~/.zshrc with selected shell integrations")
	} else {
		if err := updateZshrcBlock(zshrcPath, opts, log); err != nil {
			log(fmt.Sprintf("    Warning: failed to update .zshrc: %v", err))
		}
	}

	if !opts.DryRun {
		log("    Writing checksums to state file")
		if err := SaveState(state); err != nil {
			return fmt.Errorf("save state: %w", err)
		}
	}

	return nil
}

func selectedConfigMappings(opts Options) []manifest.ConfigMapping {
	var configs []manifest.ConfigMapping
	for _, cfg := range manifest.Configs() {
		if configMappingSelected(cfg, opts) {
			configs = append(configs, cfg)
		}
	}
	return configs
}

func configMappingSelected(cfg manifest.ConfigMapping, opts Options) bool {
	switch cfg.Source {
	case "aerospace/aerospace.toml":
		return opts.packageSelected("nikitabobko/tap/aerospace")
	case "ghostty/config":
		return opts.packageSelected("ghostty")
	case "tmux/tmux.conf":
		return opts.packageSelected("tmux")
	case "starship.toml":
		return opts.packageSelected("starship")
	case "zed/keymap.json", "zed/settings.json":
		return opts.packageSelected("zed")
	case "Raycast 2026-06-11 15.59.34.rayconfig":
		return opts.packageSelected("raycast")
	case "omachy/dev-session.sh":
		return opts.allPackagesSelected("tmux", "neovim", "opencode", "lazygit")
	case "yazi/keymap.toml", "yazi/theme.toml", "yazi/yazi.toml":
		return opts.packageSelected("yazi")
	case "zshrc/.aliases":
		return opts.anyPackageSelected("neovim", "eza", "lazygit", "lazydocker", "bat")
	default:
		return true
	}
}

func configProgress(i, total int) int {
	if total == 0 {
		return 80
	}
	return 60 + ((i+1)*20)/total
}

const (
	zshrcMarkerStart = "# ── Omachy managed (do not edit between these markers) ──"
	zshrcMarkerEnd   = "# ── End Omachy managed ──"
	zshrcExtrasPath  = "configs/zshrc/extras"
)

// shellIntegrations are the init lines for tools that need shell configuration.
type shellIntegration struct {
	pkg   string
	check string // string to search for in existing .zshrc
	line  string // line to add
}

var shellIntegrations = []shellIntegration{
	{`mise`, `mise activate zsh`, `eval "$(mise activate zsh)"`},
	{`starship`, `starship init zsh`, `eval "$(starship init zsh)"`},
	{`fzf`, `fzf --zsh`, `eval "$(fzf --zsh)"`},
	{`atuin`, `atuin init zsh`, `eval "$(atuin init zsh)"`},
	{``, `set -o vi`, `set -o vi`},
	{`zsh-syntax-highlighting`, `zsh-syntax-highlighting.zsh`, `source $(brew --prefix)/share/zsh-syntax-highlighting/zsh-syntax-highlighting.zsh`},
	{`zsh-autosuggestions`, `zsh-autosuggestions.zsh`, `source $(brew --prefix)/share/zsh-autosuggestions/zsh-autosuggestions.zsh`},
	{`fastfetch`, `fastfetch`, `fastfetch`},
	{``, `dev()`, `dev() { sh ~/.config/omachy/dev-session.sh "$@"; }`},
}

func updateZshrcBlock(path string, opts Options, log func(string)) error {
	return updateZshrcBlockWithExtrasAndIntegrations(path, selectedZshrcExtras(opts), selectedShellIntegrations(opts), log)
}

func loadZshrcExtras() (string, error) {
	data, err := EmbeddedConfigs.ReadFile(zshrcExtrasPath)
	if err != nil {
		return "", fmt.Errorf("read embedded %s: %w", zshrcExtrasPath, err)
	}
	return string(data), nil
}

func updateZshrcBlockWithExtras(path, extras string, log func(string)) error {
	return updateZshrcBlockWithExtrasAndIntegrations(path, extras, shellIntegrations, log)
}

func updateZshrcBlockWithExtrasAndIntegrations(path, extras string, integrations []shellIntegration, log func(string)) error {
	// Read existing content, or start with empty
	existing := ""
	if data, err := os.ReadFile(path); err == nil {
		existing = string(data)
	}

	// Remove existing managed block so we can rebuild it with all integrations
	cleaned := removeManagedBlock(existing)

	// Check for orphaned markers — if only one marker exists, don't add a new block
	// to avoid duplicates. Warn the user instead.
	if strings.Contains(cleaned, zshrcMarkerStart) || strings.Contains(cleaned, zshrcMarkerEnd) {
		log("    Warning: found orphaned Omachy markers in .zshrc — please fix manually")
		return nil
	}

	// Build the managed block. Skip integrations that already exist outside
	// the managed block (i.e. the user set them up themselves).
	var lines []string
	for _, si := range integrations {
		if strings.Contains(cleaned, si.check) {
			log(fmt.Sprintf("    Already present outside managed block: %s", si.check))
		} else {
			lines = append(lines, si.line)
			log(fmt.Sprintf("    Adding: %s", si.line))
		}
	}
	trimmedExtras := strings.Trim(extras, "\n")
	if trimmedExtras != "" {
		if strings.Contains(cleaned, trimmedExtras) {
			log("    Already present outside managed block: zshrc extras")
		} else {
			lines = append(lines, trimmedExtras)
			log("    Adding: zshrc extras")
		}
	}

	if len(lines) == 0 {
		log("==> Shell integrations already configured in ~/.zshrc")
		return nil
	}

	log("==> Updating ~/.zshrc with shell integrations")

	// Build new managed block with all integrations
	block := zshrcMarkerStart + "\n"
	for _, line := range lines {
		block += line + "\n"
	}
	block += zshrcMarkerEnd + "\n"

	// Append to file, ensuring exactly one blank line separator
	base := strings.TrimRight(cleaned, "\n")
	var newContent string
	if base == "" {
		newContent = block
	} else {
		newContent = base + "\n\n" + block
	}

	return os.WriteFile(path, []byte(newContent), 0644)
}

func selectedShellIntegrations(opts Options) []shellIntegration {
	var selected []shellIntegration
	for _, si := range shellIntegrations {
		if si.check == "dev()" {
			if configMappingSelected(manifest.ConfigMapping{Source: "omachy/dev-session.sh"}, opts) {
				selected = append(selected, si)
			}
			continue
		}
		if si.pkg == "" || opts.packageSelected(si.pkg) {
			selected = append(selected, si)
		}
	}
	return selected
}

func selectedZshrcExtras(opts Options) string {
	var b strings.Builder
	if configMappingSelected(manifest.ConfigMapping{Source: "zshrc/.aliases"}, opts) {
		b.WriteString("\n# Add aliases\nsource ~/.aliases\n")
	}
	if opts.packageSelected("yazi") {
		b.WriteString(`
function y() {
	local tmp="$(mktemp -t "yazi-cwd.XXXXXX")" cwd
	command yazi "$@" --cwd-file="$tmp"
	IFS= read -r -d '' cwd < "$tmp"
	[ "$cwd" != "$PWD" ] && [ -d "$cwd" ] && builtin cd "$cwd"
	rm -f -- "$tmp"
}
`)
	}
	if opts.packageSelected("neovim") {
		b.WriteString("\n# Set default editor to nvim\nexport EDITOR=nvim\n")
	}
	return b.String()
}

func hasSelectedZshrcContent(opts Options) bool {
	return selectedZshrcExtras(opts) != "" || len(selectedShellIntegrations(opts)) > 0
}

func removeManagedBlock(content string) string {
	startIdx := strings.Index(content, zshrcMarkerStart)
	endIdx := strings.Index(content, zshrcMarkerEnd)
	if startIdx == -1 || endIdx == -1 || endIdx < startIdx {
		return content
	}
	before := content[:startIdx]
	after := content[endIdx+len(zshrcMarkerEnd):]
	if len(after) > 0 && after[0] == '\n' {
		after = after[1:]
	}
	return before + after
}

func deployFile(source, dest string, mode fs.FileMode) error {
	data, err := EmbeddedConfigs.ReadFile(filepath.Join("configs", source))
	if err != nil {
		return fmt.Errorf("read embedded %s: %w", source, err)
	}

	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return err
	}

	return os.WriteFile(dest, data, mode)
}

func deployDir(source, dest string, mode uint32) error {
	if err := os.MkdirAll(dest, 0755); err != nil {
		return err
	}

	srcDir := filepath.Join("configs", source)
	return fs.WalkDir(EmbeddedConfigs, srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip .gitkeep files
		if d.Name() == ".gitkeep" {
			return nil
		}

		rel, _ := filepath.Rel(srcDir, path)
		target := filepath.Join(dest, rel)

		if d.IsDir() {
			return os.MkdirAll(target, 0755)
		}

		data, err := EmbeddedConfigs.ReadFile(path)
		if err != nil {
			return err
		}

		fileMode := fs.FileMode(mode)
		// Scripts get 0755
		if strings.HasSuffix(d.Name(), ".sh") || strings.HasPrefix(d.Name(), "plugin.") {
			fileMode = 0755
		}

		return os.WriteFile(target, data, fileMode)
	})
}

func expandHome(path, home string) string {
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(home, path[2:])
	}
	return path
}
