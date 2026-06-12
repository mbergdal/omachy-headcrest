package tui

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func newTestApp() App {
	return NewApp(
		[]string{"Preflight", "Packages", "Configs"},
		func(p *tea.Program, selectedPackages []string) {},
		SplashOptions{},
		"test",
		nil,
	)
}

func newTestAppWithPackages() App {
	return NewApp(
		[]string{"Preflight", "Packages", "Configs"},
		func(p *tea.Program, selectedPackages []string) {},
		SplashOptions{},
		"test",
		[]PackageChoice{
			{Name: "ghostty", Label: "Ghostty", Kind: "Cask", Selected: true},
			{Name: "zed", Label: "Zed", Kind: "Cask", Installed: true},
		},
	)
}

func TestAppSplashToInstall(t *testing.T) {
	app := newTestApp()

	// Set a size so View() doesn't return "Loading..."
	model, _ := app.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	app = model.(App)

	if app.started {
		t.Error("app should start in splash mode")
	}

	// View should show splash (ASCII art logo)
	view := app.View()
	if !strings.Contains(view, "___") {
		t.Error("splash should be visible before start")
	}

	// Press Enter to start
	model, _ = app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	app = model.(App)

	if !app.started {
		t.Error("app should transition to started after Enter")
	}
}

func TestAppSplashToSelectorToInstall(t *testing.T) {
	app := newTestAppWithPackages()

	model, _ := app.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	app = model.(App)

	model, _ = app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	app = model.(App)
	if !app.selecting {
		t.Error("app should show package selector after splash")
	}
	if app.started {
		t.Error("app should not start installer while selector is visible")
	}

	model, _ = app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	app = model.(App)
	if !app.started {
		t.Error("app should start installer after confirming selector")
	}
	if app.selecting {
		t.Error("selector should close after starting installer")
	}
}

func TestAppQuitFromSplash(t *testing.T) {
	app := newTestApp()
	model, _ := app.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	app = model.(App)

	_, cmd := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	// Should produce a quit command
	if cmd == nil {
		t.Error("q should produce a quit command")
	}
}

func TestAppPhaseUpdates(t *testing.T) {
	app := newTestApp()

	// Start the app
	model, _ := app.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	app = model.(App)
	model, _ = app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	app = model.(App)

	// Send phase events
	model, _ = app.Update(PhaseStarted{Name: "Preflight"})
	app = model.(App)
	if app.phases.Phases[0].Status != StatusActive {
		t.Error("Preflight should be active")
	}

	model, _ = app.Update(PhaseCompleted{Name: "Preflight"})
	app = model.(App)
	if app.phases.Phases[0].Status != StatusDone {
		t.Error("Preflight should be done")
	}

	model, _ = app.Update(PhaseFailed{Name: "Packages", Error: errors.New("fail")})
	app = model.(App)
	if app.phases.Phases[1].Status != StatusFailed {
		t.Error("Packages should be failed")
	}
}

func TestAppInstallFinished(t *testing.T) {
	app := newTestApp()

	model, _ := app.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	app = model.(App)
	model, _ = app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	app = model.(App)

	model, _ = app.Update(InstallFinished{})
	app = model.(App)

	if !app.finished {
		t.Error("app should be finished")
	}
	if !app.help.Finished {
		t.Error("help should show finished state")
	}
}

func TestAppProgressUpdate(t *testing.T) {
	app := newTestApp()

	model, _ := app.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	app = model.(App)
	model, _ = app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	app = model.(App)

	model, _ = app.Update(ProgressUpdate{Percent: 75})
	app = model.(App)

	if app.header.Percent != 75 {
		t.Errorf("header percent should be 75, got %d", app.header.Percent)
	}
}
