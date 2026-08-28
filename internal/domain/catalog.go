package domain

import (
	"errors"
	"sort"
	"strings"
)

type ResourceCatalog struct {
	Job     CaptureJob
	Pages   map[string]DocumentPage
	Assets  map[string]AssetResource
	Notices map[string]ExternalLinkNotice
}

func NewResourceCatalog(job CaptureJob) *ResourceCatalog {
	return &ResourceCatalog{Job: job, Pages: map[string]DocumentPage{}, Assets: map[string]AssetResource{}, Notices: map[string]ExternalLinkNotice{}}
}

func (c *ResourceCatalog) AddPage(page DocumentPage) error {
	if c == nil {
		return errors.New("catalog is nil")
	}
	if page.JobID != c.Job.ID {
		return errors.New("page belongs to another job")
	}
	if strings.TrimSpace(page.URL) == "" {
		return errors.New("page URL is required")
	}
	c.Pages[page.URL] = page
	return nil
}

func (c *ResourceCatalog) AddAsset(asset AssetResource) error {
	if c == nil {
		return errors.New("catalog is nil")
	}
	if asset.JobID != c.Job.ID {
		return errors.New("asset belongs to another job")
	}
	if strings.TrimSpace(asset.URL) == "" {
		return errors.New("asset URL is required")
	}
	c.Assets[asset.URL] = asset
	return nil
}

func (c *ResourceCatalog) AddNotice(notice ExternalLinkNotice) error {
	if c == nil {
		return errors.New("catalog is nil")
	}
	if notice.JobID != c.Job.ID {
		return errors.New("notice belongs to another job")
	}
	if strings.TrimSpace(notice.TargetURL) == "" {
		return errors.New("target URL is required")
	}
	c.Notices[notice.ID] = notice
	return nil
}

func (c *ResourceCatalog) PageList() []DocumentPage {
	if c == nil {
		return nil
	}
	keys := sortedKeys(c.Pages)
	out := make([]DocumentPage, 0, len(keys))
	for _, key := range keys {
		out = append(out, c.Pages[key])
	}
	return out
}

func (c *ResourceCatalog) AssetList() []AssetResource {
	if c == nil {
		return nil
	}
	keys := sortedKeys(c.Assets)
	out := make([]AssetResource, 0, len(keys))
	for _, key := range keys {
		out = append(out, c.Assets[key])
	}
	return out
}

func (c *ResourceCatalog) NoticeList() []ExternalLinkNotice {
	if c == nil {
		return nil
	}
	keys := sortedKeys(c.Notices)
	out := make([]ExternalLinkNotice, 0, len(keys))
	for _, key := range keys {
		out = append(out, c.Notices[key])
	}
	return out
}

func (c *ResourceCatalog) FailureURLs() []string {
	if c == nil {
		return nil
	}
	urls := make([]string, 0)
	for _, page := range c.Pages {
		if page.IsFailed() {
			urls = append(urls, page.URL)
		}
	}
	for _, asset := range c.Assets {
		if asset.IsFailed() {
			urls = append(urls, asset.URL)
		}
	}
	sort.Strings(urls)
	return urls
}

func (c *ResourceCatalog) IsIncomplete() bool {
	return len(c.FailureURLs()) > 0
}

func (c *ResourceCatalog) Finalize(indexPath string) BundleManifest {
	pages := c.PageList()
	assets := c.AssetList()
	manifest := BuildManifest(c.Job.ID, pages, assets, indexPath)
	for _, notice := range c.NoticeList() {
		manifest.ExternalLinks = append(manifest.ExternalLinks, notice.TargetURL)
	}
	manifest.ExternalLinks = SortURLs(manifest.ExternalLinks)
	manifest.Incomplete = c.IsIncomplete()
	return manifest
}

func sortedKeys[T any](values map[string]T) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func (c *ResourceCatalog) Counts() OutcomeCounts {
	if c == nil {
		return OutcomeCounts{}
	}
	return CountOutcomes(c.PageList(), c.AssetList())
}

func (c *ResourceCatalog) ReplacePage(page DocumentPage) error {
	if _, ok := c.Pages[page.URL]; !ok {
		return errors.New("page is not present")
	}
	return c.AddPage(page)
}

func (c *ResourceCatalog) ReplaceAsset(asset AssetResource) error {
	if _, ok := c.Assets[asset.URL]; !ok {
		return errors.New("asset is not present")
	}
	return c.AddAsset(asset)
}
