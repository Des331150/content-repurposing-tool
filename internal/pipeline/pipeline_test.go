package pipeline

import (
	"context"
	"errors"
	"testing"
)

type mockAudioExtraction struct {
	result *AudioTrack
	err    error
}

func (m *mockAudioExtraction) Execute(ctx context.Context, video SourceVideo, progress chan<- ProgressUpdate) (*AudioTrack, error) {
	return m.result, m.err
}

type mockTranscription struct {
	result *Transcript
	err    error
}

func (m *mockTranscription) Execute(ctx context.Context, audio AudioTrack, progress chan<- ProgressUpdate) (*Transcript, error) {
	return m.result, m.err
}

type mockNarrativeAnalysis struct {
	result []NarrativeMoment
	err    error
}

func (m *mockNarrativeAnalysis) Execute(ctx context.Context, transcript Transcript, progress chan<- ProgressUpdate) ([]NarrativeMoment, error) {
	return m.result, m.err
}

type mockSceneDetection struct {
	result []SceneBoundary
	err    error
}

func (m *mockSceneDetection) Execute(ctx context.Context, video SourceVideo, progress chan<- ProgressUpdate) ([]SceneBoundary, error) {
	return m.result, m.err
}

type mockClipSelection struct {
	result []ClipCandidate
	err    error
}

func (m *mockClipSelection) Execute(ctx context.Context, moments []NarrativeMoment, boundaries []SceneBoundary, progress chan<- ProgressUpdate) ([]ClipCandidate, error) {
	return m.result, m.err
}

type mockPlatformFormatting struct {
	result []FormattedClip
	err    error
}

func (m *mockPlatformFormatting) Execute(ctx context.Context, clips []ClipCandidate, progress chan<- ProgressUpdate) ([]FormattedClip, error) {
	return m.result, m.err
}

type mockExport struct {
	result []string
	err    error
}

func (m *mockExport) Execute(ctx context.Context, clips []FormattedClip, progress chan<- ProgressUpdate) ([]string, error) {
	return m.result, m.err
}

func TestRunner_ExecutesStagesInOrder(t *testing.T) {
	video := SourceVideo{Path: "test.mp4", Duration: 60}

	stages := PipelineStages{
		AudioExtraction: &mockAudioExtraction{
			result: &AudioTrack{Path: "audio.wav", Duration: 60},
		},
		Transcription: &mockTranscription{
			result: &Transcript{Text: "transcript", Segments: []TranscriptSegment{{Start: 0, End: 5, Text: "hello"}}},
		},
		NarrativeAnalysis: &mockNarrativeAnalysis{
			result: []NarrativeMoment{{Start: 0, End: 5, Quote: "hello", Relevance: 0.9}},
		},
		SceneDetection: &mockSceneDetection{
			result: []SceneBoundary{{Timestamp: 0}, {Timestamp: 10}},
		},
		ClipSelection: &mockClipSelection{
			result: []ClipCandidate{{ID: "clip-1", Start: 0, End: 5, Duration: 5, Accepted: true}},
		},
		PlatformFormatting: &mockPlatformFormatting{
			result: []FormattedClip{{Clip: ClipCandidate{ID: "clip-1"}, OutputPath: "clip-1.mp4"}},
		},
		Export: &mockExport{
			result: []string{"clip-1.mp4"},
		},
	}

	runner := NewRunner(stages)
	progress := make(chan ProgressUpdate, 100)
	run, err := runner.Run(context.Background(), video, progress)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if run.Status != "complete" {
		t.Fatalf("expected status 'complete', got %q", run.Status)
	}
	if len(run.ClipCandidates) != 1 || run.ClipCandidates[0].ID != "clip-1" {
		t.Fatalf("expected 1 clip with ID clip-1, got %+v", run.ClipCandidates)
	}

	progressUpdates := collectProgress(progress)
	if len(progressUpdates) == 0 {
		t.Fatal("expected progress updates")
	}
}

func TestRunner_StopsOnStageFailure(t *testing.T) {
	video := SourceVideo{Path: "test.mp4", Duration: 60}

	stages := PipelineStages{
		AudioExtraction: &mockAudioExtraction{
			err: errors.New("extraction failed"),
		},
		Transcription:      &mockTranscription{result: &Transcript{}},
		NarrativeAnalysis:  &mockNarrativeAnalysis{result: []NarrativeMoment{}},
		SceneDetection:     &mockSceneDetection{result: []SceneBoundary{}},
		ClipSelection:      &mockClipSelection{result: []ClipCandidate{}},
		PlatformFormatting: &mockPlatformFormatting{result: []FormattedClip{}},
		Export:             &mockExport{result: []string{}},
	}

	runner := NewRunner(stages)
	progress := make(chan ProgressUpdate, 100)
	run, err := runner.Run(context.Background(), video, progress)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if run.Status != "failed" {
		t.Fatalf("expected status 'failed', got %q", run.Status)
	}
}

func collectProgress(ch <-chan ProgressUpdate) []ProgressUpdate {
	var updates []ProgressUpdate
	for p := range ch {
		updates = append(updates, p)
	}
	return updates
}
