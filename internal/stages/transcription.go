package stages

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/Des331150/content-repurposing-tool/internal/pipeline"
)

type whisperOutput struct {
	SystemInfo    string `json:"systeminfo"`
	Model         string `json:"model"`
	Transcription struct {
		Text string `json:"text"`
	} `json:"transcription"`
	Segments []whisperSegment `json:"segments"`
}

type whisperSegment struct {
	Start float64 `json:"start"`
	End   float64 `json:"end"`
	Text  string  `json:"text"`
}

type Transcription struct {
	ModelSize string
	ModelDir  string
	BinPath   string
}

func NewTranscription(modelSize, modelDir, binPath string) *Transcription {
	if modelSize == "" {
		modelSize = "medium"
	}
	if modelDir == "" {
		modelDir = "models"
	}
	if binPath == "" {
		binPath = "whisper-cli"
	}
	return &Transcription{
		ModelSize: modelSize,
		ModelDir:  modelDir,
		BinPath:   binPath,
	}
}

func (s *Transcription) modelFilename() string {
	return "ggml-" + s.ModelSize + ".bin"
}

func (s *Transcription) modelURL() string {
	return "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/" + s.modelFilename()
}

func (s *Transcription) modelPath() string {
	return filepath.Join(s.ModelDir, s.modelFilename())
}

func (s *Transcription) ensureModel(ctx context.Context) error {
	path := s.modelPath()
	if _, err := os.Stat(path); err == nil {
		return nil
	}

	if err := os.MkdirAll(s.ModelDir, 0755); err != nil {
		return fmt.Errorf("creating model directory: %w", err)
	}

	url := s.modelURL()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("creating download request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("downloading model %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("downloading model %s: HTTP %d", url, resp.StatusCode)
	}

	f, err := os.Create(path + ".tmp")
	if err != nil {
		return fmt.Errorf("creating model file: %w", err)
	}

	if _, err := io.Copy(f, resp.Body); err != nil {
		f.Close()
		os.Remove(path + ".tmp")
		return fmt.Errorf("writing model file: %w", err)
	}
	f.Close()

	if err := os.Rename(path+".tmp", path); err != nil {
		return fmt.Errorf("renaming model file: %w", err)
	}

	return nil
}

func (s *Transcription) Execute(ctx context.Context, audio pipeline.AudioTrack, progress chan<- pipeline.ProgressUpdate) (*pipeline.Transcript, error) {
	sendStageProgress(progress, "transcription", 0, "Loading model...")

	if err := s.ensureModel(ctx); err != nil {
		return nil, fmt.Errorf("model setup: %w", err)
	}

	sendStageProgress(progress, "transcription", 20, "Transcribing audio...")

	outputJSON := audio.Path + ".json"
	args := []string{
		"-f", audio.Path,
		"-m", s.modelPath(),
		"--output-json",
		"--output-file", audio.Path,
	}

	cmd := exec.CommandContext(ctx, s.BinPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("whisper.cpp failed: %w\nOutput: %s", err, string(output))
	}

	sendStageProgress(progress, "transcription", 70, "Parsing transcription...")

	data, err := os.ReadFile(outputJSON)
	if err != nil {
		return nil, fmt.Errorf("reading whisper output: %w", err)
	}

	var wo whisperOutput
	if err := json.Unmarshal(data, &wo); err != nil {
		return nil, fmt.Errorf("parsing whisper output: %w", err)
	}

	segments := make([]pipeline.TranscriptSegment, len(wo.Segments))
	for i, seg := range wo.Segments {
		segments[i] = pipeline.TranscriptSegment{
			Start: seg.Start,
			End:   seg.End,
			Text:  seg.Text,
		}
	}

	sendStageProgress(progress, "transcription", 100, "Transcription complete")

	return &pipeline.Transcript{
		Text:      wo.Transcription.Text,
		Segments:  segments,
		AudioPath: audio.Path,
	}, nil
}

