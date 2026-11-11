package generator

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"

	"github.com/alexandrmotologa/distillforge/internal/config"
	"github.com/alexandrmotologa/distillforge/internal/store"
	"github.com/google/uuid"
)

// SeedItem represents a single row read from the seed data file.
type SeedItem struct {
	Prompt   string `json:"prompt"`
	Response string `json:"response,omitempty"`
}

// ProgressCallback is invoked as samples are generated to stream progress updates.
type ProgressCallback func(completed, total int, currentProgress float64)

// PipelineGenerator manages the synthetic generation workflow.
type PipelineGenerator struct {
	cfg        *config.Config
	client     TeacherClient
	evolEngine *EvolInstructEngine
	store      *store.Store
}

// NewPipelineGenerator creates a new generator instance.
func NewPipelineGenerator(cfg *config.Config, s *store.Store) (*PipelineGenerator, error) {
	client, err := NewTeacherClient(&cfg.Teacher)
	if err != nil {
		return nil, fmt.Errorf("failed to create teacher client: %w", err)
	}

	evolEngine := NewEvolInstructEngine(client, cfg.Generation.EvolutionStrategies)

	return &PipelineGenerator{
		cfg:        cfg,
		client:     client,
		evolEngine: evolEngine,
		store:      s,
	}, nil
}

// LoadSeedData reads and parses seed examples from a JSONL file.
func (g *PipelineGenerator) LoadSeedData(path, inputKey, targetKey string) ([]SeedItem, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open seed file %q: %w", path, err)
	}
	defer file.Close()

	var items []SeedItem
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var rawMap map[string]any
		if err := json.Unmarshal(line, &rawMap); err != nil {
			return nil, fmt.Errorf("invalid json on line %d: %w", lineNum, err)
		}

		promptVal, ok := rawMap[inputKey]
		if !ok {
			return nil, fmt.Errorf("missing input_key %q on line %d", inputKey, lineNum)
		}

		promptStr := fmt.Sprintf("%v", promptVal)
		responseStr := ""
		if targetVal, exists := rawMap[targetKey]; exists {
			responseStr = fmt.Sprintf("%v", targetVal)
		}

		items = append(items, SeedItem{
			Prompt:   promptStr,
			Response: responseStr,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading seed file: %w", err)
	}

	if len(items) == 0 {
		return nil, fmt.Errorf("seed file %q contains no valid entries", path)
	}

	return items, nil
}

// GenerateDataset creates synthetic training pairs and writes them to an output JSONL file.
func (g *PipelineGenerator) GenerateDataset(ctx context.Context, jobID, outputPath string, cb ProgressCallback) ([]store.Sample, error) {
	seeds, err := g.LoadSeedData(g.cfg.SeedData.File, g.cfg.SeedData.InputKey, g.cfg.SeedData.TargetKey)
	if err != nil {
		return nil, err
	}

	targetSize := g.cfg.Generation.TargetDatasetSize
	if targetSize <= 0 {
		targetSize = 100
	}

	concurrency := g.cfg.Generation.Concurrency
	if concurrency <= 0 {
		concurrency = 4
	}

	// Create directory for output file if needed
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create output dir: %w", err)
	}

	outFile, err := os.Create(outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create output file %q: %w", outputPath, err)
	}
	defer outFile.Close()

	writer := bufio.NewWriter(outFile)
	var writeMu sync.Mutex

	type task struct {
		Index    int
		Seed     SeedItem
		Strategy string
	}

	tasks := make(chan task, targetSize)
	for i := 0; i < targetSize; i++ {
		seed := seeds[i%len(seeds)]
		strategy := g.evolEngine.SelectStrategy()
		tasks <- task{
			Index:    i,
			Seed:     seed,
			Strategy: strategy,
		}
	}
	close(tasks)

	var completedCount int64
	var samplesMu sync.Mutex
	var generatedSamples []store.Sample

	var wg sync.WaitGroup
	errChan := make(chan error, concurrency)

	for w := 0; w < concurrency; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for t := range tasks {
				select {
				case <-ctx.Done():
					return
				default:
				}

				// 1. Evolve prompt
				evolvedPrompt, err := g.evolEngine.EvolvePrompt(ctx, t.Seed.Prompt, t.Strategy)
				if err != nil {
					// Fallback to seed on failure
					evolvedPrompt = t.Seed.Prompt
				}

				// 2. Query Teacher model for response
				systemRule := "You are an expert AI system executing the following domain task with extreme precision and valid structured formatting."
				teacherResp, err := g.client.Generate(ctx, systemRule, evolvedPrompt)
				if err != nil {
					errChan <- fmt.Errorf("teacher generation failed: %w", err)
					return
				}

				sample := store.Sample{
					ID:         uuid.New().String(),
					Prompt:     evolvedPrompt,
					Response:   teacherResp,
					Strategy:   t.Strategy,
					JudgeScore: 0.0, // to be populated in audit step
					IsAccepted: true,
				}

				// Write line to file
				row := map[string]string{
					"id":       sample.ID,
					"prompt":   sample.Prompt,
					"response": sample.Response,
					"strategy": sample.Strategy,
				}
				rowBytes, _ := json.Marshal(row)

				writeMu.Lock()
				writer.Write(rowBytes)
				writer.WriteString("\n")
				writer.Flush()
				writeMu.Unlock()

				samplesMu.Lock()
				generatedSamples = append(generatedSamples, sample)
				samplesMu.Unlock()

				currentDone := atomic.AddInt64(&completedCount, 1)
				if cb != nil {
					pct := float64(currentDone) / float64(targetSize) * 100.0
					cb(int(currentDone), targetSize, pct)
				}
			}
		}()
	}

	wg.Wait()
	close(errChan)

	if err, hasErr := <-errChan; hasErr {
		return nil, err
	}

	return generatedSamples, nil
}
