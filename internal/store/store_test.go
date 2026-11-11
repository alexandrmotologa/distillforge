package store

import (
	"path/filepath"
	"testing"
)

func TestStoreOperations(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "distillforge_test.db")

	s, err := NewStore(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer s.Close()

	// 1. Projects
	p, err := s.SaveProject("ticket-classifier", "Distill GPT-4o into Llama-3.2-1B")
	if err != nil {
		t.Fatalf("failed to save project: %v", err)
	}
	if p.ID == "" {
		t.Fatal("expected non-empty project ID")
	}

	projects, err := s.ListProjects()
	if err != nil || len(projects) != 1 {
		t.Fatalf("expected 1 project, got %d (err: %v)", len(projects), err)
	}

	// 2. Jobs
	job, err := s.CreateJob(p.ID, "pipeline", `{"mock":true}`)
	if err != nil {
		t.Fatalf("failed to create job: %v", err)
	}
	if job.Status != "pending" {
		t.Fatalf("expected pending status, got %s", job.Status)
	}

	if err := s.UpdateJobProgress(job.ID, 50.0, "running", ""); err != nil {
		t.Fatalf("failed to update job: %v", err)
	}

	fetchedJob, err := s.GetJob(job.ID)
	if err != nil {
		t.Fatalf("failed to get job: %v", err)
	}
	if fetchedJob.Progress != 50.0 || fetchedJob.Status != "running" {
		t.Fatalf("expected 50%% progress and running status, got %f / %s", fetchedJob.Progress, fetchedJob.Status)
	}

	// 3. Datasets and Samples
	dataset, err := s.SaveDataset(p.ID, "synthetic-tickets", "./data/clean.jsonl", "jsonl", "clean", 2)
	if err != nil {
		t.Fatalf("failed to save dataset: %v", err)
	}

	samples := []Sample{
		{
			DatasetID:  dataset.ID,
			Prompt:     "My internet is down after storm",
			Response:   `{"category": "Network Outage", "confidence": 0.98}`,
			Strategy:   "in_depth",
			JudgeScore: 0.95,
			IsAccepted: true,
		},
		{
			DatasetID:  dataset.ID,
			Prompt:     "Cancel my subscription renewal",
			Response:   `{"category": "Billing", "confidence": 0.99}`,
			Strategy:   "adversarial",
			JudgeScore: 0.92,
			IsAccepted: true,
		},
	}
	if err := s.InsertSampleBatch(samples); err != nil {
		t.Fatalf("failed to insert samples: %v", err)
	}

	fetchedSamples, err := s.ListSamples(dataset.ID, 10, 0)
	if err != nil || len(fetchedSamples) != 2 {
		t.Fatalf("expected 2 samples, got %d (err: %v)", len(fetchedSamples), err)
	}

	// 4. Training Metrics
	metric := &TrainingMetric{
		JobID:        job.ID,
		Step:         100,
		Epoch:        1.0,
		Loss:         0.42,
		LearningRate: 0.0002,
		Perplexity:   1.52,
		VRAMMB:       3850.5,
	}
	if err := s.RecordMetric(metric); err != nil {
		t.Fatalf("failed to record metric: %v", err)
	}

	metrics, err := s.GetMetricsByJob(job.ID)
	if err != nil || len(metrics) != 1 {
		t.Fatalf("expected 1 metric, got %d (err: %v)", len(metrics), err)
	}
	if metrics[0].Loss != 0.42 {
		t.Fatalf("expected loss 0.42, got %f", metrics[0].Loss)
	}

	// 5. Benchmarks
	bm := &Benchmark{
		JobID:             job.ID,
		StudentModel:      "Llama-3.2-1B-Instruct-LoRA",
		TeacherModel:      "gpt-4o",
		ExactMatch:        0.94,
		SchemaAdherence:   1.0,
		AvgLatencyMs:      18.4,
		P95LatencyMs:      24.1,
		TokensPerSec:      68.2,
		MonthlySavingsUSD: 3840.0,
		EvaluationSummary: "Student model matched GPT-4o accuracy with 96% cost savings and 18ms latency.",
	}
	if err := s.SaveBenchmark(bm); err != nil {
		t.Fatalf("failed to save benchmark: %v", err)
	}

	fetchedBM, err := s.GetBenchmarkByJob(job.ID)
	if err != nil {
		t.Fatalf("failed to fetch benchmark: %v", err)
	}
	if fetchedBM.ExactMatch != 0.94 || fetchedBM.AvgLatencyMs != 18.4 {
		t.Fatalf("unexpected benchmark values: %+v", fetchedBM)
	}
}
