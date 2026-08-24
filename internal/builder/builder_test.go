package builder

import (
	"offlinebundle/internal/domain"
	"os"
	"path/filepath"
	"testing"
)

func TestBuilderWritesFiles(t *testing.T) {
	dir := t.TempDir()
	b := New(dir)
	p := domain.NewDocumentPage("j", "https://docs.test/", "index.html")
	p.MarkFetched("Home", nil)
	if err := b.WritePage(p, []byte("<html>home</html>")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "index.html")); err != nil {
		t.Fatal(err)
	}
}
func TestBuilderWritesManifest(t *testing.T) {
	dir := t.TempDir()
	b := New(dir)
	m := domain.BuildManifest("j", nil, nil, "index.html")
	if err := b.WriteManifest(m); err != nil {
		t.Fatal(err)
	}
	got, err := b.ReadManifest()
	if err != nil || got.ID != m.ID {
		t.Fatalf("manifest: %+v %v", got, err)
	}
	if !b.Exists("manifest.json") {
		t.Fatal("missing")
	}
}
