package stages

import (
	"context"
	"path/filepath"

	"github.com/Des331150/content-repurposing-tool/internal/pipeline"
)

type PlatformFormatting struct {
	WorkDir string
}

func (s *PlatformFormatting) Execute(ctx context.Context, clips []pipeline.ClipCandidate, progress chan<- pipeline.ProgressUpdate) ([]pipeline.FormattedClip, error) {
	var formatted []pipeline.FormattedClip
	for _, clip := range clips {
		if !clip.Accepted {
			continue
		}
		outputPath := filepath.Join(s.WorkDir, "formatted_"+clip.ID+".mp4")
		formatted = append(formatted, pipeline.FormattedClip{
			Clip:       clip,
			OutputPath: outputPath,
			Platform:   "tiktok",
		})
	}

	return formatted, nil
}
