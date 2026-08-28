package storage

import (
	"errors"
	"go.etcd.io/bbolt"
	"offlinebundle/internal/domain"
)

func (s *Store) SaveBundle(job domain.CaptureJob, pages []domain.DocumentPage, assets []domain.AssetResource, manifest domain.BundleManifest) error {
	return s.SaveBundleAtomic(job, pages, assets, manifest)
}

func (s *Store) SaveBundleAtomic(job domain.CaptureJob, pages []domain.DocumentPage, assets []domain.AssetResource, manifest domain.BundleManifest) error {
	if s == nil || s.db == nil {
		return errors.New("store closed")
	}
	return s.db.Update(func(tx *bbolt.Tx) error {
		if err := putJSON(tx, bucketJobs, job.ID, job); err != nil {
			return err
		}
		for _, p := range pages {
			if err := putJSON(tx, bucketPages, p.ID, p); err != nil {
				return err
			}
		}
		for _, a := range assets {
			if err := putJSON(tx, bucketAssets, a.ID, a); err != nil {
				return err
			}
		}
		return putJSON(tx, bucketManifests, manifest.ID, manifest)
	})
}

func (s *Store) CountForJob(jobID string) (int, int, error) {
	p, e := s.ListPages(jobID)
	if e != nil {
		return 0, 0, e
	}
	a, e := s.ListAssets(jobID)
	return len(p), len(a), e
}
