package stages

import (
	"context"
	"path/filepath"

	"github.com/Des331150/content-repurposing-tool/internal/pipeline"
)

type Export struct {
	OutputDir string
}

func (s *Export) Execute(ctx context.Context, clips []pipeline.FormattedClip, progress chan<- pipeline.ProgressUpdate) ([]string, error) {
	var paths []string
	for _, clip := range clips {
		outputPath := filepath.Join(s.OutputDir, clip.Clip.ID+".mp4")
		paths = append(paths, outputPath)
	}
	return paths, nil
}
