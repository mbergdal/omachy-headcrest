package installer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dough654/Omachy/internal/brew"
	"github.com/dough654/Omachy/internal/manifest"
	"github.com/dough654/Omachy/internal/shell"
	"github.com/dough654/Omachy/internal/tui"
)

// MacOSDefault represents a defaults write operation.
type MacOSDefault struct {
	Domain string
	Key    string
	Type   string // -bool, -int, -float, -string
	Value  string
	Label  string // human-readable description
}

// MacOSDefaults is the list of system defaults that Omachy sets.
var MacOSDefaults = []MacOSDefault{
	{"com.apple.dock", "autohide", "-bool", "true", "Auto-hide Dock"},
	{"com.apple.dock", "autohide-delay", "-float", "0", "Remove Dock auto-hide delay"},
	{"com.apple.dock", "autohide-time-modifier", "-float", "0.25", "Fast Dock hide animation"},
	{"com.apple.dock", "mru-spaces", "-bool", "false", "Disable MRU Spaces reordering"},
	{"com.apple.dock", "tilesize", "-int", "48", "Set Dock icon size"},
	{"com.apple.dock", "mineffect", "-string", "scale", "Use scale minimize effect"},
	{"com.apple.dock", "show-recents", "-bool", "false", "Hide recent apps in Dock"},
	{"NSGlobalDomain", "NSAutomaticWindowAnimationsEnabled", "-bool", "false", "Disable window open/close animations"},
	{"NSGlobalDomain", "AppleShowAllExtensions", "-bool", "true", "Show all file extensions"},
	{"NSGlobalDomain", "KeyRepeat", "-int", "1", "Fastest key repeat rate"},
	{"NSGlobalDomain", "InitialKeyRepeat", "-int", "10", "Shortest key repeat delay"},
	{"-g", "ApplePressAndHoldEnabled", "-bool", "false", "Disable press-and-hold for key repeat"},
	// {"NSGlobalDomain", "_HIHideMenuBar", "-bool", "true", "Auto-hide menu bar"},
	{"com.apple.WindowManager", "StandardHideWidgets", "-bool", "true", "Hide desktop widgets"},
	{"com.apple.WindowManager", "EnableStandardClickToShowDesktop", "-bool", "false", "Disable click wallpaper to show desktop"},
}

const (
	symbolicHotkeysDomain      = "com.apple.symbolichotkeys"
	symbolicHotkeysKey         = "AppleSymbolicHotKeys"
	symbolicHotkeyAbsentMarker = "__omachy_absent__"
)

type symbolicHotkeyDefault struct {
	ID         string
	Label      string
	Parameters string
}

var spotlightHotkeys = []symbolicHotkeyDefault{
	{ID: "64", Label: "Spotlight search (Cmd-Space)", Parameters: "32, 49, 1048576"},
	{ID: "65", Label: "Finder search window (Cmd-Option-Space)", Parameters: "32, 49, 1572864"},
}

func runSystem(p *tea.Program, opts Options) error {
	log := func(text string) {
		p.Send(tui.LogLine{Text: text})
	}

	state, err := LoadState()
	if err != nil {
		return fmt.Errorf("load state: %w", err)
	}

	// Apply macOS defaults
	log("==> Applying macOS defaults")
	for _, d := range MacOSDefaults {
		if opts.DryRun {
			log(fmt.Sprintf("    Would set: %s", d.Label))
			continue
		}

		// Read current value and type first for undo
		stateKey := fmt.Sprintf("%s:%s", d.Domain, d.Key)
		if _, exists := state.OriginalDefaults[stateKey]; !exists {
			current, err := readDefault(d.Domain, d.Key)
			if err == nil {
				typ, typErr := readDefaultType(d.Domain, d.Key)
				if typErr == nil {
					state.OriginalDefaults[stateKey] = typ + ":" + current
				} else {
					state.OriginalDefaults[stateKey] = d.Type + ":" + current
				}
			}
		}

		// Write new value
		if err := writeDefault(d.Domain, d.Key, d.Type, d.Value); err != nil {
			return fmt.Errorf("defaults write %s %s: %w", d.Domain, d.Key, err)
		}
		log(fmt.Sprintf("    %s", d.Label))
	}

	if err := disableSpotlightShortcuts(opts.DryRun, state, log); err != nil {
		return err
	}

	// Restart Dock and SystemUIServer to apply changes
	if !opts.DryRun {
		log("==> Restarting Dock")
		shell.Run("killall", "Dock")
		log("==> Restarting SystemUIServer (menu bar)")
		shell.Run("killall", "SystemUIServer")
	}

	p.Send(tui.ProgressUpdate{Percent: 90})

	if err := startBrewServices(opts.DryRun, state, log); err != nil {
		return err
	}

	// Record which managed processes are already running before we start them,
	// so uninstall only kills processes that Omachy started.
	if !opts.DryRun {
		for _, proc := range []string{"AeroSpace"} {
			if isProcessRunning(proc) {
				state.RunningProcesses = appendUnique(state.RunningProcesses, proc)
			}
		}
	}

	if err := ensureAeroSpace(opts.DryRun, p, log); err != nil {
		return err
	}

	if err := openRaycastImport(opts.DryRun, homeConfigPath("~/.config/omachy/raycast.rayconfig"), log); err != nil {
		return err
	}

	// Save state
	if !opts.DryRun {
		if err := SaveState(state); err != nil {
			return fmt.Errorf("save state: %w", err)
		}
	}

	return nil
}

func ensureAeroSpace(dryRun bool, p *tea.Program, log func(string)) error {
	if dryRun {
		log("==> Would start or reload AeroSpace and check Accessibility permissions")
		return nil
	}

	if isAerospaceRunning() {
		log("==> Reloading AeroSpace config")
		if err := shell.RunStreaming("aerospace", []string{"reload-config", "--no-gui"}, log); err != nil {
			log(fmt.Sprintf("    Warning: failed to reload AeroSpace config: %v", err))
		}
		return nil
	}

	log("==> Starting AeroSpace")
	shell.Run("open", "-a", "AeroSpace")
	time.Sleep(3 * time.Second)

	if isAerospaceRunning() {
		log("==> AeroSpace is running (Accessibility permissions granted)")
		return nil
	}

	log("==> AeroSpace needs Accessibility permissions")
	log("    1. A dialog should have appeared — click 'Open System Settings'")
	log("    2. Enable the toggle for AeroSpace in Privacy → Accessibility")

	done := make(chan struct{})
	p.Send(tui.WaitForUser{
		Prompt: "    When you've granted permissions, confirm below.",
		Done:   done,
	})
	<-done

	log("==> Relaunching AeroSpace...")
	shell.Run("open", "-a", "AeroSpace")
	time.Sleep(3 * time.Second)

	if isAerospaceRunning() {
		log("==> AeroSpace is running!")
	} else {
		log("==> AeroSpace still not running — you may need to open it manually")
	}
	return nil
}

func openRaycastImport(dryRun bool, configPath string, log func(string)) error {
	url := "raycast://extensions/raycast/raycast/import-settings-data"
	if dryRun {
		log("==> Would open Raycast settings import")
		log(fmt.Sprintf("    Select config file: %s", configPath))
		return nil
	}

	log("==> Opening Raycast settings import")
	log(fmt.Sprintf("    Select config file: %s", configPath))
	if _, err := shell.Run("open", url); err != nil {
		return fmt.Errorf("open Raycast import: %w", err)
	}
	return nil
}

func homeConfigPath(path string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(home, path[2:])
	}
	return path
}

func startBrewServices(dryRun bool, state *State, log func(string)) error {
	for _, svc := range manifest.Services() {
		if dryRun {
			log(fmt.Sprintf("==> Would start service: %s", svc.Name))
			continue
		}

		if err := brew.StartService(svc.Name, log); err != nil {
			return fmt.Errorf("start service %s: %w", svc.Name, err)
		}
		state.Services = appendUnique(state.Services, svc.Name)
	}
	return nil
}

func disableSpotlightShortcuts(dryRun bool, state *State, log func(string)) error {
	log("==> Disabling Spotlight keyboard shortcuts")
	for _, hk := range spotlightHotkeys {
		if dryRun {
			log(fmt.Sprintf("    Would disable: %s", hk.Label))
			continue
		}

		stateKey := symbolicHotkeyStateKey(hk.ID)
		if _, exists := state.OriginalDefaults[stateKey]; !exists {
			enabled, found, err := readSymbolicHotkeyEnabled(hk.ID)
			if err != nil {
				return fmt.Errorf("read %s shortcut state: %w", hk.Label, err)
			}
			if found {
				state.OriginalDefaults[stateKey] = fmt.Sprintf("-bool:%t", enabled)
			} else {
				state.OriginalDefaults[stateKey] = symbolicHotkeyAbsentMarker
			}
		}

		if err := writeSymbolicHotkeyEnabled(hk, false); err != nil {
			return fmt.Errorf("disable %s: %w", hk.Label, err)
		}
		log(fmt.Sprintf("    Disabled %s", hk.Label))
	}
	return nil
}

func symbolicHotkeyStateKey(id string) string {
	return fmt.Sprintf("%s:%s.%s.enabled", symbolicHotkeysDomain, symbolicHotkeysKey, id)
}

func readSymbolicHotkeyEnabled(id string) (enabled bool, found bool, err error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return false, false, err
	}
	path := filepath.Join(home, "Library", "Preferences", symbolicHotkeysDomain+".plist")
	result, err := shell.Run("/usr/libexec/PlistBuddy", "-c", fmt.Sprintf("Print :%s:%s:enabled", symbolicHotkeysKey, id), path)
	if err != nil {
		if strings.Contains(result.Stderr, "Does Not Exist") || strings.Contains(result.Stderr, "File Doesn't Exist") {
			return false, false, nil
		}
		return false, false, err
	}

	switch strings.TrimSpace(result.Stdout) {
	case "true", "1":
		return true, true, nil
	case "false", "0":
		return false, true, nil
	default:
		return false, true, fmt.Errorf("unexpected enabled value %q", strings.TrimSpace(result.Stdout))
	}
}

func writeSymbolicHotkeyEnabled(hk symbolicHotkeyDefault, enabled bool) error {
	enabledValue := 0
	if enabled {
		enabledValue = 1
	}
	definition := fmt.Sprintf("{ enabled = %d; value = { parameters = (%s); type = standard; }; }", enabledValue, hk.Parameters)
	_, err := shell.Run("defaults", "write", symbolicHotkeysDomain, symbolicHotkeysKey, "-dict-add", hk.ID, definition)
	return err
}

func deleteSymbolicHotkey(id string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	path := filepath.Join(home, "Library", "Preferences", symbolicHotkeysDomain+".plist")
	result, err := shell.Run("/usr/libexec/PlistBuddy", "-c", fmt.Sprintf("Delete :%s:%s", symbolicHotkeysKey, id), path)
	if err != nil && !strings.Contains(result.Stderr, "Does Not Exist") {
		return err
	}
	return nil
}

// RestoreSpecialDefault restores defaults that need custom write logic.
func RestoreSpecialDefault(domain, key, stored string, dryRun bool, log func(string)) (bool, error) {
	if domain != symbolicHotkeysDomain || !strings.HasPrefix(key, symbolicHotkeysKey+".") || !strings.HasSuffix(key, ".enabled") {
		return false, nil
	}

	id := strings.TrimSuffix(strings.TrimPrefix(key, symbolicHotkeysKey+"."), ".enabled")
	if stored == symbolicHotkeyAbsentMarker {
		if dryRun {
			log(fmt.Sprintf("    Would delete %s %s override", symbolicHotkeysKey, id))
			return true, nil
		}
		if err := deleteSymbolicHotkey(id); err != nil {
			return true, err
		}
		log(fmt.Sprintf("    Deleted %s %s override", symbolicHotkeysKey, id))
		return true, nil
	}

	if !strings.HasPrefix(stored, "-bool:") {
		return true, fmt.Errorf("unexpected stored value %q", stored)
	}
	enabled := strings.TrimPrefix(stored, "-bool:") == "true"
	hk, ok := symbolicHotkeyByID(id)
	if !ok {
		return true, fmt.Errorf("unknown symbolic hotkey id %s", id)
	}
	if dryRun {
		log(fmt.Sprintf("    Would restore %s to enabled=%t", hk.Label, enabled))
		return true, nil
	}
	if err := writeSymbolicHotkeyEnabled(hk, enabled); err != nil {
		return true, err
	}
	log(fmt.Sprintf("    Restored %s", hk.Label))
	return true, nil
}

func symbolicHotkeyByID(id string) (symbolicHotkeyDefault, bool) {
	for _, hk := range spotlightHotkeys {
		if hk.ID == id {
			return hk, true
		}
	}
	return symbolicHotkeyDefault{}, false
}

func readDefault(domain, key string) (string, error) {
	result, err := shell.Run("defaults", "read", domain, key)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(result.Stdout), nil
}

func readDefaultType(domain, key string) (string, error) {
	result, err := shell.Run("defaults", "read-type", domain, key)
	if err != nil {
		return "", err
	}
	// Output is like "Type is boolean", "Type is integer", "Type is float", "Type is string"
	out := strings.TrimSpace(result.Stdout)
	switch {
	case strings.Contains(out, "boolean"):
		return "-bool", nil
	case strings.Contains(out, "integer"):
		return "-int", nil
	case strings.Contains(out, "float"):
		return "-float", nil
	default:
		return "-string", nil
	}
}

func writeDefault(domain, key, typ, value string) error {
	_, err := shell.Run("defaults", "write", domain, key, typ, value)
	return err
}

func isProcessRunning(name string) bool {
	result, err := shell.Run("pgrep", "-x", name)
	return err == nil && strings.TrimSpace(result.Stdout) != ""
}

func isAerospaceRunning() bool {
	return isProcessRunning("AeroSpace")
}

func appendUnique(slice []string, item string) []string {
	for _, s := range slice {
		if s == item {
			return slice
		}
	}
	return append(slice, item)
}
