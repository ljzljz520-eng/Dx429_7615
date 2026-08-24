package fetcher

import (
	"net/url"
	"sort"
	"strings"

	"offlinebundle/internal/domain"
)

type URLSet struct{ values map[string]struct{} }

func NewURLSet() *URLSet { return &URLSet{values: map[string]struct{}{}} }
func (s *URLSet) Add(raw string) bool {
	normalized := normalize(raw)
	if normalized == "" {
		return false
	}
	if _, ok := s.values[normalized]; ok {
		return false
	}
	s.values[normalized] = struct{}{}
	return true
}
func (s *URLSet) Has(raw string) bool { _, ok := s.values[normalize(raw)]; return ok }
func (s *URLSet) Values() []string {
	out := make([]string, 0, len(s.values))
	for v := range s.values {
		out = append(out, v)
	}
	sort.Strings(out)
	return out
}
func (s *URLSet) Len() int { return len(s.values) }
func normalize(raw string) string {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return ""
	}
	u.Fragment = ""
	if u.Path == "" {
		u.Path = "/"
	}
	return u.String()
}
func FilterSameHost(root string, candidates []string) ([]string, []string) {
	same := []string{}
	external := []string{}
	for _, candidate := range candidates {
		if domain.SameHost(root, candidate) {
			same = append(same, domain.NormalizeURL(candidate))
		} else {
			external = append(external, domain.NormalizeURL(candidate))
		}
	}
	sort.Strings(same)
	sort.Strings(external)
	return same, external
}
func IsDocumentURL(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil {
		return false
	}
	return u.Scheme == "http" || u.Scheme == "https"
}
