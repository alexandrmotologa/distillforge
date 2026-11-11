package server

import (
	"bytes"
	"encoding/json"
	"fmt"
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

func TestServerUploadAndCurationEndpoints(t *testing.T) {
	srv, st, _ := setupTestServer(t)
	defer st.Close()

	handler := srv.Routes()

	// 1. Upload test
	uploadJSON := `{"filename": "test_upload.jsonl", "content": "{\"prompt\":\"Test\",\"response\":\"Answer\"}\n"}`
	req := httptest.NewRequest("POST", "/api/upload", bytes.NewBufferString(uploadJSON))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for upload, got %d (body: %s)", rec.Code, rec.Body.String())
	}

	var uploadResp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &uploadResp); err != nil {
		t.Fatalf("failed to parse upload response: %v", err)
	}
	if uploadResp["status"] != "success" {
		t.Fatalf("expected upload success, got %v", uploadResp["status"])
	}

	// 2. Create Dataset & Sample
	p, _ := st.SaveProject("curation-project", "testing curation")
	ds, _ := st.SaveDataset(p.ID, "curated", "data.jsonl", "jsonl", "clean", 1)

	sampleJSON := fmt.Sprintf(`{"dataset_id":"%s","prompt":"Original prompt","response":"Original response","strategy":"golden"}`, ds.ID)
	req = httptest.NewRequest("POST", "/api/samples", bytes.NewBufferString(sampleJSON))
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for sample, got %d", rec.Code)
	}

	var created store.Sample
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("failed to unmarshal sample: %v", err)
	}

	// 3. Update Sample
	updateJSON := `{"prompt":"Modified prompt","response":"Modified response","is_accepted":true,"judge_score":0.99}`
	req = httptest.NewRequest("PUT", "/api/samples/"+created.ID, bytes.NewBufferString(updateJSON))
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for update, got %d", rec.Code)
	}

	// 4. Re-judge Sample
	req = httptest.NewRequest("POST", "/api/samples/"+created.ID+"/re-judge", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for re-judge, got %d", rec.Code)
	}

	// 5. Bulk action
	bulkJSON := fmt.Sprintf(`{"ids":["%s"],"action":"reject"}`, created.ID)
	req = httptest.NewRequest("POST", "/api/samples/bulk", bytes.NewBufferString(bulkJSON))
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for bulk, got %d", rec.Code)
	}

	// 6. Delete Sample
	req = httptest.NewRequest("DELETE", "/api/samples/"+created.ID, nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for delete, got %d", rec.Code)
	}

	// 7. OpenAI-compatible /v1/chat/completions
	openAIReq := `{"model":"test-model","messages":[{"role":"user","content":"Classify ticket 99"}]}`
	req = httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewBufferString(openAIReq))
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for openai completions, got %d", rec.Code)
	}

	var chatResp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &chatResp); err != nil {
		t.Fatalf("failed to parse openai chat response: %v", err)
	}
	if chatResp["object"] != "chat.completion" {
		t.Errorf("expected object 'chat.completion', got %v", chatResp["object"])
	}

	// 8. OpenAI-compatible /v1/models
	req = httptest.NewRequest("GET", "/v1/models", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for openai models, got %d", rec.Code)
	}

	// 9. Dataset Analytics Endpoint
	req = httptest.NewRequest("GET", "/api/datasets/"+ds.ID+"/analytics", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for dataset analytics, got %d", rec.Code)
	}

	// 10. Advisor Recommend Endpoint
	req = httptest.NewRequest("GET", "/api/advisor/recommend?hardware=gpu_8gb&dataset_size=1000", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for advisor, got %d", rec.Code)
	}
}
