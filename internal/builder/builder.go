package builder

import (
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"offlinebundle/internal/domain"
)

type Builder struct{ OutputDir string }

func New(output string) *Builder { return &Builder{OutputDir: output} }

func (b *Builder) Prepare() error {
	if strings.TrimSpace(b.OutputDir) == "" {
		return fmt.Errorf("output directory is required")
	}
	return os.MkdirAll(b.OutputDir, 0755)
}

func (b *Builder) WritePage(page domain.DocumentPage, body []byte) error {
	if err := b.Prepare(); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(b.OutputDir, page.Path), body, 0644)
}

func (b *Builder) WriteAsset(asset domain.AssetResource, body []byte) error {
	if asset.External {
		return nil
	}
	if err := b.Prepare(); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(b.OutputDir, asset.Path), body, 0644)
}

func (b *Builder) BuildIndex(job domain.CaptureJob, pages []domain.DocumentPage, manifest domain.BundleManifest) error {
	if err := b.Prepare(); err != nil {
		return err
	}
	sort.Slice(pages, func(i, j int) bool { return pages[i].URL < pages[j].URL })
	data := struct {
		Job      domain.CaptureJob
		Pages    []domain.DocumentPage
		Manifest domain.BundleManifest
	}{job, pages, manifest}
	tmpl := template.Must(template.New("index").Parse(`<!doctype html><html><head><meta charset="utf-8"><title>Offline bundle</title></head><body><h1>Offline bundle</h1><p>Status: {{.Job.Status}}</p><p>Incomplete: {{.Manifest.Incomplete}}</p><h2>Pages</h2><ul>{{range .Pages}}{{if eq .Status "failed"}}<li><s>{{if .Title}}{{.Title}}{{else}}{{.URL}}{{end}}</s> ({{.Status}}){{if .Error}} - {{.Error}}{{end}} <a href="{{.Path}}.error.txt">failure notice</a></li>{{else}}<li><a href="{{.Path}}">{{if .Title}}{{.Title}}{{else}}{{.URL}}{{end}}</a> ({{.Status}}){{if .Error}} - {{.Error}}{{end}}</li>{{end}}{{end}}</ul>{{if .Manifest.FailedPages}}<h2>Failed pages</h2><ul>{{range .Manifest.FailedPages}}<li>{{.}}</li>{{end}}</ul>{{end}}{{if .Manifest.FailedAssets}}<h2>Failed assets</h2><ul>{{range .Manifest.FailedAssets}}<li>{{.}}</li>{{end}}</ul>{{end}}{{if .Manifest.ExternalLinks}}<h2>External links</h2><ul>{{range .Manifest.ExternalLinks}}<li><a href="{{.}}">{{.}}</a></li>{{end}}</ul>{{end}}</body></html>`))
	f, err := os.Create(filepath.Join(b.OutputDir, "index.html"))
	if err != nil {
		return err
	}
	defer f.Close()
	return tmpl.Execute(f, data)
}

func (b *Builder) WriteManifest(manifest domain.BundleManifest) error {
	if err := b.Prepare(); err != nil {
		return err
	}
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(b.OutputDir, "manifest.json"), append(data, '\n'), 0644)
}

func (b *Builder) ReadManifest() (domain.BundleManifest, error) {
	var m domain.BundleManifest
	data, err := os.ReadFile(filepath.Join(b.OutputDir, "manifest.json"))
	if err != nil {
		return m, err
	}
	err = json.Unmarshal(data, &m)
	return m, err
}

func (b *Builder) OutputPath(name string) string { return filepath.Join(b.OutputDir, name) }

func (b *Builder) Exists(name string) bool { _, err := os.Stat(b.OutputPath(name)); return err == nil }

func SanitizeBody(body []byte) []byte {
	return []byte(strings.ReplaceAll(string(body), "<base ", "<meta data-original-base "))
}
