package stages

import (
	"context"
	"fmt"
	"math"
	"os/exec"
	"regexp"
	"sort"
	"strconv"

	"github.com/Des331150/content-repurposing-tool/internal/pipeline"
)

var ptsTimeRe = regexp.MustCompile(`pts_time:(\d+\.?\d*)`)

type SceneDetection struct {
	Threshold float64
}

func (s *SceneDetection) Execute(ctx context.Context, video pipeline.SourceVideo, progress chan<- pipeline.ProgressUpdate) ([]pipeline.SceneBoundary, error) {
	sendStageProgress(progress, "scene_detection", 0, "Detecting scene boundaries...")

	threshold := s.Threshold

	args := []string{
		"-i", video.Path,
		"-filter:v", fmt.Sprintf("select='gt(scene,%f)',showinfo", threshold),
		"-f", "null",
		"-",
	}

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("ffmpeg scene detection failed: %w\nOutput: %s", err, string(output))
	}

	boundaries := parseSceneBoundaries(string(output))

	sort.Slice(boundaries, func(i, j int) bool {
		return boundaries[i].Timestamp < boundaries[j].Timestamp
	})

	if len(boundaries) == 0 || boundaries[0].Timestamp > 0 {
		boundaries = append([]pipeline.SceneBoundary{{Timestamp: 0}}, boundaries...)
	}

	msg := fmt.Sprintf("Found %d scene boundaries (threshold=%.2f)", len(boundaries), threshold)
	sendStageProgress(progress, "scene_detection", 100, msg)

	return boundaries, nil
}

func sendStageProgress(progress chan<- pipeline.ProgressUpdate, stage string, percent int, message string) {
	if progress != nil {
		progress <- pipeline.ProgressUpdate{Stage: stage, Percent: percent, Message: message}
	}
}

func parseSceneBoundaries(output string) []pipeline.SceneBoundary {
	seen := make(map[float64]bool)
	var boundaries []pipeline.SceneBoundary

	matches := ptsTimeRe.FindAllStringSubmatch(output, -1)
	for _, m := range matches {
		val, err := strconv.ParseFloat(m[1], 64)
		if err != nil {
			continue
		}
		if val == 0 || seen[val] {
			continue
		}
		seen[val] = true
		if !math.IsInf(val, 0) && !math.IsNaN(val) {
			boundaries = append(boundaries, pipeline.SceneBoundary{Timestamp: val})
		}
	}

	return boundaries
}
