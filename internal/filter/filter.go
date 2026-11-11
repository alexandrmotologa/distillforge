package filter

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alexandrmotologa/distillforge/internal/config"
	"github.com/alexandrmotologa/distillforge/internal/generator"
	"github.com/alexandrmotologa/distillforge/internal/store"
)

// AuditReport summarizes the results of the quality filtering step.
type AuditReport struct {
	TotalProcessed   int     `json:"total_processed"`
	AcceptedCount    int     `json:"accepted_count"`
	RejectedCount    int     `json:"rejected_count"`
	DuplicateCount   int     `json:"duplicate_count"`
	SchemaFailCount  int     `json:"schema_fail_count"`
	LowScoreCount    int     `json:"low_score_count"`
	LengthFailCount  int     `json:"length_fail_count"`
	AverageScore     float64 `json:"average_score"`
	PassRatePercent  float64 `json:"pass_rate_percent"`
}

// PipelineFilter manages the audit and dataset export process.
type PipelineFilter struct {
	cfg            *config.Config
	judge          *Judge
	schemaVerifier *SchemaVerifier
	dedup          *Deduplicator
}

// NewPipelineFilter initializes the filter engine.
func NewPipelineFilter(cfg *config.Config) (*PipelineFilter, error) {
	teacherClient, err := generator.NewTeacherClient(&cfg.Teacher)
	if err != nil {
		return nil, fmt.Errorf("failed to create teacher client for judge: %w", err)
	}

	verifier, err := NewSchemaVerifier(cfg.Filter.SchemaFile)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize schema verifier: %w", err)
	}

	return &PipelineFilter{
		cfg:            cfg,
		judge:          NewJudge(teacherClient),
		schemaVerifier: verifier,
		dedup:          NewDeduplicator(cfg.Filter.DedupThreshold),
	}, nil
}

// AuditDataset processes raw samples, scores them, removes duplicates, and filters out defective rows.
func (f *PipelineFilter) AuditDataset(ctx context.Context, inputPath, cleanOutputPath string) (*AuditReport, []store.Sample, error) {
	file, err := os.Open(inputPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open raw dataset %q: %w", inputPath, err)
	}
	defer file.Close()

	if err := os.MkdirAll(filepath.Dir(cleanOutputPath), 0755); err != nil {
		return nil, nil, err
	}

	outFile, err := os.Create(cleanOutputPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create clean dataset file: %w", err)
	}
	defer outFile.Close()

	writer := bufio.NewWriter(outFile)
	defer writer.Flush()

	scanner := bufio.NewScanner(file)
	report := &AuditReport{}
	var acceptedSamples []store.Sample
	var totalScore float64

	minLen := f.cfg.Filter.MinResponseChars
	maxLen := f.cfg.Filter.MaxResponseChars
	qualityThreshold := f.cfg.Generation.QualityThreshold

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		report.TotalProcessed++

		var raw map[string]any
		if err := json.Unmarshal(line, &raw); err != nil {
			continue
		}

		prompt := fmt.Sprintf("%v", raw["prompt"])
		response := fmt.Sprintf("%v", raw["response"])
		strategy := "seed"
		if st, ok := raw["strategy"]; ok {
			strategy = fmt.Sprintf("%v", st)
		}

		// 1. Length check
		respLen := len(strings.TrimSpace(response))
		if (minLen > 0 && respLen < minLen) || (maxLen > 0 && respLen > maxLen) {
			report.LengthFailCount++
			report.RejectedCount++
			continue
		}

		// 2. Deduplication check
		if f.dedup.IsDuplicate(prompt) {
			report.DuplicateCount++
			report.RejectedCount++
			continue
		}

		// 3. Schema validation
		validSchema, _ := f.schemaVerifier.Validate(response)
		if !validSchema {
			report.SchemaFailCount++
			report.RejectedCount++
			continue
		}

		// 4. LLM-as-a-Judge evaluation
		eval, err := f.judge.Evaluate(ctx, prompt, response)
		if err != nil || eval.OverallScore < qualityThreshold {
			report.LowScoreCount++
			report.RejectedCount++
			continue
		}

		// Accepted sample
		sample := store.Sample{
			Prompt:     prompt,
			Response:   response,
			Strategy:   strategy,
			JudgeScore: eval.OverallScore,
			IsAccepted: true,
		}

		cleanRow := map[string]any{
			"prompt":      prompt,
			"response":    response,
			"strategy":    strategy,
			"judge_score": eval.OverallScore,
			"feedback":    eval.Feedback,
		}
		rowBytes, _ := json.Marshal(cleanRow)
		writer.Write(rowBytes)
		writer.WriteString("\n")

		acceptedSamples = append(acceptedSamples, sample)
		report.AcceptedCount++
		totalScore += eval.OverallScore
	}

	if report.AcceptedCount > 0 {
		report.AverageScore = totalScore / float64(report.AcceptedCount)
		report.PassRatePercent = (float64(report.AcceptedCount) / float64(report.TotalProcessed)) * 100.0
	}

	return report, acceptedSamples, nil
}

// ExportFormat writes accepted samples in Alpaca, ShareGPT, or OpenAI JSONL formats.
func ExportFormat(samples []store.Sample, targetPath, format string) error {
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return err
	}

	file, err := os.Create(targetPath)
	if err != nil {
		return fmt.Errorf("failed to create export file %q: %w", targetPath, err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	switch strings.ToLower(format) {
	case "alpaca":
		var alpacaRows []map[string]string
		for _, s := range samples {
			alpacaRows = append(alpacaRows, map[string]string{
				"instruction": s.Prompt,
				"input":       "",
				"output":      s.Response,
			})
		}
		data, err := json.MarshalIndent(alpacaRows, "", "  ")
		if err != nil {
			return err
		}
		_, err = writer.Write(data)
		return err

	case "sharegpt":
		var shareGPTRows []map[string]any
		for _, s := range samples {
			shareGPTRows = append(shareGPTRows, map[string]any{
				"conversations": []map[string]string{
					{"from": "human", "value": s.Prompt},
					{"from": "gpt", "value": s.Response},
				},
			})
		}
		data, err := json.MarshalIndent(shareGPTRows, "", "  ")
		if err != nil {
			return err
		}
		_, err = writer.Write(data)
		return err

	case "jsonl", "openai":
		fallthrough
	default:
		for _, s := range samples {
			row := map[string]any{
				"messages": []map[string]string{
					{"role": "user", "content": s.Prompt},
					{"role": "assistant", "content": s.Response},
				},
			}
			bytes, _ := json.Marshal(row)
			writer.Write(bytes)
			writer.WriteString("\n")
		}
		return nil
	}
}
