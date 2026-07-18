package pipeline

import (
	"context"
	"fmt"
)

type Runner struct {
	stages PipelineStages
}

func NewRunner(stages PipelineStages) *Runner {
	return &Runner{stages: stages}
}

func (r *Runner) Run(ctx context.Context, video SourceVideo, progress chan<- ProgressUpdate) (*GenerationRun, error) {
	defer close(progress)

	run := &GenerationRun{
		SourceVideo: video,
		Status:      "running",
	}

	if err := r.runStage(ctx, progress, "audio_extraction", "Extracting audio...", "Audio extracted", func() error {
		audio, err := r.stages.AudioExtraction.Execute(ctx, video, progress)
		if err != nil {
			return err
		}
		run.AudioTrack = audio
		return nil
	}); err != nil {
		run.Status = "failed"
		run.Error = fmt.Sprintf("audio extraction failed: %v", err)
		return run, err
	}

	if err := r.runStage(ctx, progress, "transcription", "Transcribing audio...", "Transcription complete", func() error {
		transcript, err := r.stages.Transcription.Execute(ctx, *run.AudioTrack, progress)
		if err != nil {
			return err
		}
		run.Transcript = transcript
		return nil
	}); err != nil {
		run.Status = "failed"
		run.Error = fmt.Sprintf("transcription failed: %v", err)
		return run, err
	}

	if err := r.runStage(ctx, progress, "narrative_analysis", "Analyzing narrative moments...", "Narrative analysis complete", func() error {
		moments, err := r.stages.NarrativeAnalysis.Execute(ctx, *run.Transcript, progress)
		if err != nil {
			return err
		}
		run.NarrativeMoments = moments
		return nil
	}); err != nil {
		run.Status = "failed"
		run.Error = fmt.Sprintf("narrative analysis failed: %v", err)
		return run, err
	}

	if err := r.runStage(ctx, progress, "scene_detection", "Detecting scene boundaries...", "Scene detection complete", func() error {
		boundaries, err := r.stages.SceneDetection.Execute(ctx, video, progress)
		if err != nil {
			return err
		}
		run.SceneBoundaries = boundaries
		return nil
	}); err != nil {
		run.Status = "failed"
		run.Error = fmt.Sprintf("scene detection failed: %v", err)
		return run, err
	}

	if err := r.runStage(ctx, progress, "clip_selection", "Selecting clips...", "Clip selection complete", func() error {
		clips, err := r.stages.ClipSelection.Execute(ctx, run.NarrativeMoments, run.SceneBoundaries, progress)
		if err != nil {
			return err
		}
		run.ClipCandidates = clips
		return nil
	}); err != nil {
		run.Status = "failed"
		run.Error = fmt.Sprintf("clip selection failed: %v", err)
		return run, err
	}

	if err := r.runStage(ctx, progress, "platform_formatting", "Formatting clips...", "Formatting complete", func() error {
		formatted, err := r.stages.PlatformFormatting.Execute(ctx, run.ClipCandidates, progress)
		if err != nil {
			return err
		}
		run.FormattedClips = formatted
		return nil
	}); err != nil {
		run.Status = "failed"
		run.Error = fmt.Sprintf("platform formatting failed: %v", err)
		return run, err
	}

	if err := r.runStage(ctx, progress, "export", "Exporting clips...", "Export complete", func() error {
		_, err := r.stages.Export.Execute(ctx, run.FormattedClips, progress)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		run.Status = "failed"
		run.Error = fmt.Sprintf("export failed: %v", err)
		return run, err
	}

	run.Status = "complete"
	return run, nil
}

func (r *Runner) runStage(ctx context.Context, progress chan<- ProgressUpdate, stageName, startMsg, endMsg string, fn func() error) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	sendProgress(progress, stageName, 0, startMsg)
	if err := fn(); err != nil {
		return err
	}
	sendProgress(progress, stageName, 100, endMsg)
	return nil
}

func sendProgress(progress chan<- ProgressUpdate, stage string, percent int, message string) {
	if progress != nil {
		progress <- ProgressUpdate{Stage: stage, Percent: percent, Message: message}
	}
}
