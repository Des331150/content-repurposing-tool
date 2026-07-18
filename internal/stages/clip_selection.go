package stages

import (
	"context"
	"fmt"
	"math"
	"sort"

	"github.com/Des331150/content-repurposing-tool/internal/pipeline"
)

const maxClipDuration = 60.0

type ClipSelection struct{}

func (s *ClipSelection) Execute(ctx context.Context, moments []pipeline.NarrativeMoment, boundaries []pipeline.SceneBoundary, progress chan<- pipeline.ProgressUpdate) ([]pipeline.ClipCandidate, error) {
	if len(moments) == 0 {
		return nil, nil
	}

	sort.Slice(boundaries, func(i, j int) bool {
		return boundaries[i].Timestamp < boundaries[j].Timestamp
	})

	used := make(map[string]bool)
	var clips []pipeline.ClipCandidate

	for i, moment := range moments {
		startBoundary := findNearestBoundaryBefore(moment.Start, boundaries)
		endBoundary := findNearestBoundaryAfter(moment.End, boundaries)

		clipStart := startBoundary
		clipEnd := endBoundary

		if clipEnd-clipStart > maxClipDuration {
			clipEnd = clipStart + maxClipDuration
		}

		key := fmt.Sprintf("%.2f-%.2f", clipStart, clipEnd)
		if used[key] {
			continue
		}
		used[key] = true

		clips = append(clips, pipeline.ClipCandidate{
			ID:                fmt.Sprintf("clip-%d", i+1),
			Start:             clipStart,
			End:               clipEnd,
			Duration:          clipEnd - clipStart,
			TranscriptSnippet: moment.Quote,
			Accepted:          true,
		})
	}

	return clips, nil
}

func findNearestBoundaryBefore(timestamp float64, boundaries []pipeline.SceneBoundary) float64 {
	nearest := 0.0
	for _, b := range boundaries {
		if b.Timestamp <= timestamp && b.Timestamp >= nearest {
			nearest = b.Timestamp
		}
	}
	return nearest
}

func findNearestBoundaryAfter(timestamp float64, boundaries []pipeline.SceneBoundary) float64 {
	nearest := math.MaxFloat64
	for _, b := range boundaries {
		if b.Timestamp >= timestamp && b.Timestamp < nearest {
			nearest = b.Timestamp
		}
	}
	return nearest
}
