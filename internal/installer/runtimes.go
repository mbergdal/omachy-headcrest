package installer

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dough654/Omachy/internal/shell"
	"github.com/dough654/Omachy/internal/tui"
)

// MiseRuntime describes a runtime that should be managed by mise.
type MiseRuntime struct {
	Name string
	Spec string
}

// MiseRuntimes are installed globally for the user's development environment.
var MiseRuntimes = []MiseRuntime{
	{Name: "node", Spec: "node@latest"},
	{Name: "python", Spec: "python@latest"},
	{Name: "bun", Spec: "bun@latest"},
	{Name: "pnpm", Spec: "pnpm@latest"},
	{Name: "erlang", Spec: "erlang@latest"},
	{Name: "elixir", Spec: "elixir@latest"},
}

func runRuntimes(p *tea.Program, opts Options) error {
	log := func(text string) {
		p.Send(tui.LogLine{Text: text})
	}
	if !opts.packageSelected("mise") {
		log("==> Skipping mise runtimes (mise not selected)")
		return nil
	}

	return installMiseRuntimes(opts.DryRun, log)
}

func installMiseRuntimes(dryRun bool, log func(string)) error {
	log("==> Installing mise runtimes")

	if dryRun {
		for _, rt := range MiseRuntimes {
			log(fmt.Sprintf("    Would install %s with mise (%s)", rt.Name, rt.Spec))
		}
		log("    Would update npm to latest using mise-managed Node")
		return nil
	}

	if _, found := shell.Which("mise"); !found {
		return fmt.Errorf("mise not found; package phase should install it first")
	}

	state, err := LoadState()
	if err != nil {
		return fmt.Errorf("load state: %w", err)
	}

	for _, rt := range MiseRuntimes {
		log(fmt.Sprintf("==> Installing %s with mise (%s)", rt.Name, rt.Spec))
		if err := shell.RunStreaming("mise", []string{"use", "--global", rt.Spec}, log); err != nil {
			return fmt.Errorf("mise use --global %s: %w", rt.Spec, err)
		}

		state.InstalledRuntimes = appendUniqueRuntime(state.InstalledRuntimes, InstalledRuntime{Name: rt.Name, Spec: rt.Spec})
		if err := SaveState(state); err != nil {
			return fmt.Errorf("save state: %w", err)
		}

		if rt.Name == "node" {
			log("==> Updating npm to latest using mise-managed Node")
			if err := shell.RunStreaming("mise", []string{"exec", "--", "npm", "install", "-g", "npm@latest"}, log); err != nil {
				return fmt.Errorf("update npm: %w", err)
			}
		}
	}

	return nil
}

func appendUniqueRuntime(slice []InstalledRuntime, item InstalledRuntime) []InstalledRuntime {
	for _, existing := range slice {
		if existing.Name == item.Name && existing.Spec == item.Spec {
			return slice
		}
	}
	return append(slice, item)
}
