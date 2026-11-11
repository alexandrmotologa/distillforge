package store

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
	mu sync.RWMutex
}

func NewStore(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	// Optimize SQLite performance for single-process concurrency
	if _, err := db.Exec(`
		PRAGMA journal_mode = WAL;
		PRAGMA synchronous = NORMAL;
		PRAGMA foreign_keys = ON;
	`); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to configure sqlite pragmas: %w", err)
	}

	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run database migrations: %w", err)
	}

	return s, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS projects (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		description TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);

	CREATE TABLE IF NOT EXISTS jobs (
		id TEXT PRIMARY KEY,
		project_id TEXT NOT NULL,
		type TEXT NOT NULL,
		status TEXT NOT NULL,
		progress REAL NOT NULL DEFAULT 0.0,
		error_message TEXT DEFAULT '',
		config_json TEXT DEFAULT '',
		started_at DATETIME,
		completed_at DATETIME,
		created_at DATETIME NOT NULL
	);

	CREATE TABLE IF NOT EXISTS datasets (
		id TEXT PRIMARY KEY,
		project_id TEXT NOT NULL,
		name TEXT NOT NULL,
		file_path TEXT NOT NULL,
		sample_count INTEGER NOT NULL DEFAULT 0,
		format TEXT NOT NULL DEFAULT 'jsonl',
		status TEXT NOT NULL DEFAULT 'raw',
		created_at DATETIME NOT NULL
	);

	CREATE TABLE IF NOT EXISTS samples (
		id TEXT PRIMARY KEY,
		dataset_id TEXT NOT NULL,
		prompt TEXT NOT NULL,
		response TEXT NOT NULL,
		strategy TEXT NOT NULL,
		judge_score REAL NOT NULL DEFAULT 0.0,
		is_accepted INTEGER NOT NULL DEFAULT 1,
		metadata_json TEXT DEFAULT '',
		created_at DATETIME NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_samples_dataset ON samples(dataset_id);
	CREATE INDEX IF NOT EXISTS idx_samples_accepted ON samples(dataset_id, is_accepted);

	CREATE TABLE IF NOT EXISTS training_metrics (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		job_id TEXT NOT NULL,
		step INTEGER NOT NULL,
		epoch REAL NOT NULL,
		loss REAL NOT NULL,
		learning_rate REAL NOT NULL,
		perplexity REAL NOT NULL,
		vram_mb REAL NOT NULL,
		created_at DATETIME NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_metrics_job ON training_metrics(job_id, step);

	CREATE TABLE IF NOT EXISTS benchmarks (
		id TEXT PRIMARY KEY,
		job_id TEXT NOT NULL,
		student_model TEXT NOT NULL,
		teacher_model TEXT NOT NULL,
		exact_match REAL NOT NULL,
		schema_adherence REAL NOT NULL,
		avg_latency_ms REAL NOT NULL,
		p95_latency_ms REAL NOT NULL,
		tokens_per_sec REAL NOT NULL,
		monthly_savings_usd REAL NOT NULL,
		evaluation_summary TEXT DEFAULT '',
		created_at DATETIME NOT NULL
	);
	`

	_, err := s.db.Exec(schema)
	return err
}

func (s *Store) SaveProject(name, description string) (*Project, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	p := &Project{
		ID:          uuid.New().String(),
		Name:        name,
		Description: description,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	query := `INSERT INTO projects (id, name, description, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`
	_, err := s.db.Exec(query, p.ID, p.Name, p.Description, p.CreatedAt, p.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to save project: %w", err)
	}

	return p, nil
}

func (s *Store) ListProjects() ([]Project, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query(`SELECT id, name, description, created_at, updated_at FROM projects ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var p Project
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, nil
}

func (s *Store) CreateJob(projectID, jobType, configJSON string) (*Job, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	job := &Job{
		ID:          uuid.New().String(),
		ProjectID:   projectID,
		Type:        jobType,
		Status:      "pending",
		Progress:    0.0,
		ConfigJSON:  configJSON,
		StartedAt:   &now,
		CreatedAt:   now,
	}

	query := `INSERT INTO jobs (id, project_id, type, status, progress, config_json, started_at, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := s.db.Exec(query, job.ID, job.ProjectID, job.Type, job.Status, job.Progress, job.ConfigJSON, job.StartedAt, job.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create job: %w", err)
	}

	return job, nil
}

func (s *Store) UpdateJobProgress(id string, progress float64, status, errorMessage string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var completedAt *time.Time
	if status == "completed" || status == "failed" {
		now := time.Now().UTC()
		completedAt = &now
	}

	query := `UPDATE jobs SET progress = ?, status = ?, error_message = ?, completed_at = ? WHERE id = ?`
	_, err := s.db.Exec(query, progress, status, errorMessage, completedAt, id)
	return err
}

func (s *Store) GetJob(id string) (*Job, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	job := &Job{}
	query := `SELECT id, project_id, type, status, progress, error_message, config_json, started_at, completed_at, created_at FROM jobs WHERE id = ?`
	err := s.db.QueryRow(query, id).Scan(
		&job.ID, &job.ProjectID, &job.Type, &job.Status, &job.Progress,
		&job.ErrorMessage, &job.ConfigJSON, &job.StartedAt, &job.CompletedAt, &job.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return job, nil
}

func (s *Store) ListJobs(limit int) ([]Job, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 {
		limit = 50
	}

	rows, err := s.db.Query(`SELECT id, project_id, type, status, progress, error_message, config_json, started_at, completed_at, created_at FROM jobs ORDER BY created_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []Job
	for rows.Next() {
		var j Job
		if err := rows.Scan(&j.ID, &j.ProjectID, &j.Type, &j.Status, &j.Progress, &j.ErrorMessage, &j.ConfigJSON, &j.StartedAt, &j.CompletedAt, &j.CreatedAt); err != nil {
			return nil, err
		}
		jobs = append(jobs, j)
	}
	return jobs, nil
}

func (s *Store) SaveDataset(projectID, name, filePath, format, status string, count int) (*Dataset, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	d := &Dataset{
		ID:          uuid.New().String(),
		ProjectID:   projectID,
		Name:        name,
		FilePath:    filePath,
		SampleCount: count,
		Format:      format,
		Status:      status,
		CreatedAt:   time.Now().UTC(),
	}

	query := `INSERT INTO datasets (id, project_id, name, file_path, sample_count, format, status, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := s.db.Exec(query, d.ID, d.ProjectID, d.Name, d.FilePath, d.SampleCount, d.Format, d.Status, d.CreatedAt)
	if err != nil {
		return nil, err
	}

	return d, nil
}

func (s *Store) ListDatasets() ([]Dataset, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query(`SELECT id, project_id, name, file_path, sample_count, format, status, created_at FROM datasets ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var datasets []Dataset
	for rows.Next() {
		var d Dataset
		if err := rows.Scan(&d.ID, &d.ProjectID, &d.Name, &d.FilePath, &d.SampleCount, &d.Format, &d.Status, &d.CreatedAt); err != nil {
			return nil, err
		}
		datasets = append(datasets, d)
	}
	return datasets, nil
}

func (s *Store) InsertSampleBatch(samples []Sample) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`INSERT INTO samples (id, dataset_id, prompt, response, strategy, judge_score, is_accepted, metadata_json, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, sm := range samples {
		id := sm.ID
		if id == "" {
			id = uuid.New().String()
		}
		createdAt := sm.CreatedAt
		if createdAt.IsZero() {
			createdAt = time.Now().UTC()
		}
		acceptedInt := 0
		if sm.IsAccepted {
			acceptedInt = 1
		}

		if _, err := stmt.Exec(id, sm.DatasetID, sm.Prompt, sm.Response, sm.Strategy, sm.JudgeScore, acceptedInt, sm.MetadataJSON, createdAt); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *Store) InsertSample(sm *Sample) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if sm.ID == "" {
		sm.ID = uuid.New().String()
	}
	if sm.CreatedAt.IsZero() {
		sm.CreatedAt = time.Now().UTC()
	}
	acceptedInt := 0
	if sm.IsAccepted {
		acceptedInt = 1
	}

	query := `INSERT INTO samples (id, dataset_id, prompt, response, strategy, judge_score, is_accepted, metadata_json, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := s.db.Exec(query, sm.ID, sm.DatasetID, sm.Prompt, sm.Response, sm.Strategy, sm.JudgeScore, acceptedInt, sm.MetadataJSON, sm.CreatedAt)
	return err
}

func (s *Store) GetSample(id string) (*Sample, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sm := &Sample{}
	var acceptedInt int
	query := `SELECT id, dataset_id, prompt, response, strategy, judge_score, is_accepted, metadata_json, created_at FROM samples WHERE id = ?`
	err := s.db.QueryRow(query, id).Scan(
		&sm.ID, &sm.DatasetID, &sm.Prompt, &sm.Response, &sm.Strategy,
		&sm.JudgeScore, &acceptedInt, &sm.MetadataJSON, &sm.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	sm.IsAccepted = (acceptedInt == 1)
	return sm, nil
}

func (s *Store) UpdateSample(id, prompt, response string, isAccepted bool, judgeScore float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	acceptedInt := 0
	if isAccepted {
		acceptedInt = 1
	}

	query := `UPDATE samples SET prompt = ?, response = ?, is_accepted = ?, judge_score = ? WHERE id = ?`
	_, err := s.db.Exec(query, prompt, response, acceptedInt, judgeScore, id)
	return err
}

func (s *Store) DeleteSample(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec(`DELETE FROM samples WHERE id = ?`, id)
	return err
}

func (s *Store) BulkUpdateSamplesStatus(ids []string, isAccepted bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(ids) == 0 {
		return nil
	}

	acceptedInt := 0
	if isAccepted {
		acceptedInt = 1
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`UPDATE samples SET is_accepted = ? WHERE id = ?`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, id := range ids {
		if _, err := stmt.Exec(acceptedInt, id); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *Store) BulkDeleteSamples(ids []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(ids) == 0 {
		return nil
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`DELETE FROM samples WHERE id = ?`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, id := range ids {
		if _, err := stmt.Exec(id); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *Store) ListSamples(datasetID string, limit, offset int) ([]Sample, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 {
		limit = 100
	}

	rows, err := s.db.Query(`SELECT id, dataset_id, prompt, response, strategy, judge_score, is_accepted, metadata_json, created_at FROM samples WHERE dataset_id = ? ORDER BY created_at ASC LIMIT ? OFFSET ?`, datasetID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var samples []Sample
	for rows.Next() {
		var sm Sample
		var acceptedInt int
		if err := rows.Scan(&sm.ID, &sm.DatasetID, &sm.Prompt, &sm.Response, &sm.Strategy, &sm.JudgeScore, &acceptedInt, &sm.MetadataJSON, &sm.CreatedAt); err != nil {
			return nil, err
		}
		sm.IsAccepted = (acceptedInt == 1)
		samples = append(samples, sm)
	}
	return samples, nil
}

func (s *Store) RecordMetric(m *TrainingMetric) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now().UTC()
	}

	query := `INSERT INTO training_metrics (job_id, step, epoch, loss, learning_rate, perplexity, vram_mb, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	res, err := s.db.Exec(query, m.JobID, m.Step, m.Epoch, m.Loss, m.LearningRate, m.Perplexity, m.VRAMMB, m.CreatedAt)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err == nil {
		m.ID = id
	}
	return nil
}

func (s *Store) GetMetricsByJob(jobID string) ([]TrainingMetric, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query(`SELECT id, job_id, step, epoch, loss, learning_rate, perplexity, vram_mb, created_at FROM training_metrics WHERE job_id = ? ORDER BY step ASC`, jobID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []TrainingMetric
	for rows.Next() {
		var m TrainingMetric
		if err := rows.Scan(&m.ID, &m.JobID, &m.Step, &m.Epoch, &m.Loss, &m.LearningRate, &m.Perplexity, &m.VRAMMB, &m.CreatedAt); err != nil {
			return nil, err
		}
		metrics = append(metrics, m)
	}
	return metrics, nil
}

func (s *Store) SaveBenchmark(b *Benchmark) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if b.ID == "" {
		b.ID = uuid.New().String()
	}
	if b.CreatedAt.IsZero() {
		b.CreatedAt = time.Now().UTC()
	}

	query := `INSERT INTO benchmarks (id, job_id, student_model, teacher_model, exact_match, schema_adherence, avg_latency_ms, p95_latency_ms, tokens_per_sec, monthly_savings_usd, evaluation_summary, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := s.db.Exec(query, b.ID, b.JobID, b.StudentModel, b.TeacherModel, b.ExactMatch, b.SchemaAdherence, b.AvgLatencyMs, b.P95LatencyMs, b.TokensPerSec, b.MonthlySavingsUSD, b.EvaluationSummary, b.CreatedAt)
	return err
}

func (s *Store) GetBenchmarkByJob(jobID string) (*Benchmark, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	b := &Benchmark{}
	query := `SELECT id, job_id, student_model, teacher_model, exact_match, schema_adherence, avg_latency_ms, p95_latency_ms, tokens_per_sec, monthly_savings_usd, evaluation_summary, created_at FROM benchmarks WHERE job_id = ? ORDER BY created_at DESC LIMIT 1`
	err := s.db.QueryRow(query, jobID).Scan(
		&b.ID, &b.JobID, &b.StudentModel, &b.TeacherModel, &b.ExactMatch,
		&b.SchemaAdherence, &b.AvgLatencyMs, &b.P95LatencyMs, &b.TokensPerSec,
		&b.MonthlySavingsUSD, &b.EvaluationSummary, &b.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return b, nil
}
