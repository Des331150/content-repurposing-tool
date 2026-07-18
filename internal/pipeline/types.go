package pipeline

import "context"

type SourceVideo struct {
	Path     string
	Duration float64
}

type AudioTrack struct {
	Path     string
	Duration float64
}

type Transcript struct {
	Text      string
	Segments  []TranscriptSegment
	AudioPath string
}

type TranscriptSegment struct {
	Start float64
	End   float64
	Text  string
}

type NarrativeMoment struct {
	Start     float64
	End       float64
	Quote     string
	Relevance float64
}

type SceneBoundary struct {
	Timestamp float64
}

type ClipCandidate struct {
	ID             string
	Start          float64
	End            float64
	Duration       float64
	TranscriptSnippet string
	ThumbnailPath  string
	Accepted       bool
	SourceVideoPath string
}

type FormattedClip struct {
	Clip         ClipCandidate
	OutputPath   string
	Platform     string
}

type GenerationRun struct {
	ID              string
	SourceVideo     SourceVideo
	AudioTrack      *AudioTrack
	Transcript      *Transcript
	NarrativeMoments []NarrativeMoment
	SceneBoundaries  []SceneBoundary
	ClipCandidates   []ClipCandidate
	FormattedClips   []FormattedClip
	Status          string
	Error           string
}

type ProgressUpdate struct {
	Stage   string
	Percent int
	Message string
}

type AudioExtractionStage interface {
	Execute(ctx context.Context, video SourceVideo, progress chan<- ProgressUpdate) (*AudioTrack, error)
}

type TranscriptionStage interface {
	Execute(ctx context.Context, audio AudioTrack, progress chan<- ProgressUpdate) (*Transcript, error)
}

type NarrativeAnalysisStage interface {
	Execute(ctx context.Context, transcript Transcript, progress chan<- ProgressUpdate) ([]NarrativeMoment, error)
}

type SceneDetectionStage interface {
	Execute(ctx context.Context, video SourceVideo, progress chan<- ProgressUpdate) ([]SceneBoundary, error)
}

type ClipSelectionStage interface {
	Execute(ctx context.Context, moments []NarrativeMoment, boundaries []SceneBoundary, progress chan<- ProgressUpdate) ([]ClipCandidate, error)
}

type PlatformFormattingStage interface {
	Execute(ctx context.Context, clips []ClipCandidate, progress chan<- ProgressUpdate) ([]FormattedClip, error)
}

type ExportStage interface {
	Execute(ctx context.Context, clips []FormattedClip, progress chan<- ProgressUpdate) ([]string, error)
}

type PipelineStages struct {
	AudioExtraction    AudioExtractionStage
	Transcription      TranscriptionStage
	NarrativeAnalysis  NarrativeAnalysisStage
	SceneDetection     SceneDetectionStage
	ClipSelection      ClipSelectionStage
	PlatformFormatting PlatformFormattingStage
	Export             ExportStage
}
