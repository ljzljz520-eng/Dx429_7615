package domain

import (
	"sort"
	"strings"
)

type OutcomeCounts struct {
	Pages          int
	FetchedPages   int
	FailedPages    int
	Assets         int
	FetchedAssets  int
	FailedAssets   int
	ExternalAssets int
}

func CountOutcomes(pages []DocumentPage, assets []AssetResource) OutcomeCounts {
	c := OutcomeCounts{Pages: len(pages), Assets: len(assets)}
	for _, p := range pages {
		if p.Status == PageFetched {
			c.FetchedPages++
		}
		if p.Status == PageFailed {
			c.FailedPages++
		}
	}
	for _, a := range assets {
		switch a.Status {
		case AssetFetched:
			c.FetchedAssets++
		case AssetFailed:
			c.FailedAssets++
		case AssetExternal:
			c.ExternalAssets++
		}
	}
	return c
}
func (c OutcomeCounts) Failed() int        { return c.FailedPages + c.FailedAssets }
func (c OutcomeCounts) IsIncomplete() bool { return c.Failed() > 0 }
func (c OutcomeCounts) Text() string {
	return strings.Join([]string{"pages=" + itoa(c.Pages), "fetched_pages=" + itoa(c.FetchedPages), "failed_pages=" + itoa(c.FailedPages), "assets=" + itoa(c.Assets), "fetched_assets=" + itoa(c.FetchedAssets), "failed_assets=" + itoa(c.FailedAssets), "external_assets=" + itoa(c.ExternalAssets)}, " ")
}
func itoa(v int) string {
	if v == 0 {
		return "0"
	}
	digits := []byte{}
	for v > 0 {
		digits = append(digits, byte('0'+v%10))
		v /= 10
	}
	for i, j := 0, len(digits)-1; i < j; i, j = i+1, j-1 {
		digits[i], digits[j] = digits[j], digits[i]
	}
	return string(digits)
}
func SortURLs(values []string) []string {
	out := append([]string(nil), values...)
	sort.Strings(out)
	return out
}
func MergeLinks(a, b []string) []string {
	out := append([]string(nil), a...)
	for _, v := range b {
		found := false
		for _, existing := range out {
			if existing == v {
				found = true
				break
			}
		}
		if !found {
			out = append(out, v)
		}
	}
	return SortURLs(out)
}
func NormalizeStatuses(pages []DocumentPage, assets []AssetResource) {
	for i := range pages {
		if pages[i].Status == "" {
			pages[i].Status = PagePending
		}
	}
	for i := range assets {
		if assets[i].Status == "" {
			assets[i].Status = AssetPending
		}
	}
}
