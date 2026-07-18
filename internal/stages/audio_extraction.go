package stages

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/Des331150/content-repurposing-tool/internal/pipeline"
)

type AudioExtraction struct {
	WorkDir string
}

func (s *AudioExtraction) Execute(ctx context.Context, video pipeline.SourceVideo, progress chan<- pipeline.ProgressUpdate) (*pipeline.AudioTrack, error) {
	outputPath := filepath.Join(s.WorkDir, "audio.wav")

	args := []string{
		"-i", video.Path,
		"-vn",
		"-acodec", "pcm_s16le",
		"-ar", "16000",
		"-ac", "1",
		"-f", "wav",
		"-y",
		outputPath,
	}

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("ffmpeg audio extraction failed: %w\nOutput: %s", err, string(output))
	}

	return &pipeline.AudioTrack{
		Path:     outputPath,
		Duration: video.Duration,
	}, nil
}
