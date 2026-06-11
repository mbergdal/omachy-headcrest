package manifest

// Package describes a Homebrew package to install.
type Package struct {
	Name         string
	Tap          string // empty if no tap needed
	Cask         bool
	Service      bool   // true if this package runs as a brew service
	SkipIfBinary string // if set, skip install when this binary is already on PATH
}

// ConfigMapping describes where an embedded config should be deployed.
type ConfigMapping struct {
	Source         string // path relative to embedded configs/ dir
	Dest           string // absolute destination path (~ expanded at runtime)
	IsDir          bool   // true if this is a directory of files
	Mode           uint32 // file permission mode (0644 for configs, 0755 for scripts)
	NeverOverwrite bool   // if true, skip deployment when destination already exists
}

// Taps returns the unique set of taps needed.
func Taps() []string {
	seen := map[string]bool{}
	var taps []string
	for _, pkg := range Packages() {
		if pkg.Tap != "" && !seen[pkg.Tap] {
			seen[pkg.Tap] = true
			taps = append(taps, pkg.Tap)
		}
	}
	return taps
}

// Packages returns the full list of packages to install.
func Packages() []Package {
	return []Package{
		{Name: "nikitabobko/tap/aerospace", Tap: "nikitabobko/tap", Cask: true},
		{Name: "ghostty", Cask: true},
		{Name: "neovim"},
		{Name: "tree-sitter"},
		{Name: "tmux"},
		{Name: "font-hack-nerd-font", Cask: true},
		{Name: "font-jetbrains-mono", Cask: true},
		{Name: "zed", Cask: true},
		{Name: "zen", Cask: true},
		{Name: "raycast", Cask: true},
		{Name: "1password", Cask: true},
		{Name: "superhuman", Cask: true},
		{Name: "starship"},
		{Name: "fzf"},
		{Name: "docker"},
		{Name: "docker-compose"},
		{Name: "colima", Service: true},
		{Name: "lazygit"},
		{Name: "opencode"},
		{Name: "gh"},
		{Name: "eza"},
		{Name: "bat"},
		{Name: "yazi"},
		{Name: "lazydocker"},
		{Name: "atuin"},
		{Name: "zsh-syntax-highlighting"},
		{Name: "zsh-autosuggestions"},
		{Name: "fastfetch"},
		Package{Name: "mise"},
	}
}

// Services returns packages that should be started as brew services.
func Services() []Package {
	var svcs []Package
	for _, pkg := range Packages() {
		if pkg.Service {
			svcs = append(svcs, pkg)
		}
	}
	return svcs
}

// Configs returns the config file mappings.
func Configs() []ConfigMapping {
	return []ConfigMapping{
		{Source: "aerospace/aerospace.toml", Dest: "~/.config/aerospace/aerospace.toml", Mode: 0644},
		{Source: "ghostty/config", Dest: "~/Library/Application Support/com.mitchellh.ghostty/config", Mode: 0644},
		{Source: "tmux/tmux.conf", Dest: "~/.tmux.conf", Mode: 0644, NeverOverwrite: true},
		{Source: "starship.toml", Dest: "~/.config/starship.toml", Mode: 0644},
		{Source: "zed/keymap.json", Dest: "~/.config/zed/keymap.json", Mode: 0644},
		{Source: "zed/settings.json", Dest: "~/.config/zed/settings.json", Mode: 0644},
		{Source: "Raycast 2026-06-11 15.59.34.rayconfig", Dest: "~/.config/omachy/raycast.rayconfig", Mode: 0644},
		{Source: "omachy/dev-session.sh", Dest: "~/.config/omachy/dev-session.sh", Mode: 0755},
		{Source: "yazi/keymap.toml", Dest: "~/.config/yazi/keymap.toml", Mode: 0755},
		{Source: "yazi/theme.toml", Dest: "~/.config/yazi/theme.toml", Mode: 0755},
		{Source: "yazi/yazi.toml", Dest: "~/.config/yazi/yazi.toml", Mode: 0755},
		{Source: "zshrc/.aliases", Dest: "~/.aliases", Mode: 0755},
	}
}
