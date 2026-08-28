package builder

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"offlinebundle/internal/domain"
)

type OutputCheck struct {
	Root        string
	Files       []string
	Missing     []string
	Readable    bool
	HasIndex    bool
	HasManifest bool
}

func (b *Builder) Verify(job domain.CaptureJob, pages []domain.DocumentPage, assets []domain.AssetResource, manifest domain.BundleManifest) (OutputCheck, error) {
	if b == nil {
		return OutputCheck{}, errors.New("builder is nil")
	}
	check := OutputCheck{Root: b.OutputDir, Files: []string{}, Missing: []string{}}
	if _, err := os.Stat(b.OutputDir); err != nil {
		return check, err
	}
	check.Readable = true
	for _, page := range pages {
		if page.Status != domain.PageFetched {
			continue
		}
		name := page.Path
		if _, err := os.Stat(filepath.Join(b.OutputDir, name)); err != nil {
			check.Missing = append(check.Missing, name)
		} else {
			check.Files = append(check.Files, name)
		}
	}
	for _, asset := range assets {
		if asset.External || asset.Status != domain.AssetFetched {
			continue
		}
		if _, err := os.Stat(filepath.Join(b.OutputDir, asset.Path)); err != nil {
			check.Missing = append(check.Missing, asset.Path)
		} else {
			check.Files = append(check.Files, asset.Path)
		}
	}
	_, indexErr := os.Stat(filepath.Join(b.OutputDir, "index.html"))
	check.HasIndex = indexErr == nil
	_, manifestErr := os.Stat(filepath.Join(b.OutputDir, "manifest.json"))
	check.HasManifest = manifestErr == nil
	sort.Strings(check.Files)
	sort.Strings(check.Missing)
	if !check.HasIndex || !check.HasManifest {
		return check, errors.New("bundle metadata is incomplete")
	}
	if len(check.Missing) > 0 {
		return check, fmt.Errorf("bundle is missing %d resources", len(check.Missing))
	}
	if job.ID == "" || manifest.JobID != job.ID {
		return check, errors.New("bundle metadata does not match job")
	}
	return check, nil
}

func (c OutputCheck) Complete() bool {
	return c.Readable && c.HasIndex && c.HasManifest && len(c.Missing) == 0
}

func (c OutputCheck) Summary() string {
	status := "invalid"
	if c.Complete() {
		status = "ready"
	}
	return strings.Join([]string{status, "files=" + formatCount(len(c.Files)), "missing=" + formatCount(len(c.Missing))}, " ")
}

func formatCount(value int) string {
	if value == 0 {
		return "0"
	}
	digits := []byte{}
	for value > 0 {
		digits = append(digits, byte('0'+value%10))
		value /= 10
	}
	for left, right := 0, len(digits)-1; left < right; left, right = left+1, right-1 {
		digits[left], digits[right] = digits[right], digits[left]
	}
	return string(digits)
}

func (b *Builder) WriteReadme(job domain.CaptureJob, manifest domain.BundleManifest) error {
	if err := b.Prepare(); err != nil {
		return err
	}
	status := "complete"
	if manifest.Incomplete {
		status = "incomplete"
	}
	content := []string{"Offline documentation bundle", "Job: " + job.ID, "Status: " + status, "Open index.html in a browser."}
	if len(manifest.FailedPages) > 0 {
		content = append(content, "Failed pages: "+strings.Join(manifest.FailedPages, ", "))
	}
	if len(manifest.FailedAssets) > 0 {
		content = append(content, "Failed assets: "+strings.Join(manifest.FailedAssets, ", "))
	}
	return os.WriteFile(filepath.Join(b.OutputDir, "BUNDLE.txt"), []byte(strings.Join(content, "\n")+"\n"), 0644)
}
