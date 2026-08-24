package domain

import "testing"

func TestStatusTransitions(t *testing.T) {
	j, err := NewCaptureJob("j1", "https://docs.test/", "out")
	if err != nil {
		t.Fatal(err)
	}
	if err = j.Start(); err != nil {
		t.Fatal(err)
	}
	m := BuildManifest("j1", nil, nil, "out/index.html")
	j.Finish(m)
	if j.Status != StatusComplete || !j.Complete {
		t.Fatalf("unexpected job: %+v", j)
	}
}

func TestSummaryMarksIncomplete(t *testing.T) {
	p := NewDocumentPage("j", "https://docs.test/missing", "missing.html")
	p.MarkFailed(assertError("gone"))
	m := BuildManifest("j", []DocumentPage{p}, nil, "index.html")
	if !m.Incomplete || len(m.FailedPages) != 1 {
		t.Fatalf("unexpected manifest: %+v", m)
	}
}

type assertError string

func (e assertError) Error() string { return string(e) }

func TestPathHelpers(t *testing.T) {
	if PagePath("https://docs.test/") != "index.html" {
		t.Fatal("root path")
	}
	if !SameHost("https://docs.test/a", "https://docs.test/b") {
		t.Fatal("host")
	}
	if SameHost("https://docs.test/a", "https://other.test/b") {
		t.Fatal("different host")
	}
}
