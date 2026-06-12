package installer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUpdateZshrcBlockWithExtras(t *testing.T) {
	tmp := t.TempDir()
	zshrc := filepath.Join(tmp, ".zshrc")
	if err := os.WriteFile(zshrc, []byte("# user config\n"), 0644); err != nil {
		t.Fatal(err)
	}

	extras := `
# Add aliases
source ~/.aliases

function y() {
	command yazi "$@"
}

export EDITOR=nvim
`
	if err := updateZshrcBlockWithExtras(zshrc, extras, func(string) {}); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(zshrc)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	for _, want := range []string{
		"# user config",
		zshrcMarkerStart,
		`eval "$(mise activate zsh)"`,
		"# Add aliases",
		"source ~/.aliases",
		"function y()",
		"export EDITOR=nvim",
		zshrcMarkerEnd,
	} {
		if !strings.Contains(content, want) {
			t.Errorf(".zshrc missing %q in:\n%s", want, content)
		}
	}
}

func TestUpdateZshrcBlockWithExtrasIsIdempotent(t *testing.T) {
	tmp := t.TempDir()
	zshrc := filepath.Join(tmp, ".zshrc")
	extras := "export EDITOR=nvim\n"

	if err := updateZshrcBlockWithExtras(zshrc, extras, func(string) {}); err != nil {
		t.Fatal(err)
	}
	if err := updateZshrcBlockWithExtras(zshrc, extras, func(string) {}); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(zshrc)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	if got := strings.Count(content, zshrcMarkerStart); got != 1 {
		t.Fatalf("managed block count = %d, want 1 in:\n%s", got, content)
	}
	if got := strings.Count(content, "export EDITOR=nvim"); got != 1 {
		t.Fatalf("EDITOR line count = %d, want 1 in:\n%s", got, content)
	}
}

func TestSelectedConfigMappings(t *testing.T) {
	configs := selectedConfigMappings(Options{
		PackageSelectionEnabled: true,
		SelectedPackages:        []string{"ghostty", "zed"},
	})

	seen := map[string]bool{}
	for _, cfg := range configs {
		seen[cfg.Source] = true
	}

	for _, want := range []string{"ghostty/config", "zed/keymap.json", "zed/settings.json"} {
		if !seen[want] {
			t.Fatalf("missing selected config %q in %v", want, seen)
		}
	}
	for _, unwanted := range []string{"aerospace/aerospace.toml", "tmux/tmux.conf", "Raycast 2026-06-11 15.59.34.rayconfig"} {
		if seen[unwanted] {
			t.Fatalf("unselected config %q should be skipped", unwanted)
		}
	}
}

func TestSelectedShellIntegrations(t *testing.T) {
	integrations := selectedShellIntegrations(Options{
		PackageSelectionEnabled: true,
		SelectedPackages:        []string{"starship"},
	})

	seen := map[string]bool{}
	for _, integration := range integrations {
		seen[integration.check] = true
	}
	if !seen["starship init zsh"] {
		t.Fatalf("starship integration missing: %v", seen)
	}
	if seen["mise activate zsh"] || seen["fzf --zsh"] {
		t.Fatalf("unselected integrations should be skipped: %v", seen)
	}
}
