package web

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Des331150/content-repurposing-tool/internal/pipeline"
)

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	s.templates.ExecuteTemplate(w, "layout.html", map[string]interface{}{
		"Title": "Content Repurposing Tool",
		"Page":  "upload",
	})
}

func (s *Server) handleUpload(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(500 << 20); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("video")
	if err != nil {
		http.Error(w, "Failed to read uploaded file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = ".mp4"
	}
	videoPath := filepath.Join(s.workDir, "source"+ext)
	dst, err := os.Create(videoPath)
	if err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()
	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	s.mu.Lock()
	runID := fmt.Sprintf("run-%d", len(s.runs)+1)
	s.mu.Unlock()

	video := pipeline.SourceVideo{
		Path:     videoPath,
		Duration: 60,
	}

	progress := make(chan pipeline.ProgressUpdate, 100)
	go func() {
		run, err := s.pipeline.Run(context.Background(), video, progress)
		s.mu.Lock()
		s.runs[runID] = run
		s.mu.Unlock()
		if err != nil {
			log.Printf("Pipeline run %s failed: %v", runID, err)
		}
	}()

	go s.trackProgress(runID, progress)

	http.Redirect(w, r, fmt.Sprintf("/run/%s", runID), http.StatusSeeOther)
}

func (s *Server) handleViewRun(w http.ResponseWriter, r *http.Request) {
	runID := r.PathValue("id")
	s.mu.RLock()
	run, ok := s.runs[runID]
	s.mu.RUnlock()

	if !ok {
		http.Error(w, "Run not found", http.StatusNotFound)
		return
	}

	s.templates.ExecuteTemplate(w, "layout.html", map[string]interface{}{
		"Title": "Run " + runID,
		"Page":  "review",
		"Run":   run,
		"RunID": runID,
	})
}

func (s *Server) handleAcceptClip(w http.ResponseWriter, r *http.Request) {
	runID := r.PathValue("id")
	clipID := r.PathValue("clipID")

	s.mu.Lock()
	run, ok := s.runs[runID]
	if ok {
		for i := range run.ClipCandidates {
			if run.ClipCandidates[i].ID == clipID {
				run.ClipCandidates[i].Accepted = true
				break
			}
		}
	}
	s.mu.Unlock()

	http.Redirect(w, r, fmt.Sprintf("/run/%s", runID), http.StatusSeeOther)
}

func (s *Server) handleRejectClip(w http.ResponseWriter, r *http.Request) {
	runID := r.PathValue("id")
	clipID := r.PathValue("clipID")

	s.mu.Lock()
	run, ok := s.runs[runID]
	if ok {
		for i := range run.ClipCandidates {
			if run.ClipCandidates[i].ID == clipID {
				run.ClipCandidates[i].Accepted = false
				break
			}
		}
	}
	s.mu.Unlock()

	http.Redirect(w, r, fmt.Sprintf("/run/%s", runID), http.StatusSeeOther)
}

func (s *Server) handleEditClip(w http.ResponseWriter, r *http.Request) {
	runID := r.PathValue("id")
	clipID := r.PathValue("clipID")

	startStr := r.FormValue("start")
	endStr := r.FormValue("end")

	s.mu.Lock()
	run, ok := s.runs[runID]
	if ok {
		for i := range run.ClipCandidates {
			if run.ClipCandidates[i].ID == clipID {
				if startStr != "" {
					var start float64
					fmt.Sscanf(startStr, "%f", &start)
					run.ClipCandidates[i].Start = start
				}
				if endStr != "" {
					var end float64
					fmt.Sscanf(endStr, "%f", &end)
					run.ClipCandidates[i].End = end
				}
				run.ClipCandidates[i].Duration = run.ClipCandidates[i].End - run.ClipCandidates[i].Start
				break
			}
		}
	}
	s.mu.Unlock()

	http.Redirect(w, r, fmt.Sprintf("/run/%s", runID), http.StatusSeeOther)
}

func (s *Server) handleDownloadClip(w http.ResponseWriter, r *http.Request) {
	runID := r.PathValue("id")
	clipID := r.PathValue("clipID")

	s.mu.RLock()
	run, ok := s.runs[runID]
	s.mu.RUnlock()

	if !ok {
		http.Error(w, "Run not found", http.StatusNotFound)
		return
	}

	for _, clip := range run.ClipCandidates {
		if clip.ID == clipID {
			outputPath := filepath.Join(s.outputDir, clipID+".mp4")
			w.Header().Set("Content-Type", "video/mp4")
			w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.mp4", clipID))
			http.ServeFile(w, r, outputPath)
			return
		}
	}
	http.Error(w, "Clip not found", http.StatusNotFound)
}

func (s *Server) handleProgress(w http.ResponseWriter, r *http.Request) {
	runID := r.PathValue("id")
	s.mu.RLock()
	run, ok := s.runs[runID]
	s.mu.RUnlock()

	if !ok {
		http.Error(w, "Run not found", http.StatusNotFound)
		return
	}

	s.templates.ExecuteTemplate(w, "layout.html", map[string]interface{}{
		"Title": "Run " + runID,
		"Page":  "progress",
		"Run":   run,
		"RunID": runID,
	})
}

func (s *Server) handleDownloadAll(w http.ResponseWriter, r *http.Request) {
	runID := r.PathValue("id")
	s.mu.RLock()
	run, ok := s.runs[runID]
	s.mu.RUnlock()

	if !ok {
		http.Error(w, "Run not found", http.StatusNotFound)
		return
	}

	zipPath := filepath.Join(s.workDir, runID+".zip")
	if err := createZip(zipPath, run.ClipCandidates, s.outputDir); err != nil {
		http.Error(w, "Failed to create archive", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=clips.zip")
	http.ServeFile(w, r, zipPath)
}

func (s *Server) trackProgress(runID string, progress <-chan pipeline.ProgressUpdate) {
	for p := range progress {
		log.Printf("[%s] %s: %d%% - %s", runID, p.Stage, p.Percent, p.Message)
	}
}

func createZip(zipPath string, clips []pipeline.ClipCandidate, outputDir string) error {
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	for _, clip := range clips {
		if !clip.Accepted {
			continue
		}
		clipPath := filepath.Join(outputDir, clip.ID+".mp4")
		if _, err := os.Stat(clipPath); os.IsNotExist(err) {
			continue
		}
		f, err := os.Open(clipPath)
		if err != nil {
			return err
		}

		w, err := zipWriter.Create(clip.ID + ".mp4")
		if err != nil {
			f.Close()
			return err
		}
		if _, err := io.Copy(w, f); err != nil {
			f.Close()
			return err
		}
		f.Close()
	}

	return nil
}
