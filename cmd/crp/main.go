package main

import (
	"flag"
	"log"
	"os"

	"github.com/Des331150/content-repurposing-tool/internal/config"
	"github.com/Des331150/content-repurposing-tool/internal/pipeline"
	"github.com/Des331150/content-repurposing-tool/internal/stages"
	"github.com/Des331150/content-repurposing-tool/internal/web"
)

func main() {
	port := flag.Int("port", 8080, "HTTP server port")
	configPath := flag.String("config", "config.toml", "Path to config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	workDir, err := os.MkdirTemp("", "crp-*")
	if err != nil {
		log.Fatalf("Failed to create work directory: %v", err)
	}
	defer os.RemoveAll(workDir)

	outputDir := cfg.Export.OutputDir
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	stageSet := pipeline.PipelineStages{
		AudioExtraction:    &stages.AudioExtraction{WorkDir: workDir},
		Transcription:      stages.NewTranscription(cfg.Whisper.ModelSize, cfg.Whisper.ModelDir, ""),
		NarrativeAnalysis:  &stages.NarrativeAnalysis{},
		SceneDetection:     &stages.SceneDetection{Threshold: cfg.SceneDetection.Threshold},
		ClipSelection:      &stages.ClipSelection{},
		PlatformFormatting: &stages.PlatformFormatting{WorkDir: workDir},
		Export:             &stages.Export{OutputDir: cfg.Export.OutputDir},
	}

	runner := pipeline.NewRunner(stageSet)
	server := web.NewServer(*port, workDir, cfg.Export.OutputDir, runner)
	log.Fatal(server.Start())
}