package report

import (
	"fmt"
	"sort"
	"strings"

	"offlinebundle/internal/domain"
)

func Format(job domain.CaptureJob, manifest domain.BundleManifest, pages []domain.DocumentPage, assets []domain.AssetResource) string {
	return FormatWithNotices(job, manifest, pages, assets, nil)
}

func FormatWithNotices(job domain.CaptureJob, manifest domain.BundleManifest, pages []domain.DocumentPage, assets []domain.AssetResource, notices []domain.ExternalLinkNotice) string {
	lines := []string{fmt.Sprintf("job: %s", job.ID), fmt.Sprintf("status: %s", domain.StatusLabel(job.Status)), fmt.Sprintf("output: %s", job.OutputDir), fmt.Sprintf("pages: %d", len(pages)), fmt.Sprintf("assets: %d", len(assets)), fmt.Sprintf("incomplete: %t", manifest.Incomplete)}
	failed := FailedPageURLs(pages)
	if len(failed) > 0 {
		lines = append(lines, "failed pages:")
		for _, u := range failed {
			lines = append(lines, "- "+u)
		}
	} else {
		lines = append(lines, "failed pages: none")
	}
	failedAssets := FailedAssetURLs(assets)
	if len(failedAssets) > 0 {
		lines = append(lines, "failed assets:")
		for _, u := range failedAssets {
			lines = append(lines, "- "+u)
		}
	}
	if len(manifest.ExternalLinks) > 0 {
		lines = append(lines, "external links:")
		for _, u := range UniqueSorted(manifest.ExternalLinks) {
			lines = append(lines, "- "+u)
		}
	}
	if len(notices) > 0 {
		lines = append(lines, "external notices:")
		for _, notice := range notices {
			lines = append(lines, "- "+notice.Display())
		}
	}
	return strings.Join(lines, "\n") + "\n"
}

func FailedPageURLs(pages []domain.DocumentPage) []string {
	out := []string{}
	for _, p := range pages {
		if p.IsFailed() {
			out = append(out, p.URL)
		}
	}
	sort.Strings(out)
	return out
}
func FailedAssetURLs(assets []domain.AssetResource) []string {
	out := []string{}
	for _, a := range assets {
		if a.IsFailed() {
			out = append(out, a.URL)
		}
	}
	sort.Strings(out)
	return out
}
func UniqueSorted(values []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, v := range values {
		if !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	sort.Strings(out)
	return out
}
func IsIncomplete(job domain.CaptureJob, manifest domain.BundleManifest) bool {
	return job.Status == domain.StatusIncomplete || manifest.Incomplete
}
func SummaryLine(manifest domain.BundleManifest) string {
	return fmt.Sprintf("%d pages, %d assets, %d failures", manifest.PageCount, manifest.AssetCount, len(manifest.FailedPages)+len(manifest.FailedAssets))
}
