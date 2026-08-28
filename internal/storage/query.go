package storage

import (
	"errors"
	"sort"
	"strings"

	"go.etcd.io/bbolt"
	"offlinebundle/internal/domain"
)

type BundleStats struct {
	JobID         string
	Pages         int
	FetchedPages  int
	FailedPages   int
	Assets        int
	FetchedAssets int
	FailedAssets  int
	External      int
	Notices       int
}

func (s *Store) Stats(jobID string) (BundleStats, error) {
	if strings.TrimSpace(jobID) == "" {
		return BundleStats{}, errors.New("job id is required")
	}
	pages, err := s.ListPages(jobID)
	if err != nil {
		return BundleStats{}, err
	}
	assets, err := s.ListAssets(jobID)
	if err != nil {
		return BundleStats{}, err
	}
	notices, err := s.ListExternalLinkNotices(jobID)
	if err != nil {
		return BundleStats{}, err
	}
	counts := domain.CountOutcomes(pages, assets)
	return BundleStats{JobID: jobID, Pages: counts.Pages, FetchedPages: counts.FetchedPages, FailedPages: counts.FailedPages, Assets: counts.Assets, FetchedAssets: counts.FetchedAssets, FailedAssets: counts.FailedAssets, External: counts.ExternalAssets, Notices: len(notices)}, nil
}

func (s BundleStats) Complete() bool {
	return s.FailedPages == 0 && s.FailedAssets == 0
}

func (s BundleStats) Status() domain.JobStatus {
	if s.Complete() {
		return domain.StatusComplete
	}
	return domain.StatusIncomplete
}

func (s BundleStats) Fields() []string {
	return []string{"job=" + s.JobID, "pages=" + formatInt(s.Pages), "fetched_pages=" + formatInt(s.FetchedPages), "failed_pages=" + formatInt(s.FailedPages), "assets=" + formatInt(s.Assets), "fetched_assets=" + formatInt(s.FetchedAssets), "failed_assets=" + formatInt(s.FailedAssets), "external=" + formatInt(s.External), "notices=" + formatInt(s.Notices)}
}

func (s BundleStats) String() string { return strings.Join(s.Fields(), " ") }

func formatInt(value int) string {
	if value == 0 {
		return "0"
	}
	negative := value < 0
	if negative {
		value = -value
	}
	digits := []byte{}
	for value > 0 {
		digits = append(digits, byte('0'+value%10))
		value /= 10
	}
	if negative {
		digits = append(digits, '-')
	}
	for left, right := 0, len(digits)-1; left < right; left, right = left+1, right-1 {
		digits[left], digits[right] = digits[right], digits[left]
	}
	return string(digits)
}

func (s *Store) ListJobIDs() ([]string, error) {
	jobs, err := listJSON[domain.CaptureJob](s, bucketJobs)
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(jobs))
	for _, job := range jobs {
		ids = append(ids, job.ID)
	}
	sort.Strings(ids)
	return ids, nil
}

func (s *Store) DeleteBundle(id string) error {
	if strings.TrimSpace(id) == "" {
		return errors.New("job id is required")
	}
	pages, err := s.ListPages(id)
	if err != nil {
		return err
	}
	assets, err := s.ListAssets(id)
	if err != nil {
		return err
	}
	notices, err := s.ListExternalLinkNotices(id)
	if err != nil {
		return err
	}
	return s.db.Update(func(tx *bbolt.Tx) error {
		if err := tx.Bucket(bucketJobs).Delete([]byte(id)); err != nil {
			return err
		}
		for _, page := range pages {
			if err := tx.Bucket(bucketPages).Delete([]byte(page.ID)); err != nil {
				return err
			}
		}
		for _, asset := range assets {
			if err := tx.Bucket(bucketAssets).Delete([]byte(asset.ID)); err != nil {
				return err
			}
		}
		for _, notice := range notices {
			if err := tx.Bucket(bucketExternalLinks).Delete([]byte(notice.ID)); err != nil {
				return err
			}
		}
		return tx.Bucket(bucketManifests).Delete([]byte(id + "|manifest"))
	})
}
