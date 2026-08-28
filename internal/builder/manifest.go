package builder

import (
	"encoding/json"
	"offlinebundle/internal/domain"
	"os"
	"path/filepath"
)

func SaveManifestAt(dir string, manifest domain.BundleManifest) error {
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "manifest.json"), append(data, '\n'), 0644)
}

func LoadManifestAt(dir string) (domain.BundleManifest, error) {
	var m domain.BundleManifest
	data, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		return m, err
	}
	err = json.Unmarshal(data, &m)
	return m, err
}

func ManifestIncomplete(manifest domain.BundleManifest) bool {
	return manifest.Incomplete || len(manifest.FailedPages) > 0 || len(manifest.FailedAssets) > 0
}
