package web

import (
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sync"

	"github.com/Des331150/content-repurposing-tool/internal/pipeline"
)

//go:embed templates/*.html
var templateFS embed.FS

type Server struct {
	port      int
	workDir   string
	outputDir string
	pipeline  *pipeline.Runner
	templates *template.Template
	mu        sync.RWMutex
	runs      map[string]*pipeline.GenerationRun
}

func NewServer(port int, workDir, outputDir string, p *pipeline.Runner) *Server {
	s := &Server{
		port:      port,
		workDir:   workDir,
		outputDir: outputDir,
		pipeline:  p,
		runs:      make(map[string]*pipeline.GenerationRun),
	}
	s.loadTemplates()
	return s
}

func (s *Server) loadTemplates() {
	s.templates = template.Must(template.ParseFS(templateFS, "templates/*.html"))
}

func (s *Server) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", s.handleIndex)
	mux.HandleFunc("POST /upload", s.handleUpload)
	mux.HandleFunc("GET /run/{id}", s.handleViewRun)
	mux.HandleFunc("POST /run/{id}/clip/{clipID}/accept", s.handleAcceptClip)
	mux.HandleFunc("POST /run/{id}/clip/{clipID}/reject", s.handleRejectClip)
	mux.HandleFunc("POST /run/{id}/clip/{clipID}/edit", s.handleEditClip)
	mux.HandleFunc("GET /run/{id}/clip/{clipID}/download", s.handleDownloadClip)
	mux.HandleFunc("GET /run/{id}/progress", s.handleProgress)
	mux.HandleFunc("GET /run/{id}/download-all", s.handleDownloadAll)

	addr := fmt.Sprintf(":%d", s.port)
	log.Printf("Starting server on http://localhost%s", addr)
	return http.ListenAndServe(addr, mux)
}