package stages

import (
	"context"
	"path/filepath"

	"github.com/Des331150/content-repurposing-tool/internal/pipeline"
)

type AudioExtraction struct {
	WorkDir string
}

func (s *AudioExtraction) Execute(ctx context.Context, video pipeline.SourceVideo, progress chan<- pipeline.ProgressUpdate) (*pipeline.AudioTrack, error) {
	outputPath := filepath.Join(s.WorkDir, "audio.wav")
	return &pipeline.AudioTrack{
		Path:     outputPath,
		Duration: video.Duration,
	}, nil
}
