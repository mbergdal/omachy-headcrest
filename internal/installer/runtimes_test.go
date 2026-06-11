package installer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallMiseRuntimes(t *testing.T) {
	tmp := t.TempDir()
	statePathOverride = filepath.Join(tmp, "state.json")
	defer func() { statePathOverride = "" }()

	logPath := filepath.Join(tmp, "mise.log")
	writeFakeMise(t, tmp, logPath)

	if err := installMiseRuntimes(false, func(string) {}); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	commands := string(data)

	for _, want := range []string{
		"use --global node@latest",
		"use --global python@latest",
		"use --global bun@latest",
		"use --global pnpm@latest",
		"use --global erlang@latest",
		"use --global elixir@latest",
		"exec -- npm install -g npm@latest",
	} {
		if !strings.Contains(commands, want) {
			t.Errorf("missing mise command %q in:\n%s", want, commands)
		}
	}

	state, err := LoadState()
	if err != nil {
		t.Fatal(err)
	}
	if len(state.InstalledRuntimes) != len(MiseRuntimes) {
		t.Fatalf("InstalledRuntimes: got %d, want %d", len(state.InstalledRuntimes), len(MiseRuntimes))
	}
	if state.InstalledRuntimes[0].Name != "node" || state.InstalledRuntimes[0].Spec != "node@latest" {
		t.Errorf("first runtime = %+v, want node@latest", state.InstalledRuntimes[0])
	}
}

func TestInstallMiseRuntimesDryRunDoesNotRequireMise(t *testing.T) {
	tmp := t.TempDir()
	statePathOverride = filepath.Join(tmp, "state.json")
	defer func() { statePathOverride = "" }()

	var lines []string
	if err := installMiseRuntimes(true, func(line string) { lines = append(lines, line) }); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(statePathOverride); !os.IsNotExist(err) {
		t.Fatalf("dry run should not write state, stat err: %v", err)
	}
	if !strings.Contains(strings.Join(lines, "\n"), "Would install node") {
		t.Fatalf("dry run logs did not include node install: %v", lines)
	}
}

func writeFakeMise(t *testing.T, tmp, logPath string) {
	t.Helper()

	script := filepath.Join(tmp, "mise")
	body := "#!/bin/sh\nprintf '%s\\n' \"$*\" >> \"$MISE_LOG\"\n"
	if err := os.WriteFile(script, []byte(body), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MISE_LOG", logPath)
	t.Setenv("PATH", tmp+":"+os.Getenv("PATH"))
}
