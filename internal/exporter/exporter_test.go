package exporter

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alexandrmotologa/distillforge/internal/config"
)

func TestModelfileBuilder(t *testing.T) {
	opts := ModelfileOptions{
		GGUFPath:     "./dist/llama3.2-1b-q4_k_m.gguf",
		BaseModel:    "meta-llama/Llama-3.2-1B-Instruct",
		SystemPrompt: "Custom prompt",
		Temperature:  0.8,
		TopP:         0.95,
	}

	content := BuildModelfileContent(opts)

	if !strings.Contains(content, "FROM ./dist/llama3.2-1b-q4_k_m.gguf") {
		t.Errorf("expected FROM clause with gguf path, got:\n%s", content)
	}
	if !strings.Contains(content, "<|start_header_id|>") {
		t.Errorf("expected Llama-3 header formatting, got:\n%s", content)
	}
	if !strings.Contains(content, "PARAMETER temperature 0.80") {
		t.Errorf("expected temperature 0.80, got:\n%s", content)
	}
	if !strings.Contains(content, "SYSTEM \"\"\"Custom prompt\"\"\"") {
		t.Errorf("expected system prompt, got:\n%s", content)
	}
}

func TestPipelineExporterExportModel(t *testing.T) {
	tempDir := t.TempDir()
	checkpointDir := filepath.Join(tempDir, "checkpoints")
	outputDir := filepath.Join(tempDir, "dist")
	modelfilePath := filepath.Join(tempDir, "dist", "Modelfile")

	if err := os.MkdirAll(checkpointDir, 0755); err != nil {
		t.Fatalf("failed to create dummy checkpoint dir: %v", err)
	}

	cfg := &config.Config{
		Project: config.ProjectConfig{Name: "ticket-classifier"},
		Teacher: config.TeacherConfig{Provider: "mock"},
		Student: config.StudentConfig{
			BaseModel:          "meta-llama/Llama-3.2-1B-Instruct",
			QuantizationTarget: "q4_k_m",
		},
		Export: config.ExportConfig{
			OllamaModelfilePath: modelfilePath,
			OutputDir:           outputDir,
		},
	}
	cfg.SetDefaults()

	exporter := NewPipelineExporter(cfg)
	ggufPath, mfPath, err := exporter.ExportModel(context.Background(), checkpointDir, outputDir, true)
	if err != nil {
		t.Fatalf("export failed: %v", err)
	}

	// Verify GGUF file exists and has magic header
	stat, err := os.Stat(ggufPath)
	if err != nil || stat.Size() < 4 {
		t.Fatalf("expected valid GGUF file at %s, err: %v", ggufPath, err)
	}

	// Verify Modelfile exists and contains FROM clause
	mfContent, err := os.ReadFile(mfPath)
	if err != nil {
		t.Fatalf("failed to read Modelfile: %v", err)
	}
	if !strings.Contains(string(mfContent), "FROM ") {
		t.Errorf("expected Modelfile to have FROM line, got:\n%s", string(mfContent))
	}

	// Test Multi-Quantization Benchmarking
	matrix, err := BenchmarkQuantizations(context.Background(), cfg, checkpointDir, outputDir, []string{"q4_k_m", "q8_0"}, true)
	if err != nil {
		t.Fatalf("quantization benchmarking failed: %v", err)
	}
	if len(matrix) != 2 {
		t.Fatalf("expected 2 quantization results, got %d", len(matrix))
	}
	if matrix[0].FileSizeMB <= 0 || matrix[1].FileSizeMB <= 0 {
		t.Fatal("expected positive file sizes in quant matrix")
	}
}
