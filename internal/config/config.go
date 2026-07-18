package config

import (
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Whisper        WhisperConfig        `toml:"whisper"`
	Export         ExportConfig         `toml:"export"`
	OpenRouter     OpenRouterConfig     `toml:"openrouter"`
	SceneDetection SceneDetectionConfig `toml:"scene_detection"`
}

type WhisperConfig struct {
	ModelSize string `toml:"model_size"`
	ModelDir  string `toml:"model_dir"`
}

type ExportConfig struct {
	OutputDir       string `toml:"output_dir"`
	DefaultPlatform string `toml:"default_platform"`
}

type OpenRouterConfig struct {
	Model string `toml:"model"`
}

type SceneDetectionConfig struct {
	Threshold float64 `toml:"threshold"`
}

func Load(path string) (*Config, error) {
	cfg := &Config{
		Whisper: WhisperConfig{
			ModelSize: "medium",
			ModelDir:  "models",
		},
		Export: ExportConfig{
			OutputDir:       "./clips",
			DefaultPlatform: "tiktok",
		},
		OpenRouter: OpenRouterConfig{
			Model: "meta-llama/llama-3.2-3b-instruct:free",
		},
		SceneDetection: SceneDetectionConfig{
			Threshold: 0.4,
		},
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return cfg, nil
	}

	_, err := toml.DecodeFile(path, cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
