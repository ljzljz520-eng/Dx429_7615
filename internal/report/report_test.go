package report

import (
	"offlinebundle/internal/domain"
	"strings"
	"testing"
)

func TestReportIncludesFailedPages(t *testing.T) {
	j, _ := domain.NewCaptureJob("j", "https://docs.test/", "out")
	j.Status = domain.StatusIncomplete
	p := domain.NewDocumentPage("j", "https://docs.test/missing", "missing.html")
	p.MarkFailed(testErr("404"))
	m := domain.BuildManifest("j", []domain.DocumentPage{p}, nil, "index.html")
	text := Format(j, m, []domain.DocumentPage{p}, nil)
	if !strings.Contains(text, "failed pages") || !strings.Contains(text, "missing") {
		t.Fatal(text)
	}
}
func TestReportSummary(t *testing.T) {
	m := domain.BundleManifest{PageCount: 2, AssetCount: 3, FailedPages: []string{"x"}}
	if !strings.Contains(SummaryLine(m), "2 pages") {
		t.Fatal(SummaryLine(m))
	}
}

type testErr string

func (e testErr) Error() string { return string(e) }
