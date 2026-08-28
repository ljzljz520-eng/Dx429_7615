package fetcher

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"offlinebundle/internal/domain"
)

type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

type Crawler struct {
	Client   HTTPClient
	Source   ResourceSource
	MaxPages int
	Root     string
	JobID    string
}

type Result struct {
	Pages         []domain.DocumentPage
	Assets        []domain.AssetResource
	ExternalLinks []string
	Bodies        map[string][]byte
	AssetBodies   map[string][]byte
}

func NewCrawler(client HTTPClient, root, jobID string) *Crawler {
	return &Crawler{Client: client, Source: NewHTTPSource(client), MaxPages: 32, Root: domain.NormalizeURL(root), JobID: jobID}
}

func (c *Crawler) Crawl() (Result, error) {
	if c.Client == nil {
		return Result{}, errors.New("HTTP client is required")
	}
	rootURL, err := url.Parse(c.Root)
	if err != nil {
		return Result{}, err
	}
	queue := []string{c.Root}
	seen := map[string]bool{}
	result := Result{Pages: []domain.DocumentPage{}, Assets: []domain.AssetResource{}, ExternalLinks: []string{}, Bodies: map[string][]byte{}, AssetBodies: map[string][]byte{}}
	for len(queue) > 0 && len(result.Pages) < c.MaxPages {
		current := queue[0]
		queue = queue[1:]
		if seen[current] {
			continue
		}
		seen[current] = true
		page := domain.NewDocumentPage(c.JobID, current, domain.PagePath(current))
		body, status, fetchErr := c.fetch(current)
		if fetchErr != nil || status >= 400 {
			continue
		}
		title, links, assets := ParseDocument(current, body)
		page.MarkFetched(title, links)
		result.Pages = append(result.Pages, page)
		result.Bodies[current] = append([]byte(nil), body...)
		for _, link := range links {
			if isAssetReference(link) {
				kind := classifyAsset(link)
				asset := domain.NewAssetResource(c.JobID, link, domain.AssetPath(link, kind), kind)
				if !domain.SameHost(rootURL.String(), link) {
					asset.MarkExternal()
					result.ExternalLinks = appendUnique(result.ExternalLinks, link)
				} else {
					assetBody, assetStatus, assetErr := c.fetch(link)
					if assetErr != nil || assetStatus >= 400 {
						asset.MarkFailed(fetchError(assetStatus, assetErr))
					} else {
						asset.MarkFetched()
						result.AssetBodies[link] = append([]byte(nil), assetBody...)
					}
				}
				result.Assets = append(result.Assets, asset)
				continue
			}
			if domain.SameHost(rootURL.String(), link) {
				normalized := domain.NormalizeURL(link)
				if !seen[normalized] {
					queue = append(queue, normalized)
				}
			} else {
				result.ExternalLinks = appendUnique(result.ExternalLinks, link)
			}
		}
		for _, asset := range assets {
			if !domain.SameHost(rootURL.String(), asset.URL) {
				asset.MarkExternal()
				result.ExternalLinks = appendUnique(result.ExternalLinks, asset.URL)
			}
			result.Assets = append(result.Assets, asset)
		}
	}
	if len(result.Pages) == 0 {
		return result, errors.New("root page could not be fetched")
	}
	return result, nil
}

func (c *Crawler) fetch(raw string) ([]byte, int, error) {
	if c.Source != nil {
		response, err := c.Source.Fetch(raw)
		return response.Body, response.StatusCode, err
	}
	req, err := http.NewRequest(http.MethodGet, raw, nil)
	if err != nil {
		return nil, 0, err
	}
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	body, readErr := io.ReadAll(resp.Body)
	return body, resp.StatusCode, readErr
}

func ParseDocument(base string, body []byte) (string, []string, []domain.AssetResource) {
	text := string(body)
	title := extractTitle(text)
	links := []string{}
	assets := []domain.AssetResource{}
	for _, raw := range extractAttribute(text, "href") {
		if resolved := resolveURL(base, raw); resolved != "" {
			links = appendUnique(links, resolved)
		}
	}
	for _, raw := range extractAttribute(text, "src") {
		if resolved := resolveURL(base, raw); resolved != "" {
			kind := classifyAsset(raw)
			assets = append(assets, domain.NewAssetResource("", resolved, domain.AssetPath(resolved, kind), kind))
		}
	}
	for _, raw := range extractAttribute(text, "data-file") {
		if resolved := resolveURL(base, raw); resolved != "" {
			assets = append(assets, domain.NewAssetResource("", resolved, domain.AssetPath(resolved, "example"), "example"))
		}
	}
	return title, links, assets
}

func extractTitle(body string) string {
	re := regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title>`)
	m := re.FindStringSubmatch(body)
	if len(m) > 1 {
		return strings.TrimSpace(regexp.MustCompile(`<[^>]+>`).ReplaceAllString(m[1], ""))
	}
	return ""
}

func extractAttribute(body, attr string) []string {
	re := regexp.MustCompile(`(?is)` + attr + `\s*=\s*["']([^"']+)["']`)
	matches := re.FindAllStringSubmatch(body, -1)
	out := make([]string, 0, len(matches))
	for _, m := range matches {
		if len(m) > 1 {
			out = append(out, m[1])
		}
	}
	return out
}

func resolveURL(base, raw string) string {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Scheme == "mailto" || u.Scheme == "javascript" {
		return ""
	}
	b, err := url.Parse(base)
	if err != nil {
		return ""
	}
	return b.ResolveReference(u).String()
}

func classifyAsset(raw string) string {
	lower := strings.ToLower(raw)
	if strings.HasSuffix(lower, ".css") {
		return "stylesheet"
	}
	if strings.HasSuffix(lower, ".png") || strings.HasSuffix(lower, ".jpg") || strings.HasSuffix(lower, ".jpeg") || strings.HasSuffix(lower, ".gif") || strings.HasSuffix(lower, ".svg") {
		return "image"
	}
	return "example"
}

func isAssetReference(raw string) bool {
	lower := strings.ToLower(raw)
	return strings.HasSuffix(lower, ".css") || strings.HasSuffix(lower, ".png") || strings.HasSuffix(lower, ".jpg") || strings.HasSuffix(lower, ".jpeg") || strings.HasSuffix(lower, ".gif") || strings.HasSuffix(lower, ".svg") || strings.HasSuffix(lower, ".zip") || strings.HasSuffix(lower, ".go") || strings.HasSuffix(lower, ".json")
}

func fetchError(status int, err error) error {
	if err != nil {
		return err
	}
	if status >= 400 {
		return fmt.Errorf("resource returned HTTP %d", status)
	}
	return errors.New("resource fetch failed")
}

func appendUnique(values []string, candidate string) []string {
	for _, v := range values {
		if v == candidate {
			return values
		}
	}
	return append(values, candidate)
}

func SummarizeResult(result Result) string {
	return fmt.Sprintf("pages=%d assets=%d external=%d", len(result.Pages), len(result.Assets), len(result.ExternalLinks))
}
