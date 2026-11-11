package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

// SSEBroadcaster manages active SSE client connections per job.
type SSEBroadcaster struct {
	mu      sync.RWMutex
	clients map[string]map[chan []byte]struct{}
}

// NewSSEBroadcaster creates a broadcaster instance.
func NewSSEBroadcaster() *SSEBroadcaster {
	return &SSEBroadcaster{
		clients: make(map[string]map[chan []byte]struct{}),
	}
}

// Subscribe registers a new HTTP client channel for a job ID.
func (b *SSEBroadcaster) Subscribe(jobID string) chan []byte {
	b.mu.Lock()
	defer b.mu.Unlock()

	ch := make(chan []byte, 32)
	if _, ok := b.clients[jobID]; !ok {
		b.clients[jobID] = make(map[chan []byte]struct{})
	}
	b.clients[jobID][ch] = struct{}{}
	return ch
}

// Unsubscribe removes a client channel.
func (b *SSEBroadcaster) Unsubscribe(jobID string, ch chan []byte) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if subs, ok := b.clients[jobID]; ok {
		delete(subs, ch)
		close(ch)
		if len(subs) == 0 {
			delete(b.clients, jobID)
		}
	}
}

// Broadcast sends an event payload to all subscribers for a given job.
func (b *SSEBroadcaster) Broadcast(jobID, eventType string, payload any) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	subs, ok := b.clients[jobID]
	if !ok || len(subs) == 0 {
		return
	}

	dataBytes, err := json.Marshal(payload)
	if err != nil {
		return
	}

	msg := fmt.Sprintf("event: %s\ndata: %s\n\n", eventType, string(dataBytes))
	msgBytes := []byte(msg)

	for ch := range subs {
		select {
		case ch <- msgBytes:
		default:
		}
	}
}

// HandleSSE streams events to an HTTP client.
func (b *SSEBroadcaster) HandleSSE(w http.ResponseWriter, r *http.Request, jobID string) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	ch := b.Subscribe(jobID)
	defer b.Unsubscribe(jobID, ch)

	// Send initial ping
	fmt.Fprintf(w, "event: connected\ndata: {\"job_id\": %q}\n\n", jobID)
	flusher.Flush()

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			_, _ = w.Write(msg)
			flusher.Flush()
		}
	}
}
