package stages

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Des331150/content-repurposing-tool/internal/pipeline"
)

func TestNarrativeAnalysis_SendsTranscriptAndParsesResponse(t *testing.T) {
	var receivedBody map[string]any
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-key" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		json.NewDecoder(r.Body).Decode(&receivedBody)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"choices": [{
				"message": {
					"content": "{\"moments\":[{\"start\":10.0,\"end\":20.5,\"quote\":\"first quote\",\"relevance\":0.9},{\"start\":30.0,\"end\":40.0,\"quote\":\"second quote\",\"relevance\":0.8}]}"
				}
			}]
		}`))
	}))
	defer ts.Close()

	s := &NarrativeAnalysis{
		APIKey:    "test-key",
		Model:     "test-model",
		BaseURL:   ts.URL,
		HTTPDoer:  ts.Client(),
	}

	transcript := pipeline.Transcript{
		Text: "full transcript text",
		Segments: []pipeline.TranscriptSegment{
			{Start: 0, End: 5, Text: "hello world"},
			{Start: 5, End: 10, Text: "this is a test"},
		},
	}

	moments, err := s.Execute(context.Background(), transcript, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(moments) != 2 {
		t.Fatalf("expected 2 moments, got %d", len(moments))
	}
	if moments[0].Start != 10.0 || moments[0].End != 20.5 || moments[0].Quote != "first quote" || moments[0].Relevance != 0.9 {
		t.Fatalf("unexpected first moment: %+v", moments[0])
	}
	if moments[1].Start != 30.0 || moments[1].End != 40.0 || moments[1].Quote != "second quote" || moments[1].Relevance != 0.8 {
		t.Fatalf("unexpected second moment: %+v", moments[1])
	}

	messages, ok := receivedBody["messages"].([]any)
	if !ok || len(messages) != 2 {
		t.Fatalf("expected 2 messages (system + user), got %v", receivedBody["messages"])
	}
	userMsg := messages[1].(map[string]any)
	content, ok := userMsg["content"].(string)
	if !ok {
		t.Fatalf("expected user message content to be a string")
	}
	if !strings.Contains(content, "hello world") || !strings.Contains(content, "this is a test") {
		t.Fatalf("expected transcript content in prompt, got: %s", content)
	}

	model, ok := receivedBody["model"].(string)
	if !ok || model != "test-model" {
		t.Fatalf("expected model 'test-model', got %v", receivedBody["model"])
	}
}

func TestNarrativeAnalysis_RetriesOnRateLimit(t *testing.T) {
	attempts := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"choices": [{
				"message": {
					"content": "{\"moments\":[]}"
				}
			}]
		}`))
	}))
	defer ts.Close()

	s := &NarrativeAnalysis{
		APIKey:    "key",
		Model:     "m",
		BaseURL:   ts.URL,
		HTTPDoer:  ts.Client(),
		MaxRetries: 3,
		RetryDelay: time.Millisecond,
	}

	transcript := pipeline.Transcript{
		Text: "test",
		Segments: []pipeline.TranscriptSegment{
			{Start: 0, End: 1, Text: "test"},
		},
	}

	_, err := s.Execute(context.Background(), transcript, nil)
	if err != nil {
		t.Fatalf("expected no error after retries, got %v", err)
	}
	if attempts != 3 {
		t.Fatalf("expected 3 attempts (2 failures + 1 success), got %d", attempts)
	}
}

func TestNarrativeAnalysis_ReturnsErrorAfterMaxRetries(t *testing.T) {
	attempts := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer ts.Close()

	s := &NarrativeAnalysis{
		APIKey:    "key",
		Model:     "m",
		BaseURL:   ts.URL,
		HTTPDoer:  ts.Client(),
		MaxRetries: 3,
		RetryDelay: time.Millisecond,
	}

	transcript := pipeline.Transcript{
		Text: "test",
		Segments: []pipeline.TranscriptSegment{
			{Start: 0, End: 1, Text: "test"},
		},
	}

	_, err := s.Execute(context.Background(), transcript, nil)
	if err == nil {
		t.Fatal("expected error after exhausting retries")
	}
	if !strings.Contains(err.Error(), "rate limited") {
		t.Fatalf("expected rate limit error, got: %v", err)
	}
	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
}

func TestNarrativeAnalysis_HandlesNetworkError(t *testing.T) {
	s := &NarrativeAnalysis{
		APIKey:    "key",
		Model:     "m",
		BaseURL:   "http://127.0.0.1:1",
		HTTPDoer:  http.DefaultClient,
		MaxRetries: 1,
		RetryDelay: time.Millisecond,
	}

	transcript := pipeline.Transcript{
		Text: "test",
		Segments: []pipeline.TranscriptSegment{
			{Start: 0, End: 1, Text: "test"},
		},
	}

	_, err := s.Execute(context.Background(), transcript, nil)
	if err == nil {
		t.Fatal("expected error for network failure")
	}
}

func TestNarrativeAnalysis_HandlesBadJSONResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`not json`))
	}))
	defer ts.Close()

	s := &NarrativeAnalysis{
		APIKey:    "key",
		Model:     "m",
		BaseURL:   ts.URL,
		HTTPDoer:  ts.Client(),
		MaxRetries: 1,
		RetryDelay: time.Millisecond,
	}

	transcript := pipeline.Transcript{
		Text: "test",
		Segments: []pipeline.TranscriptSegment{
			{Start: 0, End: 1, Text: "test"},
		},
	}

	_, err := s.Execute(context.Background(), transcript, nil)
	if err == nil {
		t.Fatal("expected error for bad JSON response")
	}
}

func TestNarrativeAnalysis_HandlesEmptyMomentsResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"choices": [{
				"message": {
					"content": "{\"moments\":[]}"
				}
			}]
		}`))
	}))
	defer ts.Close()

	s := &NarrativeAnalysis{
		APIKey:   "key",
		Model:    "m",
		BaseURL:  ts.URL,
		HTTPDoer: ts.Client(),
	}

	transcript := pipeline.Transcript{
		Text: "test",
		Segments: []pipeline.TranscriptSegment{
			{Start: 0, End: 1, Text: "test"},
		},
	}

	moments, err := s.Execute(context.Background(), transcript, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(moments) != 0 {
		t.Fatalf("expected 0 moments, got %d", len(moments))
	}
}

func TestNarrativeAnalysis_HandlesNon200Status(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "bad request"}`))
	}))
	defer ts.Close()

	s := &NarrativeAnalysis{
		APIKey:    "key",
		Model:     "m",
		BaseURL:   ts.URL,
		HTTPDoer:  ts.Client(),
		MaxRetries: 1,
		RetryDelay: time.Millisecond,
	}

	transcript := pipeline.Transcript{
		Text: "test",
		Segments: []pipeline.TranscriptSegment{
			{Start: 0, End: 1, Text: "test"},
		},
	}

	_, err := s.Execute(context.Background(), transcript, nil)
	if err == nil {
		t.Fatal("expected error for bad request")
	}
}
