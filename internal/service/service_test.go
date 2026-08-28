package service

import (
	"net/http"
	"net/http/httptest"
	"offlinebundle/internal/domain"
	"offlinebundle/internal/storage"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkflowCreateBundle(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<title>Docs</title><a href="/guide">Guide</a><link href="/site.css">`))
	}))
	defer srv.Close()
	store, err := storage.Open(filepath.Join(t.TempDir(), "db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	svc := New(store, srv.Client())
	res, err := svc.CreateBundle("job", srv.URL+"/", t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if res.Job.Status != "complete" || res.Manifest.Incomplete {
		t.Fatalf("result: %+v", res)
	}
	text, err := svc.InspectBundle("job")
	if err != nil || !strings.Contains(text, "status: COMPLETE") {
		t.Fatalf("inspect: %s %v", text, err)
	}
}

func TestWorkflowInspectIncompleteBundle(t *testing.T) {
	store, err := storage.Open(filepath.Join(t.TempDir(), "db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	j, _ := domain.NewCaptureJob("inspect", "https://docs.test/", t.TempDir())
	j.Status = domain.StatusIncomplete
	p := domain.NewDocumentPage("inspect", "https://docs.test/missing", "missing.html")
	p.MarkFailed(testServiceErr("not found"))
	m := domain.BuildManifest("inspect", []domain.DocumentPage{p}, nil, "index.html")
	if err = store.SaveBundleAtomic(j, []domain.DocumentPage{p}, nil, m); err != nil {
		t.Fatal(err)
	}
	svc := New(store, http.DefaultClient)
	text, err := svc.InspectBundle("inspect")
	if err != nil || !strings.Contains(text, "INCOMPLETE") || !strings.Contains(text, "missing") {
		t.Fatalf("text=%s err=%v", text, err)
	}
}

type testServiceErr string

func (e testServiceErr) Error() string { return string(e) }
