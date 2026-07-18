package stages

import (
	"context"
	"testing"

	"github.com/Des331150/content-repurposing-tool/internal/pipeline"
)

func TestParseSceneBoundaries_ValidOutput(t *testing.T) {
	output := `[Parsed_showinfo_0 @ 000001234] n:  24 pts:  96000 pts_time:4.000 duration:1001 duration_time:0.041
[Parsed_showinfo_0 @ 000005678] n:  48 pts: 192000 pts_time:8.000 duration:1001 duration_time:0.041
[Parsed_showinfo_0 @ 000009012] n:  72 pts: 288000 pts_time:12.000 duration:1001 duration_time:0.041`
	boundaries := parseSceneBoundaries(output)
	if len(boundaries) != 3 {
		t.Fatalf("expected 3 boundaries, got %d", len(boundaries))
	}
	expected := []float64{4, 8, 12}
	for i, b := range boundaries {
		if b.Timestamp != expected[i] {
			t.Errorf("boundary %d: expected timestamp %.1f, got %.1f", i, expected[i], b.Timestamp)
		}
	}
}

func TestParseSceneBoundaries_EmptyOutput(t *testing.T) {
	output := `ffmpeg version ... nothing useful here`
	boundaries := parseSceneBoundaries(output)
	if len(boundaries) != 0 {
		t.Fatalf("expected 0 boundaries, got %d", len(boundaries))
	}
}

func TestParseSceneBoundaries_DeduplicatesTimestamps(t *testing.T) {
	output := `pts_time:5.000
pts_time:5.000
pts_time:10.000
pts_time:10.000`
	boundaries := parseSceneBoundaries(output)
	if len(boundaries) != 2 {
		t.Fatalf("expected 2 boundaries (deduplicated), got %d", len(boundaries))
	}
}

func TestParseSceneBoundaries_SkipsZeroTimestamp(t *testing.T) {
	output := `pts_time:0.000
pts_time:5.000`
	boundaries := parseSceneBoundaries(output)
	for _, b := range boundaries {
		if b.Timestamp == 0 {
			t.Error("expected timestamp 0 to be excluded from parsing")
		}
	}
}

func TestSceneDetection_ExecuteErrorsWithoutFFmpeg(t *testing.T) {
	s := &SceneDetection{Threshold: 0.4}
	_, err := s.Execute(context.Background(), pipeline.SourceVideo{Path: "nonexistent.mp4", Duration: 60}, nil)
	if err == nil {
		t.Fatal("expected error when ffmpeg is not available or video is missing, got nil")
	}
}
