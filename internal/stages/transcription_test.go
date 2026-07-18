package stages

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewTranscription_Defaults(t *testing.T) {
	s := NewTranscription("", "", "")
	if s.ModelSize != "medium" {
		t.Fatalf("expected default model size 'medium', got %q", s.ModelSize)
	}
	if s.ModelDir != "models" {
		t.Fatalf("expected default model dir 'models', got %q", s.ModelDir)
	}
	if s.BinPath != "whisper-cli" {
		t.Fatalf("expected default bin path 'whisper-cli', got %q", s.BinPath)
	}
}

func TestNewTranscription_CustomValues(t *testing.T) {
	s := NewTranscription("tiny", "/tmp/models", "/usr/local/bin/whisper")
	if s.ModelSize != "tiny" {
		t.Fatalf("expected model size 'tiny', got %q", s.ModelSize)
	}
	if s.ModelDir != "/tmp/models" {
		t.Fatalf("expected model dir '/tmp/models', got %q", s.ModelDir)
	}
	if s.BinPath != "/usr/local/bin/whisper" {
		t.Fatalf("expected bin path '/usr/local/bin/whisper', got %q", s.BinPath)
	}
}

func TestModelFilename(t *testing.T) {
	tests := []struct {
		size     string
		expected string
	}{
		{"tiny", "ggml-tiny.bin"},
		{"base", "ggml-base.bin"},
		{"small", "ggml-small.bin"},
		{"medium", "ggml-medium.bin"},
		{"large", "ggml-large.bin"},
	}
	for _, tt := range tests {
		s := NewTranscription(tt.size, "", "")
		got := s.modelFilename()
		if got != tt.expected {
			t.Errorf("modelFilename(%q) = %q, want %q", tt.size, got, tt.expected)
		}
	}
}

func TestModelURL(t *testing.T) {
	s := NewTranscription("tiny", "", "")
	url := s.modelURL()
	if !strings.HasPrefix(url, "https://huggingface.co/") {
		t.Fatalf("expected HuggingFace URL, got %q", url)
	}
	if !strings.HasSuffix(url, "ggml-tiny.bin") {
		t.Fatalf("expected URL ending with ggml-tiny.bin, got %q", url)
	}
}

func TestModelPath(t *testing.T) {
	s := NewTranscription("base", "/custom/models", "")
	path := s.modelPath()
	expected := filepath.Join("/custom/models", "ggml-base.bin")
	if path != expected {
		t.Fatalf("expected %q, got %q", expected, path)
	}
}

func TestModelURL_AllSizes(t *testing.T) {
	sizes := []string{"tiny", "base", "small", "medium", "large"}
	for _, size := range sizes {
		s := NewTranscription(size, "", "")
		url := s.modelURL()
		expected := "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-" + size + ".bin"
		if url != expected {
			t.Errorf("modelURL(%q) = %q, want %q", size, url, expected)
		}
	}
}

func TestWhisperJSONParsing(t *testing.T) {
	input := `{
		"systeminfo": "whisper_version",
		"model": "tiny",
		"transcription": {
			"text": "Hello world. This is a test."
		},
		"segments": [
			{"start": 0.0, "end": 2.5, "text": "Hello world.", "tokens": []},
			{"start": 2.5, "end": 5.0, "text": "This is a test.", "tokens": []}
		]
	}`

	var wo whisperOutput
	if err := json.Unmarshal([]byte(input), &wo); err != nil {
		t.Fatalf("failed to parse whisper JSON: %v", err)
	}

	if wo.Transcription.Text != "Hello world. This is a test." {
		t.Fatalf("unexpected transcription text: %q", wo.Transcription.Text)
	}

	if len(wo.Segments) != 2 {
		t.Fatalf("expected 2 segments, got %d", len(wo.Segments))
	}

	if wo.Segments[0].Start != 0.0 || wo.Segments[0].End != 2.5 || wo.Segments[0].Text != "Hello world." {
		t.Fatalf("unexpected first segment: %+v", wo.Segments[0])
	}

	if wo.Segments[1].Start != 2.5 || wo.Segments[1].End != 5.0 || wo.Segments[1].Text != "This is a test." {
		t.Fatalf("unexpected second segment: %+v", wo.Segments[1])
	}
}

func TestEnsureModel_ModelAlreadyExists(t *testing.T) {
	dir := t.TempDir()
	s := NewTranscription("tiny", dir, "")
	path := s.modelPath()
	if err := os.WriteFile(path, []byte("fake model"), 0644); err != nil {
		t.Fatalf("failed to write fake model: %v", err)
	}
	if err := s.ensureModel(context.Background()); err != nil {
		t.Fatalf("expected no error when model exists, got %v", err)
	}
}

func TestWhisperOutputJSONParsingRoundTrip(t *testing.T) {
	dir := t.TempDir()
	jsonPath := filepath.Join(dir, "output.json")
	input := `{
		"systeminfo": "whisper_version",
		"model": "tiny",
		"transcription": {
			"text": "Full transcription text."
		},
		"segments": [
			{"start": 0.0, "end": 3.0, "text": "Full transcription text.", "tokens": []}
		]
	}`
	if err := os.WriteFile(jsonPath, []byte(input), 0644); err != nil {
		t.Fatalf("failed to write test JSON: %v", err)
	}
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("failed to read test JSON: %v", err)
	}
	var wo whisperOutput
	if err := json.Unmarshal(data, &wo); err != nil {
		t.Fatalf("failed to parse whisper JSON: %v", err)
	}
	if wo.Transcription.Text != "Full transcription text." {
		t.Fatalf("unexpected text: %q", wo.Transcription.Text)
	}
}


