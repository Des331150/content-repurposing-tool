package stages

import (
	"context"

	"github.com/Des331150/content-repurposing-tool/internal/pipeline"
)

type NarrativeAnalysis struct{}

func (s *NarrativeAnalysis) Execute(ctx context.Context, transcript pipeline.Transcript, progress chan<- pipeline.ProgressUpdate) ([]pipeline.NarrativeMoment, error) {
	return []pipeline.NarrativeMoment{
		{Start: 5, End: 15, Quote: "Today we're going to discuss some important topics that will help you understand the key concepts.", Relevance: 0.9},
		{Start: 15, End: 25, Quote: "The first key insight is that consistency matters more than intensity.", Relevance: 0.85},
		{Start: 25, End: 35, Quote: "Another important point is that you should focus on the fundamentals.", Relevance: 0.8},
		{Start: 35, End: 45, Quote: "Let me share a story that illustrates this perfectly.", Relevance: 0.75},
		{Start: 45, End: 55, Quote: "In conclusion, remember that small steps lead to big results over time.", Relevance: 0.7},
	}, nil
}
