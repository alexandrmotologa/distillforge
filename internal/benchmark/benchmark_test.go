package benchmark

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/alexandrmotologa/distillforge/internal/config"
	"github.com/alexandrmotologa/distillforge/internal/store"
)

func TestLatencyProfiler(t *testing.T) {
	p := NewLatencyProfiler()
	p.Record(10*time.Millisecond, 2*time.Millisecond, 50)
	p.Record(20*time.Millisecond, 5*time.Millisecond, 60)
	p.Record(30*time.Millisecond, 8*time.Millisecond, 70)

	res := p.Calculate()
	if res.TotalRuns != 3 {
		t.Fatalf("expected 3 total runs, got %d", res.TotalRuns)
	}
	if res.AvgLatencyMs != 20.0 {
		t.Errorf("expected avg latency 20.0ms, got %f", res.AvgLatencyMs)
	}
	if res.MinLatencyMs != 10.0 {
		t.Errorf("expected min latency 10.0ms, got %f", res.MinLatencyMs)
	}
	if res.MaxLatencyMs != 30.0 {
		t.Errorf("expected max latency 30.0ms, got %f", res.MaxLatencyMs)
	}
}

func TestROICalculator(t *testing.T) {
	cfg := &config.FinOpsConfig{
		MonthlyRequests:            1000000,
		AvgPromptTokens:            350,
		AvgCompletionTokens:        120,
		TeacherCostPerMIn:          5.00,
		TeacherCostPerMOut:         15.00,
		SelfHostedMonthlyInfraCost: 120.00,
	}

	roi := CalculateROI(cfg, 850.0, 18.0)

	// Teacher Input: 1M * 350 = 350M tokens * $5 / M = $1,750
	// Teacher Output: 1M * 120 = 120M tokens * $15 / M = $1,800
	// Teacher Total: $3,550
	// Student: $120
	// Savings: $3,550 - $120 = $3,430 / month = $41,160 / year

	if roi.TeacherMonthlyCostUSD != 3550.0 {
		t.Errorf("expected teacher cost $3,550, got $%f", roi.TeacherMonthlyCostUSD)
	}
	if roi.MonthlySavingsUSD != 3430.0 {
		t.Errorf("expected monthly savings $3,430, got $%f", roi.MonthlySavingsUSD)
	}
	if roi.CostReductionPercentage < 95.0 {
		t.Errorf("expected >95%% cost reduction, got %f%%", roi.CostReductionPercentage)
	}
	if roi.EstimatedKgCO2SavedPerYear <= 0 {
		t.Errorf("expected positive carbon savings, got %f", roi.EstimatedKgCO2SavedPerYear)
	}
}

func TestPipelineBenchmarkEvaluation(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "bm_test.db")

	st, err := store.NewStore(dbPath)
	if err != nil {
		t.Fatalf("failed to init store: %v", err)
	}
	defer st.Close()

	cfg := &config.Config{
		Project: config.ProjectConfig{Name: "bm-test"},
		Teacher: config.TeacherConfig{
			Provider: "mock",
			Model:    "gpt-4o",
		},
		Student: config.StudentConfig{
			BaseModel: "Llama-3.2-1B-Instruct",
		},
		Evaluation: config.EvaluationConfig{
			SampleSize: 2,
		},
	}
	cfg.SetDefaults()

	bm, err := NewPipelineBenchmark(cfg, st)
	if err != nil {
		t.Fatalf("failed to create pipeline benchmark: %v", err)
	}

	testSamples := []store.Sample{
		{Prompt: "Ticket regarding network outage in branch 4", Response: `{"category": "Network Outage", "confidence": 0.96}`},
		{Prompt: "Billing question on credit card renewal", Response: `{"category": "Billing", "confidence": 0.96}`},
	}

	summary, err := bm.EvaluateHoldout(context.Background(), "job-bm-1", testSamples)
	if err != nil {
		t.Fatalf("evaluation failed: %v", err)
	}

	if summary.TotalEvaluated != 2 {
		t.Fatalf("expected 2 evaluated samples, got %d", summary.TotalEvaluated)
	}
	if summary.ExactMatchRate != 1.0 {
		t.Errorf("expected 100%% exact match rate under mock, got %f", summary.ExactMatchRate)
	}
	if len(summary.Comparisons) != 2 {
		t.Fatalf("expected 2 comparisons, got %d", len(summary.Comparisons))
	}
}
