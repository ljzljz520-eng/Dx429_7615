package service

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"

	"offlinebundle/internal/builder"
	"offlinebundle/internal/domain"
	"offlinebundle/internal/fetcher"
	"offlinebundle/internal/report"
	"offlinebundle/internal/storage"
)

type Service struct {
	Store    *storage.Store
	Client   fetcher.HTTPClient
	MaxPages int
}

func New(store *storage.Store, client fetcher.HTTPClient) *Service {
	if client == nil {
		client = &http.Client{}
	}
	return &Service{Store: store, Client: client, MaxPages: 32}
}

type CreateResult struct {
	Job      domain.CaptureJob
	Pages    []domain.DocumentPage
	Assets   []domain.AssetResource
	Notices  []domain.ExternalLinkNotice
	Manifest domain.BundleManifest
}

func (s *Service) CreateBundle(id, root, output string) (CreateResult, error) {
	if s.Store == nil {
		return CreateResult{}, errors.New("store is required")
	}
	job, err := domain.NewCaptureJob(id, root, output)
	if err != nil {
		return CreateResult{}, err
	}
	if err = job.Start(); err != nil {
		return CreateResult{}, err
	}
	plan, err := NewCapturePlan(id, root, output, domain.DefaultCapturePolicy())
	if err != nil {
		return CreateResult{}, err
	}
	plan.Record()
	if err = s.Store.SaveJob(job); err != nil {
		return CreateResult{}, err
	}
	crawler := fetcher.NewCrawler(s.Client, root, id)
	if s.MaxPages > 0 {
		crawler.MaxPages = s.MaxPages
	}
	result, crawlErr := crawler.Crawl()
	plan.Discover(result)
	plan.Fetch(result)
	if crawlErr != nil && len(result.Pages) == 0 {
		job.Status = domain.StatusFailed
		s.Store.SaveJob(job)
		return CreateResult{Job: job}, crawlErr
	}
	for i := range result.Pages {
		if err := s.Store.SavePage(result.Pages[i]); err != nil {
			return CreateResult{}, err
		}
	}
	for i := range result.Assets {
		result.Assets[i].JobID = id
		result.Assets[i].ID = id + "|asset|" + result.Assets[i].URL
		if err := s.Store.SaveAsset(result.Assets[i]); err != nil {
			return CreateResult{}, err
		}
		if body, ok := result.AssetBodies[result.Assets[i].URL]; ok && !result.Assets[i].External {
			if err := builder.New(output).WriteAsset(result.Assets[i], body); err != nil {
				return CreateResult{}, err
			}
		}
	}
	manifest := domain.BuildManifest(id, result.Pages, result.Assets, filepath.Join(output, "index.html"))
	manifest.ExternalLinks = append(manifest.ExternalLinks, result.ExternalLinks...)
	notices := make([]domain.ExternalLinkNotice, 0, len(result.ExternalLinks))
	for _, link := range result.ExternalLinks {
		notice := domain.NewExternalLinkNotice(id, root, link, "external resource")
		notices = append(notices, notice)
		if err := s.Store.SaveExternalLinkNotice(notice); err != nil {
			return CreateResult{}, err
		}
	}
	if err = s.Store.SaveManifest(manifest); err != nil {
		return CreateResult{}, err
	}
	job.Finish(manifest)
	plan.Render(manifest)
	if err = s.Store.SaveJob(job); err != nil {
		return CreateResult{}, err
	}
	b := builder.New(output)
	if err = b.Prepare(); err != nil {
		return CreateResult{}, err
	}
	for _, p := range result.Pages {
		if p.IsFailed() {
			if err = b.WriteFailureNotice(p); err != nil {
				return CreateResult{}, err
			}
			continue
		}
		if body, ok := result.Bodies[p.URL]; ok {
			if err = b.WritePage(p, builder.SanitizeBody(body)); err != nil {
				return CreateResult{}, err
			}
		}
	}
	if err = b.BuildIndex(job, result.Pages, manifest); err != nil {
		return CreateResult{}, err
	}
	if err = b.WriteManifest(manifest); err != nil {
		return CreateResult{}, err
	}
	if err = b.WriteReadme(job, manifest); err != nil {
		return CreateResult{}, err
	}
	return CreateResult{Job: job, Pages: result.Pages, Assets: result.Assets, Notices: notices, Manifest: manifest}, nil
}

func (s *Service) InspectBundle(id string) (string, error) {
	if s.Store == nil {
		return "", errors.New("store is required")
	}
	job, err := s.Store.LoadJob(id)
	if err != nil {
		return "", err
	}
	manifest, err := s.Store.LoadManifest(id + "|manifest")
	if err != nil {
		return "", err
	}
	pages, err := s.Store.ListPages(id)
	if err != nil {
		return "", err
	}
	assets, err := s.Store.ListAssets(id)
	if err != nil {
		return "", err
	}
	notices, err := s.Store.ListExternalLinkNotices(id)
	if err != nil {
		return "", err
	}
	return report.FormatWithNotices(job, manifest, pages, assets, notices), nil
}

func (s *Service) SaveAll(result CreateResult) error {
	if s.Store == nil {
		return errors.New("store is required")
	}
	if err := s.Store.SaveJob(result.Job); err != nil {
		return err
	}
	for _, p := range result.Pages {
		if err := s.Store.SavePage(p); err != nil {
			return err
		}
	}
	for _, a := range result.Assets {
		if err := s.Store.SaveAsset(a); err != nil {
			return err
		}
	}
	for _, n := range result.Notices {
		if err := s.Store.SaveExternalLinkNotice(n); err != nil {
			return err
		}
	}
	return s.Store.SaveManifest(result.Manifest)
}

func (s *Service) ValidateCreate(id, root, output string) error {
	if id == "" || root == "" || output == "" {
		return errors.New("id, root and output are required")
	}
	if len(root) < 8 {
		return fmt.Errorf("root URL is too short")
	}
	return nil
}

func (s *Service) StoreSummary(id string) (domain.CaptureJob, domain.BundleManifest, error) {
	j, e := s.Store.LoadJob(id)
	if e != nil {
		return j, domain.BundleManifest{}, e
	}
	m, e := s.Store.LoadManifest(id + "|manifest")
	return j, m, e
}

type BundleSnapshot struct {
	Job      domain.CaptureJob
	Pages    []domain.DocumentPage
	Assets   []domain.AssetResource
	Notices  []domain.ExternalLinkNotice
	Manifest domain.BundleManifest
}

func (s *Service) LoadSnapshot(id string) (BundleSnapshot, error) {
	if s.Store == nil {
		return BundleSnapshot{}, errors.New("store is required")
	}
	job, err := s.Store.LoadJob(id)
	if err != nil {
		return BundleSnapshot{}, err
	}
	pages, err := s.Store.ListPages(id)
	if err != nil {
		return BundleSnapshot{}, err
	}
	assets, err := s.Store.ListAssets(id)
	if err != nil {
		return BundleSnapshot{}, err
	}
	notices, err := s.Store.ListExternalLinkNotices(id)
	if err != nil {
		return BundleSnapshot{}, err
	}
	manifest, err := s.Store.LoadManifest(id + "|manifest")
	if err != nil {
		return BundleSnapshot{}, err
	}
	return BundleSnapshot{Job: job, Pages: pages, Assets: assets, Notices: notices, Manifest: manifest}, nil
}

func (s *Service) ReconcileSnapshot(snapshot BundleSnapshot) domain.OutcomeCounts {
	domain.NormalizeStatuses(snapshot.Pages, snapshot.Assets)
	return domain.CountOutcomes(snapshot.Pages, snapshot.Assets)
}

func (s *Service) ValidateSnapshot(snapshot BundleSnapshot) error {
	if err := domain.ValidateJob(snapshot.Job); err != nil {
		return err
	}
	if snapshot.Manifest.JobID != snapshot.Job.ID {
		return fmt.Errorf("manifest belongs to %s, expected %s", snapshot.Manifest.JobID, snapshot.Job.ID)
	}
	if snapshot.Manifest.Incomplete != (len(snapshot.Manifest.FailedPages)+len(snapshot.Manifest.FailedAssets) > 0) {
		return errors.New("manifest completeness does not match failures")
	}
	return nil
}

func (s *Service) ExportBundle(id, format string) (string, error) {
	snapshot, err := s.LoadSnapshot(id)
	if err != nil {
		return "", err
	}
	switch format {
	case "json":
		data, err := report.EncodeRows(snapshot.Pages, snapshot.Assets)
		return string(data), err
	case "csv":
		return report.CSVRows(snapshot.Pages, snapshot.Assets), nil
	case "summary":
		return report.FormatWithNotices(snapshot.Job, snapshot.Manifest, snapshot.Pages, snapshot.Assets, snapshot.Notices), nil
	default:
		return "", fmt.Errorf("unsupported export format %q", format)
	}
}
