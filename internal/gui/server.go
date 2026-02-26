package gui

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"limbs/internal/config"
	"limbs/internal/exporter"
)

//go:embed assets/*
var assetsFS embed.FS

type Options struct {
	Host string
	Port int
}

type Server struct {
	server *http.Server
	ln     net.Listener
	url    string

	mu        sync.Mutex
	busy      bool
	runExport func(config.Config) (exporter.Result, error)
	staticFS  fs.FS
}

func Start(opts Options) (*Server, error) {
	host := strings.TrimSpace(opts.Host)
	if host == "" {
		host = "127.0.0.1"
	}
	if opts.Port < 0 || opts.Port > 65535 {
		return nil, fmt.Errorf("invalid port: %d", opts.Port)
	}

	staticSub, err := fs.Sub(assetsFS, "assets")
	if err != nil {
		return nil, err
	}

	s := &Server{
		runExport: exporter.Run,
		staticFS:  staticSub,
	}

	mux := http.NewServeMux()
	s.registerRoutes(mux)

	addr := fmt.Sprintf("%s:%d", host, opts.Port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	s.server = &http.Server{Handler: mux}
	s.ln = ln
	s.url = fmt.Sprintf("http://%s", ln.Addr().String())

	go func() {
		_ = s.server.Serve(ln)
	}()

	return s, nil
}

func (s *Server) URL() string {
	return s.url
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.server == nil {
		return nil
	}
	err := s.server.Shutdown(ctx)
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (s *Server) registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/export", s.handleExport)
	mux.HandleFunc("/", s.handleStatic)
}

func (s *Server) handleStatic(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		http.ServeFileFS(w, r, s.staticFS, "index.html")
		return
	}
	name := path.Clean(strings.TrimPrefix(r.URL.Path, "/"))
	if name == "." || strings.HasPrefix(name, "..") {
		http.NotFound(w, r)
		return
	}
	f, err := s.staticFS.Open(name)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	_ = f.Close()
	http.ServeFileFS(w, r, s.staticFS, name)
}

func (s *Server) handleExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to read request body")
		return
	}

	var req ExportRequest
	if err := json.Unmarshal(reqBody, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}
	if err := sanitizeRequestPaths(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	cfg := config.Config{
		SourceMount:  req.SourceMount,
		ProjectsRoot: req.ProjectsRoot,
		SamplesRoot:  req.SamplesRoot,
		USBDrive:     req.USBDrive,
		ProjectName:  req.ProjectName,
		ProjectFile:  req.ProjectFile,
		DestRoot:     req.DestRoot,
		LimbsRoot:    req.LimbsRoot,
		Zip:          req.Zip,
		AllowMissing: req.AllowMissing,
	}

	resolved, err := config.Resolve(cfg)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	s.mu.Lock()
	if s.busy {
		s.mu.Unlock()
		writeError(w, http.StatusConflict, "export already running")
		return
	}
	s.busy = true
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		s.busy = false
		s.mu.Unlock()
	}()

	result, err := s.runExport(resolved)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, toExportResponse(result))
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, ErrorResponse{Error: msg})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func sanitizeRequestPaths(req *ExportRequest) error {
	var err error
	req.SourceMount, err = sanitizePathInput(req.SourceMount, "sourceMount")
	if err != nil {
		return err
	}
	req.ProjectsRoot, err = sanitizePathInput(req.ProjectsRoot, "projectsRoot")
	if err != nil {
		return err
	}
	req.SamplesRoot, err = sanitizePathInput(req.SamplesRoot, "samplesRoot")
	if err != nil {
		return err
	}
	req.USBDrive, err = sanitizePathInput(req.USBDrive, "usbDrive")
	if err != nil {
		return err
	}
	req.ProjectFile, err = sanitizePathInput(req.ProjectFile, "projectFile")
	if err != nil {
		return err
	}
	req.DestRoot, err = sanitizePathInput(req.DestRoot, "destRoot")
	if err != nil {
		return err
	}
	return nil
}

func sanitizePathInput(raw, field string) (string, error) {
	value := strings.TrimSpace(raw)
	value = trimWrappingQuotes(value)
	value = expandUserHome(value)
	value = os.ExpandEnv(value)

	if strings.Contains(value, "\"") || strings.Contains(value, "'") {
		return "", fmt.Errorf("%s contains quote characters; remove shell quoting", field)
	}
	return value, nil
}

func trimWrappingQuotes(s string) string {
	if len(s) < 2 {
		return s
	}
	first := s[0]
	last := s[len(s)-1]
	if (first == '"' && last == '"') || (first == '\'' && last == '\'') {
		return strings.TrimSpace(s[1 : len(s)-1])
	}
	return s
}

func expandUserHome(s string) string {
	if s == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return s
		}
		return home
	}

	if strings.HasPrefix(s, "~/") || strings.HasPrefix(s, "~\\") {
		home, err := os.UserHomeDir()
		if err != nil {
			return s
		}
		return filepath.Join(home, s[2:])
	}

	return s
}
