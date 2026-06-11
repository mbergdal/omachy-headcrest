package manifest

import "testing"

func TestPackages(t *testing.T) {
	pkgs := Packages()
	if len(pkgs) != 27 {
		t.Fatalf("expected 27 packages, got %d", len(pkgs))
	}
	seen := map[string]bool{}
	for i, pkg := range pkgs {
		if pkg.Name == "" {
			t.Errorf("package %d has empty Name", i)
		}
		seen[pkg.Name] = true
	}
	if !seen["mise"] {
		t.Error("expected mise package")
	}
	for _, added := range []string{"zed", "zen", "raycast", "1password", "docker", "docker-compose", "colima", "eza", "yazi"} {
		if !seen[added] {
			t.Errorf("expected %s package", added)
		}
	}
	for _, removed := range []string{"sketchybar", "borders"} {
		if seen[removed] {
			t.Errorf("%s should not be installed", removed)
		}
	}
}

func TestTaps(t *testing.T) {
	taps := Taps()
	if len(taps) == 0 {
		t.Fatal("expected at least one tap")
	}

	// Should be unique
	seen := map[string]bool{}
	for _, tap := range taps {
		if tap == "" {
			t.Error("tap is empty string")
		}
		if seen[tap] {
			t.Errorf("duplicate tap: %s", tap)
		}
		seen[tap] = true
	}

	// Every tap should be referenced by at least one package
	pkgs := Packages()
	for _, tap := range taps {
		found := false
		for _, pkg := range pkgs {
			if pkg.Tap == tap {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("tap %q not referenced by any package", tap)
		}
	}
}

func TestServices(t *testing.T) {
	svcs := Services()
	if len(svcs) != 1 {
		t.Fatalf("expected one brew service, got %d", len(svcs))
	}
	if svcs[0].Name != "colima" {
		t.Errorf("expected colima service, got %s", svcs[0].Name)
	}
}

func TestConfigs(t *testing.T) {
	configs := Configs()
	if len(configs) == 0 {
		t.Fatal("expected at least one config")
	}
	for i, cfg := range configs {
		if cfg.Source == "" {
			t.Errorf("config %d has empty Source", i)
		}
		if cfg.Dest == "" {
			t.Errorf("config %d has empty Dest", i)
		}
		if cfg.Mode == 0 {
			t.Errorf("config %d (%s) has zero Mode", i, cfg.Source)
		}
	}
}
