package trainer

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/alexandrmotologa/distillforge/internal/config"
	"github.com/alexandrmotologa/distillforge/internal/store"
)

// MetricCallback is called when a new training step metric arrives.
type MetricCallback func(metric *store.TrainingMetric)

// LogCallback is called when a log line is emitted from the training process.
type LogCallback func(line string)

// PipelineTrainer coordinates the fine-tuning execution.
type PipelineTrainer struct {
	cfg   *config.Config
	store *store.Store
}

// NewPipelineTrainer creates a training orchestrator.
func NewPipelineTrainer(cfg *config.Config, s *store.Store) *PipelineTrainer {
	return &PipelineTrainer{
		cfg:   cfg,
		store: s,
	}
}

func findScript(scriptName string) string {
	candidates := []string{
		filepath.Join("python", scriptName),
		filepath.Join("..", "python", scriptName),
		filepath.Join("..", "..", "python", scriptName),
	}
	for _, c := range candidates {
		if abs, err := filepath.Abs(c); err == nil {
			if _, err := os.Stat(abs); err == nil {
				return abs
			}
		}
	}
	return filepath.Join("python", scriptName)
}

// Train executes the fine-tuning script, streams metrics, and records progress.
func (pt *PipelineTrainer) Train(ctx context.Context, jobID, datasetPath, outputDir string, forceSimulation bool, onMetric MetricCallback, onLog LogCallback) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create checkpoint output dir: %w", err)
	}

	pythonPath, err := exec.LookPath("python")
	if err != nil {
		pythonPath, err = exec.LookPath("python3")
		if err != nil {
			return fmt.Errorf("python interpreter not found in system PATH: %w", err)
		}
	}

	scriptName := "train_lora.py"
	if pt.cfg.Training.Method == "mlx" {
		scriptName = "train_mlx.py"
	}
	scriptPath := findScript(scriptName)

	args := []string{
		scriptPath,
		"--dataset", datasetPath,
		"--output_dir", outputDir,
		"--base_model", pt.cfg.Student.BaseModel,
		"--method", pt.cfg.Training.Method,
		"--r", strconv.Itoa(pt.cfg.Training.R),
		"--lora_alpha", strconv.Itoa(pt.cfg.Training.LoRAAlpha),
		"--lora_dropout", fmt.Sprintf("%f", pt.cfg.Training.LoRADropout),
		"--learning_rate", fmt.Sprintf("%e", pt.cfg.Training.LearningRate),
		"--epochs", strconv.Itoa(pt.cfg.Training.Epochs),
		"--batch_size", strconv.Itoa(pt.cfg.Training.BatchSize),
		"--grad_accum", strconv.Itoa(pt.cfg.Training.GradientAccumulationSteps),
		"--max_seq_length", strconv.Itoa(pt.cfg.Training.MaxSeqLength),
	}

	if forceSimulation || pt.cfg.Teacher.Provider == "mock" {
		args = append(args, "--simulation")
	}

	cmd := exec.CommandContext(ctx, pythonPath, args...)

	metricHandler := func(m *store.TrainingMetric) {
		if pt.store != nil {
			_ = pt.store.RecordMetric(m)
			// Calculate progress percentage based on total expected steps
			totalSteps := pt.cfg.Training.Epochs * 10
			if totalSteps > 0 {
				pct := (float64(m.Step) / float64(totalSteps)) * 100.0
				if pct > 100.0 {
					pct = 100.0
				}
				_ = pt.store.UpdateJobProgress(jobID, pct, "running", "")
			}
		}
		if onMetric != nil {
			onMetric(m)
		}
	}

	monitor, err := NewProcessMonitor(cmd, metricHandler, onLog)
	if err != nil {
		return fmt.Errorf("failed to create process monitor: %w", err)
	}

	if pt.store != nil {
		_ = pt.store.UpdateJobProgress(jobID, 0.0, "running", "")
	}

	if err := monitor.StartAndStream(ctx, jobID); err != nil {
		if pt.store != nil {
			_ = pt.store.UpdateJobProgress(jobID, 0.0, "failed", err.Error())
		}
		return fmt.Errorf("training run failed: %w", err)
	}

	if pt.store != nil {
		_ = pt.store.UpdateJobProgress(jobID, 100.0, "completed", "")
	}

	return nil
}
