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
