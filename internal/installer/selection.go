package installer

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/dough654/Omachy/internal/brew"
	"github.com/dough654/Omachy/internal/manifest"
	"github.com/dough654/Omachy/internal/shell"
	"github.com/dough654/Omachy/internal/tui"
)

var packageMetadata = map[string]struct {
	Label       string
	Description string
}{
	"nikitabobko/tap/aerospace": {"AeroSpace", "tiling window manager"},
	"ghostty":                   {"Ghostty", "terminal emulator"},
	"neovim":                    {"Neovim + LazyVim", "text editor"},
	"tree-sitter":               {"tree-sitter", "syntax parser tooling"},
	"tmux":                      {"Tmux + TPM", "terminal multiplexer"},
	"font-hack-nerd-font":       {"Hack Nerd Font", "terminal/icon font"},
	"font-jetbrains-mono":       {"JetBrains Mono", "monospace font"},
	"zed":                       {"Zed", "graphical code editor"},
	"zen":                       {"Zen Browser", "web browser"},
	"raycast":                   {"Raycast", "launcher and automation"},
	"1password":                 {"1Password", "password manager"},
	"superhuman":                {"Superhuman", "email client"},
	"claude":                    {"Claude", "AI assistant app"},
	"chatgpt":                   {"ChatGPT", "AI assistant app"},
	"starship":                  {"Starship", "cross-shell prompt"},
	"fzf":                       {"fzf", "fuzzy finder"},
	"docker":                    {"Docker", "container CLI"},
	"docker-compose":            {"Docker Compose", "compose plugin/CLI"},
	"colima":                    {"Colima", "Docker runtime service"},
	"lazygit":                   {"Lazygit", "git TUI"},
	"opencode":                  {"opencode", "AI coding agent CLI"},
	"gh":                        {"GitHub CLI", "gh command-line client"},
	"eza":                       {"eza", "modern ls replacement"},
	"bat":                       {"bat", "syntax-highlighted cat"},
	"yazi":                      {"Yazi", "terminal file manager"},
	"lazydocker":                {"Lazydocker", "docker TUI"},
	"atuin":                     {"Atuin", "shell history search"},
	"zsh-syntax-highlighting":   {"zsh-syntax-highlighting", "Zsh syntax highlighting"},
	"zsh-autosuggestions":       {"zsh-autosuggestions", "Zsh history suggestions"},
	"fastfetch":                 {"fastfetch", "system info display"},
	"mise":                      {"mise", "runtime manager"},
}

var packageAppNames = map[string][]string{
	"nikitabobko/tap/aerospace": {"AeroSpace.app"},
	"ghostty":                   {"Ghostty.app"},
	"zed":                       {"Zed.app"},
	"zen":                       {"Zen.app"},
	"raycast":                   {"Raycast.app"},
	"1password":                 {"1Password.app"},
	"superhuman":                {"Superhuman.app"},
	"claude":                    {"Claude.app"},
	"chatgpt":                   {"ChatGPT.app"},
}

var packageBinaries = map[string][]string{
	"neovim":                  {"nvim"},
	"tree-sitter":             {"tree-sitter"},
	"tmux":                    {"tmux"},
	"starship":                {"starship"},
	"fzf":                     {"fzf"},
	"docker":                  {"docker"},
	"docker-compose":          {"docker-compose"},
	"colima":                  {"colima"},
	"lazygit":                 {"lazygit"},
	"opencode":                {"opencode"},
	"gh":                      {"gh"},
	"eza":                     {"eza"},
	"bat":                     {"bat"},
	"yazi":                    {"yazi"},
	"lazydocker":              {"lazydocker"},
	"atuin":                   {"atuin"},
	"zsh-syntax-highlighting": {"zsh-syntax-highlighting.zsh"},
	"zsh-autosuggestions":     {"zsh-autosuggestions.zsh"},
	"fastfetch":               {"fastfetch"},
	"mise":                    {"mise"},
}

// PackageChoices returns manifest packages with their current install status.
func PackageChoices() []tui.PackageChoice {
	var formulae, casks map[string]bool
	var formulaErr, caskErr error
	brewAvailable := false
	if _, found := shell.Which("brew"); found {
		brewAvailable = true
		formulae, formulaErr = brew.InstalledSet(false)
		casks, caskErr = brew.InstalledSet(true)
	}

	var choices []tui.PackageChoice
	for _, pkg := range manifest.Packages() {
		installed := installedFromSet(pkg.Name, pkg.Cask, formulae, casks)
		if ((pkg.Cask && caskErr != nil) || (!pkg.Cask && formulaErr != nil)) && brewAvailable {
			installed = brew.IsInstalled(pkg.Name, pkg.Cask)
		}
		if !installed {
			installed = packagePresentOnMachine(pkg.Name)
		}

		meta := packageMetadata[pkg.Name]
		label := meta.Label
		if label == "" {
			label = packageToken(pkg.Name)
		}
		kind := "Formula"
		if pkg.Cask {
			kind = "Cask"
		}
		choices = append(choices, tui.PackageChoice{
			Name:        pkg.Name,
			Label:       label,
			Description: meta.Description,
			Kind:        kind,
			Installed:   installed,
			Selected:    !installed,
		})
	}
	return choices
}

func packagePresentOnMachine(name string) bool {
	for _, bin := range packageBinaries[name] {
		if _, found := shell.Which(bin); found {
			return true
		}
	}
	for _, app := range packageAppNames[name] {
		for _, dir := range applicationDirs() {
			if _, err := os.Stat(filepath.Join(dir, app)); err == nil {
				return true
			}
		}
	}
	return false
}

func applicationDirs() []string {
	dirs := []string{"/Applications"}
	if home, err := os.UserHomeDir(); err == nil {
		dirs = append(dirs, filepath.Join(home, "Applications"))
	}
	return dirs
}

func installedFromSet(name string, cask bool, formulae, casks map[string]bool) bool {
	set := formulae
	if cask {
		set = casks
	}
	if set == nil {
		return false
	}
	return set[name] || set[packageToken(name)]
}

func packageToken(name string) string {
	if idx := strings.LastIndex(name, "/"); idx >= 0 {
		return name[idx+1:]
	}
	return name
}
