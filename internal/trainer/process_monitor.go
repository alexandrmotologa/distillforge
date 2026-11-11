package trainer

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/alexandrmotologa/distillforge/internal/store"
)

// ProcessMonitor wraps a running training command, streaming logs and parsing structured metrics.
type ProcessMonitor struct {
	cmd        *exec.Cmd
	stdout     io.ReadCloser
	stderr     io.ReadCloser
	onMetric   func(metric *store.TrainingMetric)
	onLog      func(line string)
}

// NewProcessMonitor creates a monitor for an exec.Cmd.
func NewProcessMonitor(cmd *exec.Cmd, onMetric func(*store.TrainingMetric), onLog func(string)) (*ProcessMonitor, error) {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to open stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to open stderr pipe: %w", err)
	}

	return &ProcessMonitor{
		cmd:      cmd,
		stdout:   stdout,
		stderr:   stderr,
		onMetric: onMetric,
		onLog:    onLog,
	}, nil
}

// StartAndStream runs the command and streams stdout/stderr until completion.
func (pm *ProcessMonitor) StartAndStream(ctx context.Context, jobID string) error {
	if err := pm.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start training process: %w", err)
	}

	errChan := make(chan error, 2)

	// Stream stdout
	go func() {
		scanner := bufio.NewScanner(pm.stdout)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "METRIC:") {
				jsonStr := strings.TrimPrefix(line, "METRIC:")
				var parsed struct {
					Step         int     `json:"step"`
					Epoch        float64 `json:"epoch"`
					Loss         float64 `json:"loss"`
					LearningRate float64 `json:"learning_rate"`
					Perplexity   float64 `json:"perplexity"`
					VRAMMB       float64 `json:"vram_mb"`
				}
				if err := json.Unmarshal([]byte(jsonStr), &parsed); err == nil {
					metric := &store.TrainingMetric{
						JobID:        jobID,
						Step:         parsed.Step,
						Epoch:        parsed.Epoch,
						Loss:         parsed.Loss,
						LearningRate: parsed.LearningRate,
						Perplexity:   parsed.Perplexity,
						VRAMMB:       parsed.VRAMMB,
					}
					if pm.onMetric != nil {
						pm.onMetric(metric)
					}
				}
			} else {
				if pm.onLog != nil {
					pm.onLog(line)
				}
			}
		}
		errChan <- scanner.Err()
	}()

	// Stream stderr
	go func() {
		scanner := bufio.NewScanner(pm.stderr)
		for scanner.Scan() {
			line := scanner.Text()
			if pm.onLog != nil {
				pm.onLog("[ERR] " + line)
			}
		}
		errChan <- scanner.Err()
	}()

	waitErr := pm.cmd.Wait()
	_ = <-errChan
	_ = <-errChan

	if waitErr != nil {
		return fmt.Errorf("training process exited with error: %w", waitErr)
	}

	return nil
}
