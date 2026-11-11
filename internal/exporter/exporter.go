package exporter

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/alexandrmotologa/distillforge/internal/config"
)

// PipelineExporter handles model conversion, GGUF quantization, and Ollama deployment packaging.
type PipelineExporter struct {
	cfg *config.Config
}

// NewPipelineExporter creates an exporter instance.
func NewPipelineExporter(cfg *config.Config) *PipelineExporter {
	return &PipelineExporter{cfg: cfg}
}

// ExportModel merges adapters, quantizes to GGUF, and generates the Ollama Modelfile.
func (e *PipelineExporter) ExportModel(ctx context.Context, checkpointDir, outputDir string, forceSimulation bool) (string, string, error) {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", "", fmt.Errorf("failed to create export output dir: %w", err)
	}

	modelBase := filepath.Base(e.cfg.Student.BaseModel)
	cleanName := strings.ToLower(strings.ReplaceAll(modelBase, "/", "-"))
	quant := e.cfg.Student.QuantizationTarget
	if quant == "" {
		quant = "q4_k_m"
	}

	ggufFileName := fmt.Sprintf("%s-%s.gguf", cleanName, quant)
	ggufPath := filepath.Join(outputDir, ggufFileName)

	// 1. Export GGUF using export_gguf.py
	pythonPath, err := exec.LookPath("python")
	if err != nil {
		pythonPath, err = exec.LookPath("python3")
		if err != nil {
			return "", "", fmt.Errorf("python interpreter not found in system PATH: %w", err)
		}
	}

	scriptPath := findExportScript()
	args := []string{
		scriptPath,
		"--checkpoint", checkpointDir,
		"--base_model", e.cfg.Student.BaseModel,
		"--output_file", ggufPath,
		"--quant", quant,
	}

	if forceSimulation || e.cfg.Teacher.Provider == "mock" {
		args = append(args, "--simulation")
	}

	cmd := exec.CommandContext(ctx, pythonPath, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", "", fmt.Errorf("gguf export failed: %w (output: %s)", err, string(out))
	}

	// 2. Generate Ollama Modelfile
	modelfileDest := e.cfg.Export.OllamaModelfilePath
	if modelfileDest == "" {
		modelfileDest = filepath.Join(outputDir, "Modelfile")
	}

	opts := ModelfileOptions{
		GGUFPath:     ggufPath,
		BaseModel:    e.cfg.Student.BaseModel,
		Temperature:  0.7,
		TopP:         0.9,
		SystemPrompt: fmt.Sprintf("You are %s, an enterprise Small Language Model distilled with distillforge.", e.cfg.Project.Name),
	}

	if err := WriteModelfile(modelfileDest, opts); err != nil {
		return "", "", fmt.Errorf("failed to write Ollama Modelfile: %w", err)
	}

	return ggufPath, modelfileDest, nil
}

func findExportScript() string {
	candidates := []string{
		filepath.Join("python", "export_gguf.py"),
		filepath.Join("..", "python", "export_gguf.py"),
		filepath.Join("..", "..", "python", "export_gguf.py"),
	}
	for _, c := range candidates {
		if abs, err := filepath.Abs(c); err == nil {
			if _, err := os.Stat(abs); err == nil {
				return abs
			}
		}
	}
	return filepath.Join("python", "export_gguf.py")
}
