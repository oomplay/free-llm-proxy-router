package proxy

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kaiser-data/free-llm-proxy-router/pkg/catalog"
	"github.com/kaiser-data/free-llm-proxy-router/pkg/config"
)

// TestHandleModelsAdvertisesCanonicalAliases verifies that /v1/models lists
// both the raw upstream IDs and the canonical alias names derived from them.
func TestHandleModelsAdvertisesCanonicalAliases(t *testing.T) {
	cfg := &config.Config{}
	cfg.Models.Canonicalization.Enabled = true
	cfg.Models.Canonicalization.FreeOnly = true

	cat := &catalog.Catalog{Entries: []catalog.CatalogEntry{
		{ProviderID: "openrouter", ModelID: "qwen/qwen3-235b-a22b:free", IsFree: true},
		{ProviderID: "huggingface", ModelID: "Qwen/Qwen3-32B", IsFree: true},
		{ProviderID: "huggingface", ModelID: "paid/model", IsFree: false}, // excluded
	}}

	s := NewServer(cfg, cat, nil, nil, nil, nil)

	rec := httptest.NewRecorder()
	s.handleModels(rec, httptest.NewRequest(http.MethodGet, "/v1/models", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var body struct {
		Object string `json:"object"`
		Data   []struct {
			ID     string `json:"id"`
			Object string `json:"object"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if body.Object != "list" {
		t.Errorf("object = %q, want %q", body.Object, "list")
	}

	ids := map[string]bool{}
	for _, m := range body.Data {
		if m.Object != "model" {
			t.Errorf("entry %q: object = %q, want %q", m.ID, m.Object, "model")
		}
		ids[m.ID] = true
	}

	// Raw upstream IDs stay listed unchanged.
	for _, raw := range []string{"qwen/qwen3-235b-a22b:free", "Qwen/Qwen3-32B"} {
		if !ids[raw] {
			t.Errorf("raw model %q missing from /v1/models", raw)
		}
	}
	// Canonical aliases are advertised.
	if !ids["qwen3"] {
		t.Errorf("canonical alias %q missing from /v1/models; got %v", "qwen3", ids)
	}
	// Non-free catalog entries must not appear.
	if ids["paid/model"] {
		t.Errorf("non-free model %q must not be listed", "paid/model")
	}
}

// TestHandleModelsWithoutAliases verifies raw-only listing when the alias
// layer is disabled.
func TestHandleModelsWithoutAliases(t *testing.T) {
	cfg := &config.Config{} // canonicalization disabled → no aliases
	cat := &catalog.Catalog{Entries: []catalog.CatalogEntry{
		{ProviderID: "huggingface", ModelID: "Qwen/Qwen3-32B", IsFree: true},
	}}
	s := NewServer(cfg, cat, nil, nil, nil, nil)

	rec := httptest.NewRecorder()
	s.handleModels(rec, httptest.NewRequest(http.MethodGet, "/v1/models", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var body struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if len(body.Data) != 1 || body.Data[0].ID != "Qwen/Qwen3-32B" {
		t.Fatalf("data = %+v, want only the raw model", body.Data)
	}
}
