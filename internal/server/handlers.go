package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/alexandrmotologa/distillforge/internal/benchmark"
	"github.com/alexandrmotologa/distillforge/internal/config"
	"github.com/alexandrmotologa/distillforge/internal/exporter"
	"github.com/alexandrmotologa/distillforge/internal/filter"
	"github.com/alexandrmotologa/distillforge/internal/generator"
	"github.com/alexandrmotologa/distillforge/internal/store"
	"github.com/alexandrmotologa/distillforge/internal/trainer"
)

type Server struct {
	cfg         *config.Config
	store       *store.Store
	broadcaster *SSEBroadcaster
	staticFS    http.FileSystem
}

func NewServer(cfg *config.Config, s *store.Store, staticFS http.FileSystem) *Server {
	return &Server{
		cfg:         cfg,
		store:       s,
		broadcaster: NewSSEBroadcaster(),
		staticFS:    staticFS,
	}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	// API Routes
	mux.HandleFunc("GET /api/health", s.handleHealth)
	mux.HandleFunc("GET /api/config", s.handleGetConfig)
	mux.HandleFunc("GET /api/projects", s.handleListProjects)
	mux.HandleFunc("POST /api/projects", s.handleCreateProject)
	mux.HandleFunc("GET /api/jobs", s.handleListJobs)
	mux.HandleFunc("GET /api/jobs/{id}", s.handleGetJob)
	mux.HandleFunc("GET /api/jobs/{id}/metrics", s.handleGetJobMetrics)
	mux.HandleFunc("GET /api/jobs/{id}/events", s.handleJobEvents)
	mux.HandleFunc("POST /api/jobs/run", s.handleRunPipeline)
	mux.HandleFunc("GET /api/datasets", s.handleListDatasets)
	mux.HandleFunc("GET /api/datasets/{id}/samples", s.handleListSamples)
	mux.HandleFunc("GET /api/benchmarks/{jobId}", s.handleGetBenchmark)
	mux.HandleFunc("POST /api/playground/chat", s.handlePlaygroundChat)

	// Static assets handler
	if s.staticFS != nil {
		fileServer := http.FileServer(s.staticFS)
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api/") {
				http.NotFound(w, r)
				return
			}
			fileServer.ServeHTTP(w, r)
		})
	} else {
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api/") {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(`<!DOCTYPE html><html><head><title>distillforge Studio</title></head><body><h1>distillforge Web Studio API Active</h1><p>Embedded studio assets will be loaded here.</p></body></html>`))
		})
	}

	return corsMiddleware(mux)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":    "healthy",
		"service":   "distillforge",
		"version":   "1.0.0",
		"timestamp": time.Now().UTC(),
	})
}

func (s *Server) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.cfg)
}

func (s *Server) handleListProjects(w http.ResponseWriter, r *http.Request) {
	projects, err := s.store.ListProjects()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, projects)
}

func (s *Server) handleCreateProject(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	p, err := s.store.SaveProject(req.Name, req.Description)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, p)
}

func (s *Server) handleListJobs(w http.ResponseWriter, r *http.Request) {
	jobs, err := s.store.ListJobs(100)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, jobs)
}

func (s *Server) handleGetJob(w http.ResponseWriter, r *http.Request) {
	jobID := r.PathValue("id")
	job, err := s.store.GetJob(jobID)
	if err != nil {
		http.Error(w, "job not found", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, job)
}

func (s *Server) handleGetJobMetrics(w http.ResponseWriter, r *http.Request) {
	jobID := r.PathValue("id")
	metrics, err := s.store.GetMetricsByJob(jobID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, metrics)
}

func (s *Server) handleJobEvents(w http.ResponseWriter, r *http.Request) {
	jobID := r.PathValue("id")
	s.broadcaster.HandleSSE(w, r, jobID)
}

func (s *Server) handleListDatasets(w http.ResponseWriter, r *http.Request) {
	datasets, err := s.store.ListDatasets()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, datasets)
}

func (s *Server) handleListSamples(w http.ResponseWriter, r *http.Request) {
	datasetID := r.PathValue("id")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	samples, err := s.store.ListSamples(datasetID, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, samples)
}

func (s *Server) handleGetBenchmark(w http.ResponseWriter, r *http.Request) {
	jobID := r.PathValue("jobId")
	bm, err := s.store.GetBenchmarkByJob(jobID)
	if err != nil {
		http.Error(w, "benchmark not found", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, bm)
}

func (s *Server) handlePlaygroundChat(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Prompt string `json:"prompt"`
		Model  string `json:"model"` // teacher or student
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	start := time.Now()
	var client generator.TeacherClient
	var err error

	if req.Model == "student" {
		client = &generator.MockTeacherClient{ModelName: s.cfg.Student.BaseModel}
	} else {
		client, err = generator.NewTeacherClient(&s.cfg.Teacher)
		if err != nil {
			client = &generator.MockTeacherClient{ModelName: s.cfg.Teacher.Model}
		}
	}

	systemPrompt := "You are a domain assistant executing this task with precision."
	resp, err := client.Generate(r.Context(), systemPrompt, req.Prompt)
	durMs := float64(time.Since(start).Microseconds()) / 1000.0

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"response":   resp,
		"latency_ms": durMs,
		"model":      req.Model,
	})
}

func (s *Server) handleRunPipeline(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ProjectID string         `json:"project_id"`
		Config    *config.Config `json:"config"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	cfg := req.Config
	if cfg == nil {
		cfg = s.cfg
	}

	projectID := req.ProjectID
	if projectID == "" {
		projects, _ := s.store.ListProjects()
		if len(projects) > 0 {
			projectID = projects[0].ID
		} else {
			p, _ := s.store.SaveProject(cfg.Project.Name, cfg.Project.Description)
			projectID = p.ID
		}
	}

	cfgBytes, _ := json.Marshal(cfg)
	job, err := s.store.CreateJob(projectID, "pipeline", string(cfgBytes))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Launch async pipeline in background
	go s.runAsyncPipeline(job.ID, projectID, cfg)

	writeJSON(w, http.StatusAccepted, job)
}

func (s *Server) runAsyncPipeline(jobID, projectID string, cfg *config.Config) {
	ctx := context.Background()
	s.broadcaster.Broadcast(jobID, "status", map[string]any{"status": "running", "progress": 5.0, "stage": "generation"})
	_ = s.store.UpdateJobProgress(jobID, 5.0, "running", "")

	// 1. Generation
	gen, err := generator.NewPipelineGenerator(cfg, s.store)
	if err != nil {
		s.failJob(jobID, "generator init failed: "+err.Error())
		return
	}

	rawPath := fmt.Sprintf("./data/%s_raw.jsonl", jobID)
	cleanPath := fmt.Sprintf("./data/%s_clean.jsonl", jobID)
	checkpointsDir := fmt.Sprintf("./checkpoints/%s", jobID)
	distDir := fmt.Sprintf("./dist/%s", jobID)

	_, err = gen.GenerateDataset(ctx, jobID, rawPath, func(done, total int, pct float64) {
		s.broadcaster.Broadcast(jobID, "progress", map[string]any{
			"stage":    "generation",
			"done":     done,
			"total":    total,
			"progress": 5.0 + (pct * 0.25),
		})
	})
	if err != nil {
		s.failJob(jobID, "dataset generation failed: "+err.Error())
		return
	}

	// 2. Audit & Filter
	s.broadcaster.Broadcast(jobID, "status", map[string]any{"status": "running", "progress": 30.0, "stage": "filtering"})
	_ = s.store.UpdateJobProgress(jobID, 30.0, "running", "")

	filt, err := filter.NewPipelineFilter(cfg)
	if err != nil {
		s.failJob(jobID, "filter init failed: "+err.Error())
		return
	}

	report, cleanSamples, err := filt.AuditDataset(ctx, rawPath, cleanPath)
	if err != nil {
		s.failJob(jobID, "audit failed: "+err.Error())
		return
	}

	dataset, _ := s.store.SaveDataset(projectID, cfg.Project.Name+"-clean", cleanPath, "jsonl", "clean", len(cleanSamples))
	if dataset != nil && len(cleanSamples) > 0 {
		for i := range cleanSamples {
			cleanSamples[i].DatasetID = dataset.ID
		}
		_ = s.store.InsertSampleBatch(cleanSamples)
	}

	s.broadcaster.Broadcast(jobID, "audit_completed", report)

	// 3. Fine-Tuning
	s.broadcaster.Broadcast(jobID, "status", map[string]any{"status": "running", "progress": 50.0, "stage": "training"})
	_ = s.store.UpdateJobProgress(jobID, 50.0, "running", "")

	trainerEngine := trainer.NewPipelineTrainer(cfg, s.store)
	err = trainerEngine.Train(ctx, jobID, cleanPath, checkpointsDir, cfg.Teacher.Provider == "mock",
		func(metric *store.TrainingMetric) {
			s.broadcaster.Broadcast(jobID, "training_metric", metric)
		},
		func(logLine string) {
			s.broadcaster.Broadcast(jobID, "log", logLine)
		},
	)
	if err != nil {
		s.failJob(jobID, "training failed: "+err.Error())
		return
	}

	// 4. Benchmark Holdout
	s.broadcaster.Broadcast(jobID, "status", map[string]any{"status": "running", "progress": 80.0, "stage": "benchmarking"})
	_ = s.store.UpdateJobProgress(jobID, 80.0, "running", "")

	bm, err := benchmark.NewPipelineBenchmark(cfg, s.store)
	if err == nil && len(cleanSamples) > 0 {
		summary, _ := bm.EvaluateHoldout(ctx, jobID, cleanSamples)
		if summary != nil {
			s.broadcaster.Broadcast(jobID, "benchmark_summary", summary)
		}
	}

	// 5. Quantize & Export
	s.broadcaster.Broadcast(jobID, "status", map[string]any{"status": "running", "progress": 95.0, "stage": "export"})
	_ = s.store.UpdateJobProgress(jobID, 95.0, "running", "")

	exp := exporter.NewPipelineExporter(cfg)
	ggufPath, mfPath, err := exp.ExportModel(ctx, checkpointsDir, distDir, cfg.Teacher.Provider == "mock")
	if err != nil {
		s.failJob(jobID, "model export failed: "+err.Error())
		return
	}

	s.broadcaster.Broadcast(jobID, "completed", map[string]any{
		"status":        "completed",
		"progress":      100.0,
		"gguf_path":     ggufPath,
		"modelfile":     mfPath,
		"report":        report,
	})
	_ = s.store.UpdateJobProgress(jobID, 100.0, "completed", "")
}

func (s *Server) failJob(jobID, errMsg string) {
	_ = s.store.UpdateJobProgress(jobID, 0.0, "failed", errMsg)
	s.broadcaster.Broadcast(jobID, "error", map[string]any{
		"status":        "failed",
		"error_message": errMsg,
	})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
