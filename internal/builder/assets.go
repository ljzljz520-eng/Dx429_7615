package builder

import (
	"offlinebundle/internal/domain"
	"os"
	"path/filepath"
)

func (b *Builder) WriteFailureNotice(page domain.DocumentPage) error {
	if err := b.Prepare(); err != nil {
		return err
	}
	body := []byte("Unable to fetch " + page.URL + "\nError: " + page.Error + "\n")
	return os.WriteFile(filepath.Join(b.OutputDir, page.Path+".error.txt"), body, 0644)
}
func (b *Builder) WriteExternalLinks(links []string) error {
	if err := b.Prepare(); err != nil {
		return err
	}
	body := []byte("External links are not downloaded.\n")
	for _, link := range links {
		body = append(body, []byte(link+"\n")...)
	}
	return os.WriteFile(filepath.Join(b.OutputDir, "external-links.txt"), body, 0644)
}
func (b *Builder) EnsureLayout() error {
	if err := b.Prepare(); err != nil {
		return err
	}
	for _, name := range []string{"pages", "assets"} {
		if err := os.MkdirAll(filepath.Join(b.OutputDir, name), 0755); err != nil {
			return err
		}
	}
	return nil
}
