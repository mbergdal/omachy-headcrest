package installer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPackagePresentOnMachineDetectsBinary(t *testing.T) {
	tmp := t.TempDir()
	bin := filepath.Join(tmp, "mise")
	if err := os.WriteFile(bin, []byte("#!/bin/sh\n"), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", tmp+":"+os.Getenv("PATH"))

	if !packagePresentOnMachine("mise") {
		t.Fatal("expected mise to be detected on PATH")
	}
}
