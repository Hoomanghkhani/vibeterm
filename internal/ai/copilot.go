package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// CopilotService provides streaming natural language to shell command generation
type CopilotService struct {
	httpClient *http.Client
}

// NewCopilotService creates a new AI Copilot Service
func NewCopilotService() *CopilotService {
	return &CopilotService{
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// PromptRequest defines the payload for an AI completion
type PromptRequest struct {
	Prompt          string `json:"prompt"`
	TerminalContext string `json:"terminalContext,omitempty"`
	OSInfo          string `json:"osInfo,omitempty"`
	Provider        string `json:"provider"` // "ollama", "openai", "anthropic", "gemini"
	APIKey          string `json:"apiKey,omitempty"`
	BaseURL         string `json:"baseUrl,omitempty"`
	Model           string `json:"model,omitempty"`
}

// StreamCompletion streams chunks of generated text/commands to the chunkCallback
func (s *CopilotService) StreamCompletion(ctx context.Context, req PromptRequest, chunkCallback func(chunk string)) error {
	switch req.Provider {
	case "ollama":
		return s.streamOllama(ctx, req, chunkCallback)
	case "openai":
		return s.streamOpenAI(ctx, req, chunkCallback)
	case "anthropic":
		return s.streamAnthropic(ctx, req, chunkCallback)
	case "gemini":
		return s.streamGemini(ctx, req, chunkCallback)
	default:
		// Fallback to Ollama
		return s.streamOllama(ctx, req, chunkCallback)
	}
}

func buildSystemPrompt(osInfo, terminalContext string) string {
	var sb strings.Builder
	sb.WriteString("You are VibeTerm AI Copilot, a principal infrastructure and CLI engineering assistant. ")
	sb.WriteString("Generate concise, safe, and accurate shell commands or explanations. ")
	sb.WriteString("If the user asks for a command, return the executable CLI command on the first line, followed by a brief 1-line explanation if necessary. ")
	if osInfo != "" {
		sb.WriteString(fmt.Sprintf("Target OS/Shell: %s. ", osInfo))
	}
	if terminalContext != "" {
		sb.WriteString(fmt.Sprintf("\nRecent Terminal Output Context:\n```\n%s\n```\n", terminalContext))
	}
	return sb.String()
}

func (s *CopilotService) streamOllama(ctx context.Context, req PromptRequest, chunkCallback func(string)) error {
	baseURL := req.BaseURL
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	model := req.Model
	if model == "" {
		model = "llama3:latest"
	}

	payload := map[string]interface{}{
		"model":  model,
		"prompt": req.Prompt,
		"system": buildSystemPrompt(req.OSInfo, req.TerminalContext),
		"stream": true,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/api/generate", bytes.NewReader(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("ollama connection error: %w (is ollama running at %s?)", err, baseURL)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ollama error %d: %s", resp.StatusCode, string(b))
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var chunk struct {
			Response string `json:"response"`
			Done     bool   `json:"done"`
		}
		if err := json.Unmarshal(line, &chunk); err == nil {
			if chunk.Response != "" {
				chunkCallback(chunk.Response)
			}
			if chunk.Done {
				break
			}
		}
	}
	return scanner.Err()
}

func (s *CopilotService) streamOpenAI(ctx context.Context, req PromptRequest, chunkCallback func(string)) error {
	if req.APIKey == "" {
		return errors.New("OpenAI API key is required")
	}
	baseURL := req.BaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	model := req.Model
	if model == "" {
		model = "gpt-4o"
	}

	messages := []map[string]string{
		{"role": "system", "content": buildSystemPrompt(req.OSInfo, req.TerminalContext)},
		{"role": "user", "content": req.Prompt},
	}

	payload := map[string]interface{}{
		"model":    model,
		"messages": messages,
		"stream":   true,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+req.APIKey)

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("OpenAI error %d: %s", resp.StatusCode, string(b))
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				break
			}
			var streamResp struct {
				Choices []struct {
					Delta struct {
						Content string `json:"content"`
					} `json:"delta"`
				} `json:"choices"`
			}
			if err := json.Unmarshal([]byte(data), &streamResp); err == nil {
				if len(streamResp.Choices) > 0 && streamResp.Choices[0].Delta.Content != "" {
					chunkCallback(streamResp.Choices[0].Delta.Content)
				}
			}
		}
	}
	return scanner.Err()
}

func (s *CopilotService) streamAnthropic(ctx context.Context, req PromptRequest, chunkCallback func(string)) error {
	if req.APIKey == "" {
		return errors.New("Anthropic API key is required")
	}
	baseURL := req.BaseURL
	if baseURL == "" {
		baseURL = "https://api.anthropic.com/v1"
	}
	model := req.Model
	if model == "" {
		model = "claude-3-5-sonnet-20241022"
	}

	payload := map[string]interface{}{
		"model":      model,
		"max_tokens": 1024,
		"system":     buildSystemPrompt(req.OSInfo, req.TerminalContext),
		"messages": []map[string]string{
			{"role": "user", "content": req.Prompt},
		},
		"stream": true,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/messages", bytes.NewReader(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", req.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("anthropic error %d: %s", resp.StatusCode, string(b))
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			var event struct {
				Type  string `json:"type"`
				Delta struct {
					Text string `json:"text"`
				} `json:"delta"`
			}
			if err := json.Unmarshal([]byte(data), &event); err == nil {
				if event.Type == "content_block_delta" && event.Delta.Text != "" {
					chunkCallback(event.Delta.Text)
				}
			}
		}
	}
	return scanner.Err()
}

func (s *CopilotService) streamGemini(ctx context.Context, req PromptRequest, chunkCallback func(string)) error {
	if req.APIKey == "" {
		return errors.New("Gemini API key is required")
	}
	model := req.Model
	if model == "" {
		model = "gemini-2.0-flash"
	}

	endpoint := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:streamGenerateContent?alt=sse&key=%s", model, req.APIKey)

	payload := map[string]interface{}{
		"systemInstruction": map[string]interface{}{
			"parts": []map[string]string{
				{"text": buildSystemPrompt(req.OSInfo, req.TerminalContext)},
			},
		},
		"contents": []map[string]interface{}{
			{
				"role": "user",
				"parts": []map[string]string{
					{"text": req.Prompt},
				},
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("gemini error %d: %s", resp.StatusCode, string(b))
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			var geminiResp struct {
				Candidates []struct {
					Content struct {
						Parts []struct {
							Text string `json:"text"`
						} `json:"parts"`
					} `json:"content"`
				} `json:"candidates"`
			}
			if err := json.Unmarshal([]byte(data), &geminiResp); err == nil {
				if len(geminiResp.Candidates) > 0 {
					for _, p := range geminiResp.Candidates[0].Content.Parts {
						if p.Text != "" {
							chunkCallback(p.Text)
						}
					}
				}
			}
		}
	}
	return scanner.Err()
}
