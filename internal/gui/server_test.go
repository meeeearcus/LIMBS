package gui

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"limbs/internal/config"
	"limbs/internal/exporter"
)

func TestHandleStaticIndex(t *testing.T) {
	s := newTestServer(func(config.Config) (exporter.Result, error) {
		return exporter.Result{}, nil
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	s.handleStatic(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("LIMBS GUI")) {
		t.Fatalf("expected index html body to contain LIMBS GUI")
	}
}

func TestAPIExportValidationError(t *testing.T) {
	s := newTestServer(func(config.Config) (exporter.Result, error) {
		t.Fatalf("runExport should not be called on validation error")
		return exporter.Result{}, nil
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/export", mustJSON(t, map[string]any{
		"projectName": "A",
	}))

	s.handleExport(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestAPIExportSuccess(t *testing.T) {
	t.Setenv("HOME", "/tmp/testhome")

	s := newTestServer(func(cfg config.Config) (exporter.Result, error) {
		if cfg.ProjectName != "MYPROJECT" {
			t.Fatalf("unexpected project name: %s", cfg.ProjectName)
		}
		if cfg.DestRoot != "/tmp/testhome/limbs" {
			t.Fatalf("expected sanitized/expanded dest root, got %q", cfg.DestRoot)
		}
		return exporter.Result{
			ProjectName: "MYPROJECT",
			OutputDir:   "/tmp/out/MYPROJECT_export",
			ZipPath:     "/tmp/out/MYPROJECT.zip",
			Stats: exporter.Stats{
				ReferencesFound:  2,
				UniqueReferences: 1,
				SamplesCopied:    1,
				MissingSamples: []exporter.MissingSample{
					{ReferencePath: "ref", ResolvedPath: "", Reason: "file not found"},
				},
			},
		}, nil
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/export", mustJSON(t, map[string]any{
		"projectName":  "MYPROJECT",
		"destRoot":     "\"$HOME/limbs\"",
		"allowMissing": true,
	}))

	s.handleExport(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var payload ExportResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.ProjectName != "MYPROJECT" {
		t.Fatalf("unexpected projectName: %s", payload.ProjectName)
	}
	if payload.Stats.SamplesCopied != 1 {
		t.Fatalf("unexpected samplesCopied: %d", payload.Stats.SamplesCopied)
	}
}

func TestAPIExportBusyConflict(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	s := newTestServer(func(config.Config) (exporter.Result, error) {
		close(started)
		<-release
		return exporter.Result{}, nil
	})

	var wg sync.WaitGroup
	wg.Add(1)
	firstRec := httptest.NewRecorder()
	firstReq := httptest.NewRequest(http.MethodPost, "/api/export", mustJSON(t, map[string]any{
		"projectName": "MYPROJECT",
		"destRoot":    "/tmp/out",
	}))

	go func() {
		defer wg.Done()
		s.handleExport(firstRec, firstReq)
	}()

	<-started
	time.Sleep(20 * time.Millisecond)

	secondRec := httptest.NewRecorder()
	secondReq := httptest.NewRequest(http.MethodPost, "/api/export", mustJSON(t, map[string]any{
		"projectName": "MYPROJECT",
		"destRoot":    "/tmp/out",
	}))

	s.handleExport(secondRec, secondReq)
	if secondRec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", secondRec.Code, secondRec.Body.String())
	}

	close(release)
	wg.Wait()
}

func TestAPIExportRejectsQuotedPathChars(t *testing.T) {
	s := newTestServer(func(config.Config) (exporter.Result, error) {
		t.Fatalf("runExport should not run for invalid quoted path")
		return exporter.Result{}, nil
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/export", mustJSON(t, map[string]any{
		"projectName": "MYPROJECT",
		"destRoot":    "/tmp/te\"st",
	}))

	s.handleExport(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func newTestServer(run func(config.Config) (exporter.Result, error)) *Server {
	staticSub, _ := fs.Sub(assetsFS, "assets")
	return &Server{runExport: run, staticFS: staticSub}
}

func mustJSON(t *testing.T, payload map[string]any) *bytes.Reader {
	t.Helper()
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	return bytes.NewReader(body)
}

func TestTrimWrappingQuotes(t *testing.T) {
	if got := trimWrappingQuotes("\"$HOME/limbs\""); got != "$HOME/limbs" {
		t.Fatalf("unexpected trimmed value: %q", got)
	}
	if got := trimWrappingQuotes("'$HOME/limbs'"); got != "$HOME/limbs" {
		t.Fatalf("unexpected trimmed value: %q", got)
	}
}

func TestExpandUserHome(t *testing.T) {
	t.Setenv("HOME", "/tmp/home")
	if got := expandUserHome("~/limbs"); got != "/tmp/home/limbs" {
		t.Fatalf("unexpected expanded value: %q", got)
	}
}
