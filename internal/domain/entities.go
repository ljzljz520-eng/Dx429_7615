package domain

import (
	"errors"
	"fmt"
	"strings"
)

type JobStatus string

const (
	StatusPending    JobStatus = "pending"
	StatusRunning    JobStatus = "running"
	StatusComplete   JobStatus = "complete"
	StatusIncomplete JobStatus = "incomplete"
	StatusFailed     JobStatus = "failed"
)

type PageStatus string

const (
	PagePending PageStatus = "pending"
	PageFetched PageStatus = "fetched"
	PageFailed  PageStatus = "failed"
)

type AssetStatus string

const (
	AssetPending  AssetStatus = "pending"
	AssetFetched  AssetStatus = "fetched"
	AssetFailed   AssetStatus = "failed"
	AssetExternal AssetStatus = "external"
)

type CaptureJob struct {
	ID        string    `json:"id"`
	RootURL   string    `json:"root_url"`
	OutputDir string    `json:"output_dir"`
	Status    JobStatus `json:"status"`
	CreatedAt string    `json:"created_at"`
	Complete  bool      `json:"complete"`
	Summary   string    `json:"summary"`
}

type DocumentPage struct {
	ID     string     `json:"id"`
	JobID  string     `json:"job_id"`
	URL    string     `json:"url"`
	Path   string     `json:"path"`
	Title  string     `json:"title"`
	Status PageStatus `json:"status"`
	Error  string     `json:"error"`
	Links  []string   `json:"links"`
}

type AssetResource struct {
	ID       string      `json:"id"`
	JobID    string      `json:"job_id"`
	URL      string      `json:"url"`
	Path     string      `json:"path"`
	Kind     string      `json:"kind"`
	Status   AssetStatus `json:"status"`
	Error    string      `json:"error"`
	External bool        `json:"external"`
}

type ExternalLinkNotice struct {
	ID        string `json:"id"`
	JobID     string `json:"job_id"`
	SourceURL string `json:"source_url"`
	TargetURL string `json:"target_url"`
	Label     string `json:"label"`
}

func NewExternalLinkNotice(jobID, sourceURL, targetURL, label string) ExternalLinkNotice {
	return ExternalLinkNotice{ID: jobID + "|external|" + sourceURL + "|" + targetURL, JobID: jobID, SourceURL: sourceURL, TargetURL: targetURL, Label: label}
}

func (n ExternalLinkNotice) Display() string {
	if n.Label == "" {
		return n.TargetURL
	}
	return n.Label + " -> " + n.TargetURL
}

type BundleManifest struct {
	ID            string   `json:"id"`
	JobID         string   `json:"job_id"`
	IndexPath     string   `json:"index_path"`
	PageCount     int      `json:"page_count"`
	AssetCount    int      `json:"asset_count"`
	FailedPages   []string `json:"failed_pages"`
	FailedAssets  []string `json:"failed_assets"`
	Incomplete    bool     `json:"incomplete"`
	ExternalLinks []string `json:"external_links"`
}

func NewCaptureJob(id, root, output string) (CaptureJob, error) {
	if strings.TrimSpace(id) == "" || strings.TrimSpace(root) == "" || strings.TrimSpace(output) == "" {
		return CaptureJob{}, errors.New("job id, root URL, and output directory are required")
	}
	return CaptureJob{ID: id, RootURL: root, OutputDir: output, Status: StatusPending, CreatedAt: "deterministic", Complete: false}, nil
}

func (j *CaptureJob) Start() error {
	if j.Status != StatusPending {
		return fmt.Errorf("cannot start job in %s state", j.Status)
	}
	j.Status = StatusRunning
	return nil
}

func (j *CaptureJob) Finish(manifest BundleManifest) {
	j.Complete = !manifest.Incomplete
	if manifest.Incomplete {
		j.Status = StatusIncomplete
	} else {
		j.Status = StatusComplete
	}
	j.Summary = fmt.Sprintf("pages=%d assets=%d failed_pages=%d failed_assets=%d", manifest.PageCount, manifest.AssetCount, len(manifest.FailedPages), len(manifest.FailedAssets))
}

func (j CaptureJob) IsTerminal() bool {
	return j.Status == StatusComplete || j.Status == StatusIncomplete || j.Status == StatusFailed
}

func NewDocumentPage(jobID, rawURL, path string) DocumentPage {
	return DocumentPage{ID: jobID + "|page|" + rawURL, JobID: jobID, URL: rawURL, Path: path, Status: PagePending, Links: []string{}}
}

func (p *DocumentPage) MarkFetched(title string, links []string) {
	p.Status = PageFetched
	p.Title = title
	p.Links = append([]string(nil), links...)
	p.Error = ""
}

func (p *DocumentPage) MarkFailed(err error) {
	p.Status = PageFailed
	if err != nil {
		p.Error = err.Error()
	}
}

func (p DocumentPage) IsFailed() bool { return p.Status == PageFailed }

func NewAssetResource(jobID, rawURL, path, kind string) AssetResource {
	return AssetResource{ID: jobID + "|asset|" + rawURL, JobID: jobID, URL: rawURL, Path: path, Kind: kind, Status: AssetPending}
}

func (a *AssetResource) MarkFetched() { a.Status = AssetFetched; a.Error = "" }

func (a *AssetResource) MarkFailed(err error) {
	a.Status = AssetFailed
	if err != nil {
		a.Error = err.Error()
	}
}

func (a *AssetResource) MarkExternal() { a.Status = AssetExternal; a.External = true }

func (a AssetResource) IsFailed() bool { return a.Status == AssetFailed }

func BuildManifest(jobID string, pages []DocumentPage, assets []AssetResource, indexPath string) BundleManifest {
	m := BundleManifest{ID: jobID + "|manifest", JobID: jobID, IndexPath: indexPath, PageCount: len(pages), AssetCount: len(assets), FailedPages: []string{}, FailedAssets: []string{}, ExternalLinks: []string{}}
	for _, p := range pages {
		if p.IsFailed() {
			m.FailedPages = append(m.FailedPages, p.URL)
		}
	}
	for _, a := range assets {
		if a.IsFailed() {
			m.FailedAssets = append(m.FailedAssets, a.URL)
		}
		if a.External {
			m.ExternalLinks = append(m.ExternalLinks, a.URL)
		}
	}
	m.Incomplete = len(m.FailedPages) > 0 || len(m.FailedAssets) > 0
	return m
}

func ValidateJob(j CaptureJob) error {
	if j.ID == "" {
		return errors.New("missing job id")
	}
	if j.RootURL == "" {
		return errors.New("missing root URL")
	}
	return nil
}

func PageDisplayName(p DocumentPage) string {
	if p.Title != "" {
		return p.Title
	}
	return p.URL
}

func StatusLabel(status JobStatus) string {
	switch status {
	case StatusComplete:
		return "COMPLETE"
	case StatusIncomplete:
		return "INCOMPLETE"
	case StatusRunning:
		return "RUNNING"
	default:
		return strings.ToUpper(string(status))
	}
}
