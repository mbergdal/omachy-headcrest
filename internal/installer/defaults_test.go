package installer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStartBrewServicesStartsColima(t *testing.T) {
	tmp := t.TempDir()
	logPath := filepath.Join(tmp, "brew.log")
	writeFakeServiceBrew(t, tmp, logPath)

	state := &State{}
	if err := startBrewServices(false, state, func(string) {}); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	commands := string(data)
	if !strings.Contains(commands, "services start colima") {
		t.Fatalf("missing colima service start in:\n%s", commands)
	}
	if len(state.Services) != 1 || state.Services[0] != "colima" {
		t.Fatalf("state.Services = %v, want [colima]", state.Services)
	}
}

func TestStartBrewServicesDryRun(t *testing.T) {
	state := &State{}
	var lines []string
	if err := startBrewServices(true, state, func(line string) { lines = append(lines, line) }); err != nil {
		t.Fatal(err)
	}
	if len(state.Services) != 0 {
		t.Fatalf("dry run should not track services, got %v", state.Services)
	}
	if !strings.Contains(strings.Join(lines, "\n"), "Would start service: colima") {
		t.Fatalf("dry run logs missing colima service start: %v", lines)
	}
}

func TestOpenRaycastImportDryRun(t *testing.T) {
	var lines []string
	if err := openRaycastImport(true, "/tmp/raycast.rayconfig", func(line string) { lines = append(lines, line) }); err != nil {
		t.Fatal(err)
	}
	output := strings.Join(lines, "\n")
	if !strings.Contains(output, "Would open Raycast settings import") {
		t.Fatalf("missing dry-run import log: %v", lines)
	}
	if !strings.Contains(output, "/tmp/raycast.rayconfig") {
		t.Fatalf("missing config path log: %v", lines)
	}
}

func TestEnsureAeroSpaceDryRun(t *testing.T) {
	var lines []string
	if err := ensureAeroSpace(true, nil, func(line string) { lines = append(lines, line) }); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.Join(lines, "\n"), "Would start or reload AeroSpace") {
		t.Fatalf("dry run logs missing AeroSpace action: %v", lines)
	}
}

func writeFakeServiceBrew(t *testing.T, tmp, logPath string) {
	t.Helper()

	script := filepath.Join(tmp, "brew")
	body := "#!/bin/sh\nprintf '%s\\n' \"$*\" >> \"$BREW_LOG\"\n"
	if err := os.WriteFile(script, []byte(body), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("BREW_LOG", logPath)
	t.Setenv("PATH", tmp+":"+os.Getenv("PATH"))
}
