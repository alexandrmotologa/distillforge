package trainer

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/alexandrmotologa/distillforge/internal/config"
	"github.com/alexandrmotologa/distillforge/internal/store"
)

func TestTrainerSimulation(t *testing.T) {
	tempDir := t.TempDir()
	datasetPath := filepath.Join(tempDir, "clean.jsonl")
	outputDir := filepath.Join(tempDir, "checkpoints")
	dbPath := filepath.Join(tempDir, "test.db")

	if err := os.WriteFile(datasetPath, []byte(`{"prompt": "Hi", "response": "Hello"}`), 0644); err != nil {
		t.Fatalf("failed to write dummy dataset: %v", err)
	}

	st, err := store.NewStore(dbPath)
	if err != nil {
		t.Fatalf("failed to init store: %v", err)
	}
	defer st.Close()

	p, err := st.SaveProject("test-train", "Unit test training execution")
	if err != nil {
		t.Fatalf("failed to save project: %v", err)
	}

	job, err := st.CreateJob(p.ID, "train", "{}")
	if err != nil {
		t.Fatalf("failed to create job: %v", err)
	}

	cfg := &config.Config{
		Project: config.ProjectConfig{Name: "test-train"},
		Teacher: config.TeacherConfig{Provider: "mock"},
		Student: config.StudentConfig{BaseModel: "meta-llama/Llama-3.2-1B-Instruct"},
		Training: config.TrainingConfig{
			Method:       "qlora",
			Epochs:       1,
			R:            16,
			LearningRate: 0.0002,
		},
	}
	cfg.SetDefaults()

	trainer := NewPipelineTrainer(cfg, st)

	var mu sync.Mutex
	var recordedMetrics []*store.TrainingMetric
	var logs []string

	err = trainer.Train(context.Background(), job.ID, datasetPath, outputDir, true,
		func(m *store.TrainingMetric) {
			mu.Lock()
			recordedMetrics = append(recordedMetrics, m)
			mu.Unlock()
		},
		func(line string) {
			mu.Lock()
			logs = append(logs, line)
			mu.Unlock()
		},
	)
	if err != nil {
		t.Fatalf("training failed: %v", err)
	}

	if len(recordedMetrics) == 0 {
		t.Fatal("expected at least one metric to be emitted")
	}

	// Verify adapter_config.json was created in outputDir
	adapterConfigPath := filepath.Join(outputDir, "adapter_config.json")
	if _, err := os.Stat(adapterConfigPath); err != nil {
		t.Fatalf("expected adapter_config.json to exist in %s, err: %v", outputDir, err)
	}

	// Verify job status in store
	updatedJob, err := st.GetJob(job.ID)
	if err != nil {
		t.Fatalf("failed to get job: %v", err)
	}
	if updatedJob.Status != "completed" {
		t.Fatalf("expected job status 'completed', got %q", updatedJob.Status)
	}
}
