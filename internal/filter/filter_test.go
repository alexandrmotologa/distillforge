package filter

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/alexandrmotologa/distillforge/internal/config"
	"github.com/alexandrmotologa/distillforge/internal/store"
)

func TestSchemaVerifier(t *testing.T) {
	tempDir := t.TempDir()
	schemaPath := filepath.Join(tempDir, "schema.json")

	schemaContent := `{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type": "object",
		"required": ["category", "confidence"],
		"properties": {
			"category": {"type": "string"},
			"confidence": {"type": "number"}
		}
	}`
	if err := os.WriteFile(schemaPath, []byte(schemaContent), 0644); err != nil {
		t.Fatalf("failed to write schema: %v", err)
	}

	verifier, err := NewSchemaVerifier(schemaPath)
	if err != nil {
		t.Fatalf("failed to create schema verifier: %v", err)
	}

	// 1. Valid JSON matching schema
	validResp := `{"category": "Network Outage", "confidence": 0.95}`
	ok, err := verifier.Validate(validResp)
	if !ok || err != nil {
		t.Fatalf("expected valid schema match, got ok=%v, err=%v", ok, err)
	}

	// 2. Missing required field
	missingResp := `{"category": "Network Outage"}`
	ok, err = verifier.Validate(missingResp)
	if ok || err == nil {
		t.Fatal("expected failure due to missing required field")
	}

	// 3. Invalid JSON syntax
	brokenResp := `not-json`
	ok, err = verifier.Validate(brokenResp)
	if ok || err == nil {
		t.Fatal("expected failure due to invalid json")
	}
}

func TestDeduplicator(t *testing.T) {
	dedup := NewDeduplicator(0.80)

	text1 := "The quick brown fox jumps over the lazy dog in the forest"
	text2 := "The quick brown fox jumps over the lazy dog in the woods" // high similarity
	text3 := "Completely distinct prompt regarding database connection timeout error"

	if dedup.IsDuplicate(text1) {
		t.Fatal("first text should not be duplicate")
	}
	if !dedup.IsDuplicate(text2) {
		t.Fatal("text2 should be recognized as duplicate of text1")
	}
	if dedup.IsDuplicate(text3) {
		t.Fatal("text3 should be distinct and not marked duplicate")
	}
}

func TestPipelineFilterAuditAndExport(t *testing.T) {
	tempDir := t.TempDir()
	inputPath := filepath.Join(tempDir, "raw.jsonl")
	cleanPath := filepath.Join(tempDir, "clean.jsonl")
	alpacaPath := filepath.Join(tempDir, "dataset_alpaca.json")

	rawContent := `{"prompt": "Internet is down", "response": "{\"category\":\"Network Outage\",\"confidence\":0.98}"}
{"prompt": "Internet is down", "response": "{\"category\":\"Network Outage\",\"confidence\":0.98}"}
{"prompt": "Short", "response": "tiny"}
{"prompt": "How do I pay?", "response": "{\"category\":\"Billing\",\"confidence\":0.95}"}
`
	if err := os.WriteFile(inputPath, []byte(rawContent), 0644); err != nil {
		t.Fatalf("failed to write raw dataset: %v", err)
	}

	cfg := &config.Config{
		Project: config.ProjectConfig{Name: "test-filter"},
		Teacher: config.TeacherConfig{Provider: "mock", Model: "mock-v1"},
		Filter: config.FilterConfig{
			DedupThreshold:   0.85,
			MinResponseChars: 10,
			MaxResponseChars: 500,
		},
		Generation: config.GenerationConfig{
			QualityThreshold: 0.80,
		},
	}
	cfg.SetDefaults()

	filter, err := NewPipelineFilter(cfg)
	if err != nil {
		t.Fatalf("failed to create pipeline filter: %v", err)
	}

	report, accepted, err := filter.AuditDataset(context.Background(), inputPath, cleanPath)
	if err != nil {
		t.Fatalf("audit failed: %v", err)
	}

	if report.TotalProcessed != 4 {
		t.Errorf("expected 4 processed, got %d", report.TotalProcessed)
	}
	if report.DuplicateCount != 1 {
		t.Errorf("expected 1 duplicate, got %d", report.DuplicateCount)
	}
	if report.LengthFailCount != 1 {
		t.Errorf("expected 1 length fail, got %d", report.LengthFailCount)
	}
	if report.AcceptedCount != 2 {
		t.Errorf("expected 2 accepted samples, got %d", report.AcceptedCount)
	}
	if len(accepted) != 2 {
		t.Errorf("expected 2 accepted samples in slice, got %d", len(accepted))
	}

	// Test Alpaca Export
	if err := ExportFormat(accepted, alpacaPath, "alpaca"); err != nil {
		t.Fatalf("failed to export alpaca format: %v", err)
	}

	stat, err := os.Stat(alpacaPath)
	if err != nil || stat.Size() == 0 {
		t.Fatal("expected non-empty alpaca export file")
	}

	// Test ShareGPT Export
	shareGPTPath := filepath.Join(tempDir, "dataset_sharegpt.json")
	if err := ExportFormat([]store.Sample{accepted[0]}, shareGPTPath, "sharegpt"); err != nil {
		t.Fatalf("failed to export sharegpt: %v", err)
	}
}
