package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("writing config: %v", err)
	}
	return path
}

// TestModelsExposureDefaults verifies that a config without a models section
// gets the backward-compatible exposure defaults: both raw and canonical
// entries advertised.
func TestModelsExposureDefaults(t *testing.T) {
	path := writeTempConfig(t, "proxy:\n  port: 9999\n")
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !cfg.Models.ExposeRaw {
		t.Errorf("expose_raw default = false, want true")
	}
	if !cfg.Models.ExposeCanonical {
		t.Errorf("expose_canonical default = false, want true")
	}
}

// TestModelsExposureParsing verifies explicit exposure flags are honoured,
// including disabling raw exposure for canonical-only listing.
func TestModelsExposureParsing(t *testing.T) {
	path := writeTempConfig(t, "models:\n  expose_raw: false\n  expose_canonical: true\n")
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Models.ExposeRaw {
		t.Errorf("expose_raw = true, want false")
	}
	if !cfg.Models.ExposeCanonical {
		t.Errorf("expose_canonical = false, want true")
	}
}

// TestModelsExposureParsingRawOnly verifies raw-only exposure parsing.
func TestModelsExposureParsingRawOnly(t *testing.T) {
	path := writeTempConfig(t, "models:\n  expose_raw: true\n  expose_canonical: false\n")
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !cfg.Models.ExposeRaw {
		t.Errorf("expose_raw = false, want true")
	}
	if cfg.Models.ExposeCanonical {
		t.Errorf("expose_canonical = true, want false")
	}
}