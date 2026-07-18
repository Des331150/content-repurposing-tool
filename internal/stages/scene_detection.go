package stages

import (
	"context"

	"github.com/Des331150/content-repurposing-tool/internal/pipeline"
)

type SceneDetection struct{}

func (s *SceneDetection) Execute(ctx context.Context, video pipeline.SourceVideo, progress chan<- pipeline.ProgressUpdate) ([]pipeline.SceneBoundary, error) {
	return []pipeline.SceneBoundary{
		{Timestamp: 0},
		{Timestamp: 10},
		{Timestamp: 20},
		{Timestamp: 30},
		{Timestamp: 40},
		{Timestamp: 50},
		{Timestamp: 60},
	}, nil
}
