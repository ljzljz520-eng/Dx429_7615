package fetcher

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCrawlerCollectsSameDomain(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(`<title>Home</title><a href="/guide">Guide</a><link href="/site.css">`))
			return
		}
		w.Write([]byte(`<title>Guide</title>`))
	}))
	defer srv.Close()
	c := NewCrawler(srv.Client(), srv.URL+"/", "j")
	r, err := c.Crawl()
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Pages) != 2 || len(r.Assets) != 1 {
		t.Fatalf("result: %+v", r)
	}
}

func TestExternalLinksAreRetained(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<a href="https://external.test/docs">External</a>`))
	}))
	defer srv.Close()
	c := NewCrawler(srv.Client(), srv.URL+"/", "j")
	r, err := c.Crawl()
	if err != nil {
		t.Fatal(err)
	}
	if len(r.ExternalLinks) != 1 {
		t.Fatalf("external: %+v", r.ExternalLinks)
	}
}

func TestURLSetDeterministic(t *testing.T) {
	s := NewURLSet()
	s.Add("https://docs.test/a#x")
	s.Add("https://docs.test/a")
	if s.Len() != 1 || !s.Has("https://docs.test/a") {
		t.Fatal(s.Values())
	}
}
