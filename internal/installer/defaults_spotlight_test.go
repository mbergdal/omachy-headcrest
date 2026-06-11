package installer

import "testing"

func TestSymbolicHotkeyStateKey(t *testing.T) {
	got := symbolicHotkeyStateKey("64")
	want := "com.apple.symbolichotkeys:AppleSymbolicHotKeys.64.enabled"
	if got != want {
		t.Fatalf("symbolicHotkeyStateKey = %q, want %q", got, want)
	}
}

func TestSymbolicHotkeyByID(t *testing.T) {
	hk, ok := symbolicHotkeyByID("64")
	if !ok {
		t.Fatal("expected Spotlight hotkey 64")
	}
	if hk.Label != "Spotlight search (Cmd-Space)" {
		t.Fatalf("Label = %q", hk.Label)
	}
}

func TestRestoreSpecialDefaultIgnoresRegularDefaults(t *testing.T) {
	handled, err := RestoreSpecialDefault("com.apple.dock", "autohide", "-bool:false", true, func(string) {})
	if err != nil {
		t.Fatal(err)
	}
	if handled {
		t.Fatal("regular default should not be handled as special")
	}
}
