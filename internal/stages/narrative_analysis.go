package stages

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Des331150/content-repurposing-tool/internal/pipeline"
)

const defaultOpenRouterURL = "https://openrouter.ai/api/v1/chat/completions"

type openRouterRequest struct {
	Model    string              `json:"model"`
	Messages []openRouterMessage `json:"messages"`
}

type openRouterMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openRouterResponse struct {
	Choices []openRouterChoice `json:"choices"`
}

type openRouterChoice struct {
	Message openRouterResponseMessage `json:"message"`
}

type openRouterResponseMessage struct {
	Content string `json:"content"`
}

type narrativeMomentsResponse struct {
	Moments []struct {
		Start     float64 `json:"start"`
		End       float64 `json:"end"`
		Quote     string  `json:"quote"`
		Relevance float64 `json:"relevance"`
	} `json:"moments"`
}

type httpDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

type NarrativeAnalysis struct {
	APIKey    string
	Model     string
	BaseURL   string
	HTTPDoer  httpDoer
	MaxRetries int
	RetryDelay time.Duration
}

func (s *NarrativeAnalysis) Execute(ctx context.Context, transcript pipeline.Transcript, progress chan<- pipeline.ProgressUpdate) ([]pipeline.NarrativeMoment, error) {
	sendStageProgress(progress, "narrative_analysis", 0, "Building prompt...")

	baseURL := s.BaseURL
	if baseURL == "" {
		baseURL = defaultOpenRouterURL
	}

	doer := s.HTTPDoer
	if doer == nil {
		doer = http.DefaultClient
	}

	maxRetries := s.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 3
	}

	delay := s.RetryDelay
	if delay <= 0 {
		delay = 2 * time.Second
	}

	model := s.Model
	if model == "" {
		model = "meta-llama/llama-3.2-3b-instruct:free"
	}

	prompt := buildPrompt(transcript)

	sendStageProgress(progress, "narrative_analysis", 20, "Calling OpenRouter LLM...")

	reqBody := openRouterRequest{
		Model: model,
		Messages: []openRouterMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: prompt},
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		if attempt > 1 {
			sendStageProgress(progress, "narrative_analysis", 20+attempt*10, fmt.Sprintf("Retrying (attempt %d/%d)...", attempt, maxRetries))
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
			delay *= 2
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL, bytes.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("creating request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+s.APIKey)

		resp, err := doer.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("HTTP request failed: %w", err)
			continue
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("reading response: %w", err)
			continue
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			lastErr = fmt.Errorf("rate limited (HTTP %d)", resp.StatusCode)
			if len(respBody) > 0 {
				lastErr = fmt.Errorf("%w: %s", lastErr, strings.TrimSpace(string(respBody)))
			}
			continue
		}

		if resp.StatusCode != http.StatusOK {
			detail := ""
			if len(respBody) > 0 {
				detail = ": " + strings.TrimSpace(string(respBody))
			}
			return nil, fmt.Errorf("OpenRouter API returned HTTP %d%s", resp.StatusCode, detail)
		}

		sendStageProgress(progress, "narrative_analysis", 80, "Parsing response...")

		moments, err := parseResponse(respBody)
		if err != nil {
			return nil, fmt.Errorf("parsing response: %w", err)
		}

		sendStageProgress(progress, "narrative_analysis", 100, fmt.Sprintf("Found %d narrative moments", len(moments)))
		return moments, nil
	}

	return nil, fmt.Errorf("OpenRouter API call failed after %d retries: %w", maxRetries, lastErr)
}

func buildPrompt(t pipeline.Transcript) string {
	var b strings.Builder
	b.WriteString("Here is the transcript of a video with timestamps:\n\n")
	for _, seg := range t.Segments {
		b.WriteString(fmt.Sprintf("[%.2f-%.2f] %s\n", seg.Start, seg.End, seg.Text))
	}
	b.WriteString("\n---\n")
	b.WriteString("Identify the top ~10 most quotable, self-contained narrative moments from this transcript. ")
	b.WriteString("For each moment, provide the start and end timestamps (in seconds), a brief quote or description, and a relevance score (0.0 to 1.0). ")
	b.WriteString("Moments should be non-overlapping and well-distributed across the video.\n\n")
	b.WriteString("Respond with ONLY a JSON object in the following format, no other text:\n")
	b.WriteString(`{"moments":[{"start":float,"end":float,"quote":"string","relevance":float}]}`)
	return b.String()
}

const systemPrompt = "You are a helpful assistant that analyzes video transcripts to identify the most quotable, shareable narrative moments. You always respond with valid JSON only."

func parseResponse(body []byte) ([]pipeline.NarrativeMoment, error) {
	var orResp openRouterResponse
	if err := json.Unmarshal(body, &orResp); err != nil {
		return nil, fmt.Errorf("unmarshaling OpenRouter response: %w", err)
	}

	if len(orResp.Choices) == 0 {
		return nil, fmt.Errorf("OpenRouter response has no choices")
	}

	content := orResp.Choices[0].Message.Content

	// The model might wrap JSON in markdown code fences
	content = extractJSON(content)

	var momentsResp narrativeMomentsResponse
	if err := json.Unmarshal([]byte(content), &momentsResp); err != nil {
		return nil, fmt.Errorf("unmarshaling narrative moments from LLM response: %w", err)
	}

	result := make([]pipeline.NarrativeMoment, len(momentsResp.Moments))
	for i, m := range momentsResp.Moments {
		result[i] = pipeline.NarrativeMoment{
			Start:     m.Start,
			End:       m.End,
			Quote:     m.Quote,
			Relevance: m.Relevance,
		}
	}

	return result, nil
}

func extractJSON(s string) string {
	s = strings.TrimSpace(s)
	// Strip markdown code fences if present
	if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```json")
		s = strings.TrimPrefix(s, "```")
		if idx := strings.LastIndex(s, "```"); idx >= 0 {
			s = s[:idx]
		}
		s = strings.TrimSpace(s)
	}
	if start := strings.Index(s, "{"); start >= 0 {
		if end := strings.LastIndex(s, "}"); end >= start {
			return s[start : end+1]
		}
	}
	return s
}
