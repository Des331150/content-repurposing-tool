package stages

import (
	"context"

	"github.com/Des331150/content-repurposing-tool/internal/pipeline"
)

type Transcription struct{}

func (s *Transcription) Execute(ctx context.Context, audio pipeline.AudioTrack, progress chan<- pipeline.ProgressUpdate) (*pipeline.Transcript, error) {
	return &pipeline.Transcript{
		Text:      "This is a stub transcript. The video discusses important topics and key insights are shared throughout the presentation.",
		AudioPath: audio.Path,
		Segments: []pipeline.TranscriptSegment{
			{Start: 0, End: 5, Text: "Welcome to this presentation."},
			{Start: 5, End: 15, Text: "Today we're going to discuss some important topics that will help you understand the key concepts."},
			{Start: 15, End: 25, Text: "The first key insight is that consistency matters more than intensity."},
			{Start: 25, End: 35, Text: "Another important point is that you should focus on the fundamentals."},
			{Start: 35, End: 45, Text: "Let me share a story that illustrates this perfectly."},
			{Start: 45, End: 55, Text: "In conclusion, remember that small steps lead to big results over time."},
		},
	}, nil
}
