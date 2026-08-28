package offlinebundle_test

import (
	"net/http"
	"net/http/httptest"
	"offlinebundle/internal/service"
	"offlinebundle/internal/storage"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkflowIntegration(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<title>Index</title><a href="/next">Next</a>`))
	}))
	defer srv.Close()
	store, err := storage.Open(filepath.Join(t.TempDir(), "db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	svc := service.New(store, srv.Client())
	result, err := svc.CreateBundle("integration", srv.URL+"/", t.TempDir())
	if err != nil || result.Job.ID != "integration" {
		t.Fatalf("result=%+v err=%v", result, err)
	}
}

func TestWorkflowCreateBundle(t *testing.T) {
	root := t.TempDir()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			w.Write([]byte(`<title>Home</title><a href="/guide">Guide</a><link href="/site.css">`))
		case "/guide":
			w.Write([]byte(`<title>Guide</title><img src="/logo.svg"><a href="https://outside.test/docs">Outside</a>`))
		case "/site.css":
			w.Write([]byte(`body { color: black }`))
		case "/logo.svg":
			w.Write([]byte(`<svg></svg>`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	output := filepath.Join(root, "output")
	store, err := storage.Open(filepath.Join(root, "db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	result, err := service.New(store, srv.Client()).CreateBundle("workflow-create", srv.URL+"/", output)
	if err != nil {
		t.Fatal(err)
	}
	if result.Job.Status != "complete" || result.Manifest.Incomplete {
		t.Fatalf("unexpected result: %+v", result)
	}
	if _, err := os.Stat(filepath.Join(output, "index.html")); err != nil {
		t.Fatal(err)
	}
}

func TestWorkflowQueryBundle(t *testing.T) {
	root := t.TempDir()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(`<title>Query</title>`)) }))
	defer srv.Close()
	dbPath := filepath.Join(root, "db")
	store, err := storage.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(root, "output")
	if _, err = service.New(store, srv.Client()).CreateBundle("workflow-query", srv.URL+"/", output); err != nil {
		t.Fatal(err)
	}
	if err = store.Close(); err != nil {
		t.Fatal(err)
	}
	store, err = storage.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	text, err := service.New(store, srv.Client()).ExportBundle("workflow-query", "summary")
	if err != nil || !strings.Contains(text, "status: COMPLETE") {
		t.Fatalf("summary=%q err=%v", text, err)
	}
}

func TestWorkflowRenderIndex(t *testing.T) {
	root := t.TempDir()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<title>Render</title><a href="https://outside.test">Outside</a>`))
	}))
	defer srv.Close()
	store, err := storage.Open(filepath.Join(root, "db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	output := filepath.Join(root, "output")
	result, err := service.New(store, srv.Client()).CreateBundle("workflow-render", srv.URL+"/", output)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(output, "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "External links") || len(result.Notices) != 1 {
		t.Fatalf("index=%s notices=%+v", data, result.Notices)
	}
}
