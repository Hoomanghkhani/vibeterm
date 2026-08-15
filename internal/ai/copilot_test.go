package ai

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBuildSystemPrompt(t *testing.T) {
	prompt := buildSystemPrompt("Linux Fedora 44", "error: permission denied")
	if !strings.Contains(prompt, "Linux Fedora 44") {
		t.Errorf("expected OS info in system prompt, got %s", prompt)
	}
	if !strings.Contains(prompt, "permission denied") {
		t.Errorf("expected terminal context in system prompt, got %s", prompt)
	}
}

func TestStreamOllama(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/generate" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"response":"sudo "}`+"\n")
		fmt.Fprintf(w, `{"response":"systemctl restart nginx","done":true}`+"\n")
	}))
	defer ts.Close()

	svc := NewCopilotService()
	var chunks []string
	err := svc.StreamCompletion(context.Background(), PromptRequest{
		Provider: "ollama",
		BaseURL:  ts.URL,
		Prompt:   "restart nginx",
	}, func(c string) {
		chunks = append(chunks, c)
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := strings.Join(chunks, "")
	if result != "sudo systemctl restart nginx" {
		t.Errorf("expected 'sudo systemctl restart nginx', got %q", result)
	}
}

func TestStreamOpenAI(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-key" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprintf(w, "data: {\"choices\":[{\"delta\":{\"content\":\"ls -la\"}}]}\n\n")
		fmt.Fprintf(w, "data: [DONE]\n\n")
	}))
	defer ts.Close()

	svc := NewCopilotService()
	var chunks []string
	err := svc.StreamCompletion(context.Background(), PromptRequest{
		Provider: "openai",
		BaseURL:  ts.URL,
		APIKey:   "test-key",
		Prompt:   "list files",
	}, func(c string) {
		chunks = append(chunks, c)
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := strings.Join(chunks, "")
	if result != "ls -la" {
		t.Errorf("expected 'ls -la', got %q", result)
	}
}
