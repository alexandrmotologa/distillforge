package generator

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/alexandrmotologa/distillforge/internal/config"
	"github.com/alexandrmotologa/distillforge/internal/store"
)

func TestGeneratorEndToEnd(t *testing.T) {
	tempDir := t.TempDir()
	seedPath := filepath.Join(tempDir, "seed.jsonl")
	outputPath := filepath.Join(tempDir, "synthetic.jsonl")
	dbPath := filepath.Join(tempDir, "test.db")

	seedContent := `{"prompt": "Can you reset my router?", "response": "{\"category\":\"Network Outage\",\"confidence\":0.95}"}
{"prompt": "I was billed twice this month", "response": "{\"category\":\"Billing\",\"confidence\":0.99}"}
{"prompt": "Where do I update my password?", "response": "{\"category\":\"Authentication\",\"confidence\":0.98}"}
`
	if err := os.WriteFile(seedPath, []byte(seedContent), 0644); err != nil {
		t.Fatalf("failed to write seed file: %v", err)
	}

	st, err := store.NewStore(dbPath)
	if err != nil {
		t.Fatalf("failed to init store: %v", err)
	}
	defer st.Close()

	cfg := &config.Config{
		Project: config.ProjectConfig{Name: "test-distill"},
		Teacher: config.TeacherConfig{
			Provider: "mock",
			Model:    "mock-v1",
		},
		SeedData: config.SeedDataConfig{
			File:      seedPath,
			InputKey:  "prompt",
			TargetKey: "response",
		},
		Generation: config.GenerationConfig{
			TargetDatasetSize: 6,
			Concurrency:       2,
			EvolutionStrategies: []config.EvolutionStrategy{
				{Name: "in_depth", Weight: 0.5},
				{Name: "adversarial", Weight: 0.5},
			},
		},
	}
	cfg.SetDefaults()

	gen, err := NewPipelineGenerator(cfg, st)
	if err != nil {
		t.Fatalf("failed to create pipeline generator: %v", err)
	}

	var progressUpdates []float64
	var progressMu sync.Mutex
	samples, err := gen.GenerateDataset(context.Background(), "job-123", outputPath, func(done, total int, pct float64) {
		progressMu.Lock()
		defer progressMu.Unlock()
		progressUpdates = append(progressUpdates, pct)
	})
	if err != nil {
		t.Fatalf("generation failed: %v", err)
	}

	if len(samples) != 6 {
		t.Fatalf("expected 6 samples generated, got %d", len(samples))
	}
	if len(progressUpdates) != 6 {
		t.Fatalf("expected 6 progress updates, got %d", len(progressUpdates))
	}

	// Verify file was written
	stat, err := os.Stat(outputPath)
	if err != nil || stat.Size() == 0 {
		t.Fatalf("expected non-empty output file, err: %v", err)
	}
}
