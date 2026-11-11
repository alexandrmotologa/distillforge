package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/alexandrmotologa/distillforge/internal/config"
	"github.com/alexandrmotologa/distillforge/internal/store"
)

func setupTestServer(t *testing.T) (*Server, *store.Store, *config.Config) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "server_test.db")

	st, err := store.NewStore(dbPath)
	if err != nil {
		t.Fatalf("failed to init store: %v", err)
	}

	cfg := &config.Config{
		Project: config.ProjectConfig{Name: "test-server", Description: "Test server suite"},
		Teacher: config.TeacherConfig{Provider: "mock", Model: "gpt-4o"},
		Student: config.StudentConfig{BaseModel: "Llama-3.2-1B-Instruct"},
	}
	cfg.SetDefaults()

	srv := NewServer(cfg, st, nil)
	return srv, st, cfg
}

func TestServerHealthAndConfig(t *testing.T) {
	srv, st, _ := setupTestServer(t)
	defer st.Close()

	handler := srv.Routes()

	// 1. Health check
	req := httptest.NewRequest("GET", "/api/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rec.Code)
	}

	var health map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &health); err != nil {
		t.Fatalf("failed to parse health response: %v", err)
	}
	if health["status"] != "healthy" {
		t.Errorf("expected healthy status, got %v", health["status"])
	}

	// 2. Config check
	req = httptest.NewRequest("GET", "/api/config", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for config, got %d", rec.Code)
	}
}

func TestServerProjectAndJobFlow(t *testing.T) {
	srv, st, _ := setupTestServer(t)
	defer st.Close()

	handler := srv.Routes()

	// Create project
	body := bytes.NewBufferString(`{"name": "support-tickets", "description": "Customer ticket classifier"}`)
	req := httptest.NewRequest("POST", "/api/projects", body)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d", rec.Code)
	}

	var p store.Project
	if err := json.Unmarshal(rec.Body.Bytes(), &p); err != nil {
		t.Fatalf("failed to parse project: %v", err)
	}
	if p.ID == "" || p.Name != "support-tickets" {
		t.Fatalf("unexpected project created: %+v", p)
	}

	// List projects
	req = httptest.NewRequest("GET", "/api/projects", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	var list []store.Project
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil || len(list) != 1 {
		t.Fatalf("expected 1 project in list, got %d (err: %v)", len(list), err)
	}
}

func TestServerPlaygroundChat(t *testing.T) {
	srv, st, _ := setupTestServer(t)
	defer st.Close()

	handler := srv.Routes()

	body := bytes.NewBufferString(`{"prompt": "Internet is broken after bad weather", "model": "student"}`)
	req := httptest.NewRequest("POST", "/api/playground/chat", body)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from playground, got %d", rec.Code)
	}

	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse playground response: %v", err)
	}
	if _, ok := resp["response"]; !ok {
		t.Fatal("expected 'response' field in playground output")
	}
}
