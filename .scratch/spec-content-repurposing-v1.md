# Content Repurposing Tool — v1 Spec

## Problem Statement

I have long-form videos (talks, tutorials, podcasts) and want to extract short, engaging clips from them to share on social media platforms like TikTok, YouTube Shorts, and Instagram Reels. Existing tools like Opus Clip and repurpose.io do this but are paid, cloud-based services. I want a free, local, personal tool that I control end-to-end.

## Solution

A single-binary Go application that serves a local web UI. The user uploads a long-form video, the application runs it through a processing pipeline (transcription → narrative analysis → scene detection → clip selection → formatting), and presents the resulting short clips in a review grid. The user can accept, reject, or edit clips, then download them as MP4 files.

## User Stories

1. As a user, I want to upload a single long-form video file (MP4, MOV, AVI), so that the tool can process it into short clips.
2. As a user, I want to paste a URL to a video (YouTube, Google Drive, direct link), so that the tool can fetch and process it without me downloading it first.
3. As a user, I want the tool to transcribe the audio of my video using whisper.cpp, so that the transcript is available for narrative analysis.
4. As a user, I want the tool to detect scene boundaries in the video using ffmpeg, so that clips don't cut mid-scene.
5. As a user, I want the tool to analyze the transcript via an OpenRouter free-tier LLM to identify the most quotable narrative moments, so that clips capture interesting content.
6. As a user, I want the tool to intersect narrative moments with scene boundaries to produce clip candidates, so that each clip is both meaningful and visually clean.
7. As a user, I want to see a grid of clip candidates with thumbnails, durations, and the transcribed text snippet for each, so that I can quickly review what was extracted.
8. As a user, I want to accept individual clips, so that only the ones I choose move to export.
9. As a user, I want to reject individual clips, so that low-quality picks are discarded without exporting.
10. As a user, I want to edit a clip's start/end time before exporting, so that I can trim it to exactly the segment I want.
11. As a user, I want burnt-in captions on my clips (auto-generated from the transcript), so that viewers can follow along without audio.
12. As a user, I want all exported clips formatted in 9:16 vertical aspect ratio with a maximum duration of 60 seconds, so that they are compatible with TikTok, YouTube Shorts, and Instagram Reels.
13. As a user, I want to download individual clips as MP4 files, so that I can share them or upload them to platforms.
14. As a user, I want to download all accepted clips as a ZIP archive, so that I can batch-export in one click.
15. As a user, I want to see a progress indicator during pipeline processing, so that I know the tool hasn't frozen.
16. As a user, I want to configure default settings (whisper model size, output directory, default target platform) in a config.toml file, so that I don't have to set them every run.
17. As a user, I want to configure per-run settings (source video path/URL, target platforms) in the web UI, so that each run is flexible.
18. As a user, I want all uploaded and generated files cleaned up after I download my clips, so that storage doesn't accumulate.
19. As a user, I want the tool to launch as a single binary with no Node.js or Python runtime dependencies, so that setup is trivial.
20. As a user, I want the web UI served from the Go binary on localhost, so that I interact with the tool in my browser.

## Implementation Decisions

- **Language & runtime:** Go single binary. No Node.js, no Python runtime dependency.
- **Frontend:** Server-rendered HTML via Go's `html/template` + HTMX + Alpine.js. No build step.
- **Transcription:** whisper.cpp (`go-whisper` bindings or shell-out to whisper.cpp binary). Models downloaded on first use.
- **Scene detection:** ffmpeg's scene detection filter (`select='gt(scene,0.4)'`), called via `go-ffmpeg` or exec.
- **Narrative analysis:** HTTP call to OpenRouter free-tier LLM (e.g., `meta-llama/llama-3.2-3b-instruct:free`). Prompt instructs the model to return timestamps and quotes for the top ~10 narrative moments.
- **Pipeline architecture:** Pipes-and-filters. Each stage implements a common interface:
  - Input → Output contracts between stages
  - Pipeline runner orchestrates the sequence
  - Any stage can be swapped independently
- **Clip Selection:** Intersection algorithm — for each narrative moment, find the nearest scene boundary before and after, producing a clip with clean visual cuts.
- **Platform formatting:** Single profile for v1 — 9:16 vertical, 60s max duration, burnt-in captions using SRT-to-image overlay via ffmpeg.
- **Storage:** Ephemeral. Source video and generated clips stored in a temp directory, cleaned after the session ends or browser tab closes.
- **Config:** Single `config.toml` at tool root:
  ```toml
  [whisper]
  model_size = "medium"    # tiny, base, small, medium, large
  
  [export]
  output_dir = "./clips"
  default_platform = "tiktok"
  
  [openrouter]
  model = "meta-llama/llama-3.2-3b-instruct:free"
  ```
- **Review UI:** Clip card grid. Each card shows thumbnail (generated by ffmpeg at clip start time), duration, transcript snippet, and accept/reject/edit controls. Preview plays clip in a lightbox.
- **Pipeline concurrency:** One video at a time, blocking processing with progress updates streamed to the UI via HTMX SSE or polling.

## Testing Decisions

- **What makes a good test:** Test external behavior through stage interfaces. Do not test implementation details (which model, which ffmpeg flags). A good test verifies that given a known input, a stage produces the expected output.
- **Seams:**
  - **Pipeline Runner (single seam):** Mock all stage interfaces. Verify that the runner calls stages in order and passes outputs correctly between them. Verify error handling (if a stage fails, the pipeline stops gracefully).
  - **Clip Selection (pure logic):** The only stage with nontrivial logic not outsourced to an external tool. Test with synthetic narrative moments + scene boundaries. Verify clips are correctly bounded, deduplicated, and capped at max duration.
  - **Platform Formatting:** Test with a known clip + target platform. Verify output has correct aspect ratio, duration within limit, and captions are burned in.
- **What NOT to test:** The quality of whisper.cpp transcription, the accuracy of ffmpeg scene detection, or the judgment of the OpenRouter LLM. These are external systems — acceptance-test them manually, don't unit test.
- **Prior art:** Greenfield project — no existing tests. Tests will live in `*_test.go` files alongside each package, following Go conventions.

## Out of Scope

- Multi-video batch processing or queuing
- Cloud deployment or multi-user support
- Per-platform custom caption styles (one style for all v1 outputs)
- Voiceover or background music overlay
- Direct upload to social media platforms from the tool
- GIF export
- Video editing beyond start/end trim
- Analytics or usage tracking

## Further Notes

- OpenRouter free tier has rate limits — the pipeline may need to retry or wait between requests. The narrative analysis stage should handle HTTP 429 responses gracefully.
- whisper.cpp models are downloaded automatically on first run if not present locally. The config specifies which model size to use. The tool should guide the user through this on first launch.
- The tool name for the binary should be something short like `crp` or `repurpose`.
