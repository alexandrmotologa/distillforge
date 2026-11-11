package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigLoadAndDefaults(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "distillforge.yaml")

	yamlContent := `
version: "1"
project:
  name: "ticket-classifier"
  description: "Classify support tickets into 12 categories"

teacher:
  provider: "mock"
  model: "mock-teacher-v1"

student:
  base_model: "meta-llama/Llama-3.2-1B-Instruct"

seed_data:
  file: "./data/sample.jsonl"
`

	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("unexpected error loading config: %v", err)
	}

	if cfg.Project.Name != "ticket-classifier" {
		t.Errorf("expected project name 'ticket-classifier', got %q", cfg.Project.Name)
	}
	if cfg.Teacher.Provider != "mock" {
		t.Errorf("expected teacher provider 'mock', got %q", cfg.Teacher.Provider)
	}
	if cfg.Student.QuantizationTarget != "q4_k_m" {
		t.Errorf("expected default quantization target 'q4_k_m', got %q", cfg.Student.QuantizationTarget)
	}
	if cfg.Training.Method != "qlora" {
		t.Errorf("expected default training method 'qlora', got %q", cfg.Training.Method)
	}
	if cfg.Generation.TargetDatasetSize != 1000 {
		t.Errorf("expected default target dataset size 1000, got %d", cfg.Generation.TargetDatasetSize)
	}
}

func TestConfigValidationFailures(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr string
	}{
		{
			name: "missing project name",
			yaml: `
teacher:
  provider: "mock"
student:
  base_model: "test-model"
seed_data:
  file: "data.jsonl"
`,
			wantErr: "project.name is required",
		},
		{
			name: "invalid teacher provider",
			yaml: `
project:
  name: "test"
teacher:
  provider: "unknown-cloud"
student:
  base_model: "test-model"
seed_data:
  file: "data.jsonl"
`,
			wantErr: "unsupported teacher provider",
		},
		{
			name: "missing seed file",
			yaml: `
project:
  name: "test"
teacher:
  provider: "mock"
student:
  base_model: "test-model"
`,
			wantErr: "seed_data.file is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			configPath := filepath.Join(tempDir, "invalid.yaml")
			if err := os.WriteFile(configPath, []byte(tt.yaml), 0644); err != nil {
				t.Fatalf("failed to write test config: %v", err)
			}

			_, err := Load(configPath)
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.wantErr)
			}
		})
	}
}
