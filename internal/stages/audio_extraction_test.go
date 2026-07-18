package stages

import (
	"context"
	"testing"

	"github.com/Des331150/content-repurposing-tool/internal/pipeline"
)

func TestAudioExtraction_ExecuteErrorsWithoutFFmpeg(t *testing.T) {
	s := &AudioExtraction{WorkDir: t.TempDir()}
	_, err := s.Execute(context.Background(), pipeline.SourceVideo{Path: "nonexistent.mp4", Duration: 60}, nil)
	if err == nil {
		t.Fatal("expected error when ffmpeg is not available or video is missing, got nil")
	}
}
