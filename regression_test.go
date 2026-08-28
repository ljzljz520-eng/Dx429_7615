package offlinebundle_test

import (
	"net/http"
	"net/http/httptest"
	"offlinebundle/internal/service"
	"offlinebundle/internal/storage"
	"path/filepath"
	"testing"
)

func TestOfflineBundleReportsFetchFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.Write([]byte(`<title>Index</title><a href="/unavailable">Unavailable</a>`))
			return
		}
		http.Error(w, "missing", http.StatusNotFound)
	}))
	defer srv.Close()
	store, err := storage.Open(filepath.Join(t.TempDir(), "db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	svc := service.New(store, srv.Client())
	result, err := svc.CreateBundle("regression", srv.URL+"/", t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if !result.Manifest.Incomplete {
		t.Fatalf("expected incomplete package, got %+v", result.Manifest)
	}
	if len(result.Manifest.FailedPages) != 1 {
		t.Fatalf("expected failed page, got %+v", result.Manifest)
	}
}
