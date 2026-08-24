package fetcher

import (
	"net/url"
	"sort"
	"strings"

	"offlinebundle/internal/domain"
)

type LinkGraph struct {
	Edges map[string][]string
	Roots []string
}

func NewLinkGraph() *LinkGraph { return &LinkGraph{Edges: map[string][]string{}, Roots: []string{}} }

func (g *LinkGraph) AddPage(page domain.DocumentPage) {
	if g == nil || page.URL == "" {
		return
	}
	links := append([]string(nil), page.Links...)
	sort.Strings(links)
	g.Edges[page.URL] = uniqueStrings(links)
	if len(g.Roots) == 0 {
		g.Roots = append(g.Roots, page.URL)
	}
}

func (g *LinkGraph) Neighbors(raw string) []string {
	if g == nil {
		return nil
	}
	return append([]string(nil), g.Edges[raw]...)
}

func (g *LinkGraph) Reachable(root string) []string {
	seen := map[string]bool{}
	queue := []string{root}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if seen[current] {
			continue
		}
		seen[current] = true
		for _, next := range g.Edges[current] {
			if !seen[next] {
				queue = append(queue, next)
			}
		}
	}
	out := make([]string, 0, len(seen))
	for value := range seen {
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func (g *LinkGraph) External(root string) []string {
	if g == nil {
		return nil
	}
	external := []string{}
	for _, links := range g.Edges {
		for _, link := range links {
			if !domain.SameHost(root, link) {
				external = append(external, link)
			}
		}
	}
	return uniqueStrings(external)
}

func IsFragmentOnly(raw string) bool {
	u, err := url.Parse(raw)
	return err == nil && u.Fragment != "" && u.Path == "" && u.Host == "" && u.Scheme == ""
}

func IsSameDocument(base, candidate string) bool {
	if IsFragmentOnly(candidate) {
		return true
	}
	baseURL, baseErr := url.Parse(base)
	candidateURL, candidateErr := url.Parse(candidate)
	if baseErr != nil || candidateErr != nil {
		return false
	}
	baseURL.Fragment = ""
	candidateURL.Fragment = ""
	return baseURL.String() == candidateURL.String()
}

func uniqueStrings(values []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}
