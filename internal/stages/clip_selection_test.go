package stages

import (
	"context"
	"testing"

	"github.com/Des331150/content-repurposing-tool/internal/pipeline"
)

func TestClipSelection_IntersectsMomentsWithBoundaries(t *testing.T) {
	moments := []pipeline.NarrativeMoment{
		{Start: 5, End: 12, Quote: "first moment", Relevance: 0.9},
		{Start: 20, End: 28, Quote: "second moment", Relevance: 0.8},
	}
	boundaries := []pipeline.SceneBoundary{
		{Timestamp: 0},
		{Timestamp: 10},
		{Timestamp: 20},
		{Timestamp: 30},
		{Timestamp: 40},
	}

	s := &ClipSelection{}
	clips, err := s.Execute(context.Background(), moments, boundaries, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(clips) != 2 {
		t.Fatalf("expected 2 clips, got %d", len(clips))
	}

	if clips[0].Start != 0 || clips[0].End != 20 {
		t.Fatalf("clip-1: expected start=0, end=20, got start=%.2f, end=%.2f", clips[0].Start, clips[0].End)
	}
	if clips[1].Start != 20 || clips[1].End != 30 {
		t.Fatalf("clip-2: expected start=20, end=30, got start=%.2f, end=%.2f", clips[1].Start, clips[1].End)
	}
}

func TestClipSelection_EmptyMomentsReturnsNoClips(t *testing.T) {
	s := &ClipSelection{}
	clips, err := s.Execute(context.Background(), nil, []pipeline.SceneBoundary{{Timestamp: 0}}, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(clips) != 0 {
		t.Fatalf("expected 0 clips, got %d", len(clips))
	}
}

func TestClipSelection_DeduplicatesOverlappingClips(t *testing.T) {
	moments := []pipeline.NarrativeMoment{
		{Start: 5, End: 12, Quote: "first", Relevance: 0.9},
		{Start: 8, End: 15, Quote: "second", Relevance: 0.8},
	}
	boundaries := []pipeline.SceneBoundary{
		{Timestamp: 0},
		{Timestamp: 10},
		{Timestamp: 20},
	}

	s := &ClipSelection{}
	clips, err := s.Execute(context.Background(), moments, boundaries, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(clips) != 1 {
		t.Fatalf("expected 1 deduplicated clip, got %d", len(clips))
	}
}

func TestClipSelection_CapsDurationAtMax(t *testing.T) {
	moments := []pipeline.NarrativeMoment{
		{Start: 5, End: 100, Quote: "long moment", Relevance: 0.9},
	}
	boundaries := []pipeline.SceneBoundary{
		{Timestamp: 0},
		{Timestamp: 200},
	}

	s := &ClipSelection{}
	clips, err := s.Execute(context.Background(), moments, boundaries, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(clips) != 1 {
		t.Fatalf("expected 1 clip, got %d", len(clips))
	}
	if clips[0].Duration > maxClipDuration+0.01 {
		t.Fatalf("expected duration <= %.0f, got %.2f", maxClipDuration, clips[0].Duration)
	}
}
