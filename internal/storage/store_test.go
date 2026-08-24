package storage

import (
	"offlinebundle/internal/domain"
	"path/filepath"
	"testing"
)

func TestStorageRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bundle.db")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	j, _ := domain.NewCaptureJob("j", "https://docs.test/", "out")
	p := domain.NewDocumentPage("j", "https://docs.test/", "index.html")
	a := domain.NewAssetResource("j", "https://docs.test/a.css", "a.css", "stylesheet")
	m := domain.BuildManifest("j", []domain.DocumentPage{p}, []domain.AssetResource{a}, "out/index.html")
	if err = s.SaveBundleAtomic(j, []domain.DocumentPage{p}, []domain.AssetResource{a}, m); err != nil {
		t.Fatal(err)
	}
	got, err := s.LoadJob("j")
	if err != nil || got.ID != "j" {
		t.Fatalf("job: %+v %v", got, err)
	}
	if err = s.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestPersistenceSurvivesReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bundle.db")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	j, _ := domain.NewCaptureJob("reopen", "https://docs.test/", "out")
	p := domain.NewDocumentPage("reopen", "https://docs.test/", "index.html")
	a := domain.NewAssetResource("reopen", "https://docs.test/a.css", "a.css", "stylesheet")
	m := domain.BuildManifest("reopen", []domain.DocumentPage{p}, []domain.AssetResource{a}, "out/index.html")
	if err = s.SaveBundleAtomic(j, []domain.DocumentPage{p}, []domain.AssetResource{a}, m); err != nil {
		t.Fatal(err)
	}
	if err = s.Close(); err != nil {
		t.Fatal(err)
	}
	s, err = Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	gotJ, err := s.LoadJob("reopen")
	if err != nil || gotJ.RootURL != j.RootURL {
		t.Fatalf("job: %+v %v", gotJ, err)
	}
	pages, err := s.ListPages("reopen")
	if err != nil || len(pages) != 1 {
		t.Fatalf("pages: %+v %v", pages, err)
	}
	assets, err := s.ListAssets("reopen")
	if err != nil || len(assets) != 1 {
		t.Fatalf("assets: %+v %v", assets, err)
	}
	gotM, err := s.LoadManifest("reopen|manifest")
	if err != nil || gotM.PageCount != 1 {
		t.Fatalf("manifest: %+v %v", gotM, err)
	}
}

func TestStorageHealth(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "x.db"))
	if err != nil {
		t.Fatal(err)
	}
	if err = s.Health(); err != nil {
		t.Fatal(err)
	}
	s.Close()
	if err = s.Health(); err == nil {
		t.Fatal("expected closed error")
	}
}
