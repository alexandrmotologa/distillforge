package store

import (
	"time"
)

type Project struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Job struct {
	ID           string     `json:"id"`
	ProjectID    string     `json:"project_id"`
	Type         string     `json:"type"` // generate, audit, train, evaluate, export, pipeline
	Status       string     `json:"status"` // pending, running, completed, failed
	Progress     float64    `json:"progress"` // 0.0 - 100.0
	ErrorMessage string     `json:"error_message,omitempty"`
	ConfigJSON   string     `json:"config_json,omitempty"`
	StartedAt    *time.Time `json:"started_at,omitempty"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

type Dataset struct {
	ID          string    `json:"id"`
	ProjectID   string    `json:"project_id"`
	Name        string    `json:"name"`
	FilePath    string    `json:"file_path"`
	SampleCount int       `json:"sample_count"`
	Format      string    `json:"format"` // jsonl, alpaca, sharegpt
	Status      string    `json:"status"` // raw, clean, exported
	CreatedAt   time.Time `json:"created_at"`
}

type Sample struct {
	ID           string    `json:"id"`
	DatasetID    string    `json:"dataset_id"`
	Prompt       string    `json:"prompt"`
	Response     string    `json:"response"`
	Strategy     string    `json:"strategy"` // seed, in_depth, in_breadth, adversarial, cot
	JudgeScore   float64   `json:"judge_score"`
	IsAccepted   bool      `json:"is_accepted"`
	MetadataJSON string    `json:"metadata_json,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

type TrainingMetric struct {
	ID           int64     `json:"id"`
	JobID        string    `json:"job_id"`
	Step         int       `json:"step"`
	Epoch        float64   `json:"epoch"`
	Loss         float64   `json:"loss"`
	LearningRate float64   `json:"learning_rate"`
	Perplexity   float64   `json:"perplexity"`
	VRAMMB       float64   `json:"vram_mb"`
	CreatedAt    time.Time `json:"created_at"`
}

type Benchmark struct {
	ID                 string    `json:"id"`
	JobID              string    `json:"job_id"`
	StudentModel       string    `json:"student_model"`
	TeacherModel       string    `json:"teacher_model"`
	ExactMatch         float64   `json:"exact_match"`
	SchemaAdherence    float64   `json:"schema_adherence"`
	AvgLatencyMs       float64   `json:"avg_latency_ms"`
	P95LatencyMs       float64   `json:"p95_latency_ms"`
	TokensPerSec       float64   `json:"tokens_per_sec"`
	MonthlySavingsUSD  float64   `json:"monthly_savings_usd"`
	EvaluationSummary  string    `json:"evaluation_summary,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
}
