package report

import (
	"strings"
	"testing"
	"time"

	"github.com/alexandrmotologa/distillforge/internal/config"
	"github.com/alexandrmotologa/distillforge/internal/store"
)

func TestReportGeneration(t *testing.T) {
	cfg := &config.Config{
		Project: config.ProjectConfig{Name: "ticket-classifier"},
		Teacher: config.TeacherConfig{Model: "gpt-4o"},
		Student: config.StudentConfig{BaseModel: "Llama-3.2-1B-Instruct"},
	}
	cfg.SetDefaults()

	job := &store.Job{
		ID:        "job-rep-1",
		Status:    "completed",
		CreatedAt: time.Now().UTC(),
	}

	bm := &store.Benchmark{
		JobID:             job.ID,
		ExactMatch:        0.98,
		SchemaAdherence:   1.0,
		AvgLatencyMs:      18.4,
		P95LatencyMs:      24.1,
		TokensPerSec:      68.0,
		MonthlySavingsUSD: 3430.0,
	}

	md := GenerateMarkdownReport(job, bm, cfg)
	if !strings.Contains(md, "Executive Distillation and FinOps Audit Report") {
		t.Fatal("expected markdown report title")
	}
	if !strings.Contains(md, "$41160") {
		t.Errorf("expected annual savings $41160 in report, got:\n%s", md)
	}

	html := GenerateHTMLReport(job, bm, cfg)
	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Fatal("expected valid html document")
	}
}
