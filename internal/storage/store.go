package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"

	"go.etcd.io/bbolt"
	"offlinebundle/internal/domain"
)

var bucketJobs = []byte("capture_jobs")
var bucketPages = []byte("document_pages")
var bucketAssets = []byte("asset_resources")
var bucketManifests = []byte("bundle_manifests")
var bucketExternalLinks = []byte("external_link_notices")

type Store struct {
	db   *bbolt.DB
	path string
}

func Open(path string) (*Store, error) {
	if path == "" {
		return nil, errors.New("storage path is required")
	}
	db, err := bbolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}
	s := &Store{db: db, path: path}
	if err := s.initBuckets(); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) initBuckets() error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		for _, name := range [][]byte{bucketJobs, bucketPages, bucketAssets, bucketManifests, bucketExternalLinks} {
			if _, err := tx.CreateBucketIfNotExists(name); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	err := s.db.Close()
	s.db = nil
	return err
}

func (s *Store) Path() string { return s.path }

func putJSON(tx *bbolt.Tx, bucket []byte, key string, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return tx.Bucket(bucket).Put([]byte(key), data)
}

func getJSON(tx *bbolt.Tx, bucket []byte, key string, out any) error {
	data := tx.Bucket(bucket).Get([]byte(key))
	if data == nil {
		return os.ErrNotExist
	}
	return json.Unmarshal(data, out)
}

func (s *Store) SaveJob(job domain.CaptureJob) error {
	return s.db.Update(func(tx *bbolt.Tx) error { return putJSON(tx, bucketJobs, job.ID, job) })
}

func (s *Store) LoadJob(id string) (domain.CaptureJob, error) {
	var j domain.CaptureJob
	err := s.db.View(func(tx *bbolt.Tx) error { return getJSON(tx, bucketJobs, id, &j) })
	return j, err
}

func (s *Store) SavePage(page domain.DocumentPage) error {
	return s.db.Update(func(tx *bbolt.Tx) error { return putJSON(tx, bucketPages, page.ID, page) })
}

func (s *Store) LoadPage(id string) (domain.DocumentPage, error) {
	var p domain.DocumentPage
	err := s.db.View(func(tx *bbolt.Tx) error { return getJSON(tx, bucketPages, id, &p) })
	return p, err
}

func (s *Store) SaveAsset(asset domain.AssetResource) error {
	return s.db.Update(func(tx *bbolt.Tx) error { return putJSON(tx, bucketAssets, asset.ID, asset) })
}

func (s *Store) LoadAsset(id string) (domain.AssetResource, error) {
	var a domain.AssetResource
	err := s.db.View(func(tx *bbolt.Tx) error { return getJSON(tx, bucketAssets, id, &a) })
	return a, err
}

func (s *Store) SaveManifest(manifest domain.BundleManifest) error {
	return s.db.Update(func(tx *bbolt.Tx) error { return putJSON(tx, bucketManifests, manifest.ID, manifest) })
}

func (s *Store) SaveExternalLinkNotice(notice domain.ExternalLinkNotice) error {
	if s == nil || s.db == nil {
		return errors.New("store closed")
	}
	return s.db.Update(func(tx *bbolt.Tx) error { return putJSON(tx, bucketExternalLinks, notice.ID, notice) })
}

func (s *Store) LoadExternalLinkNotice(id string) (domain.ExternalLinkNotice, error) {
	var notice domain.ExternalLinkNotice
	if s == nil || s.db == nil {
		return notice, errors.New("store closed")
	}
	err := s.db.View(func(tx *bbolt.Tx) error { return getJSON(tx, bucketExternalLinks, id, &notice) })
	return notice, err
}

func (s *Store) ListExternalLinkNotices(jobID string) ([]domain.ExternalLinkNotice, error) {
	all, err := listJSON[domain.ExternalLinkNotice](s, bucketExternalLinks)
	if err != nil {
		return nil, err
	}
	out := make([]domain.ExternalLinkNotice, 0)
	for _, notice := range all {
		if notice.JobID == jobID {
			out = append(out, notice)
		}
	}
	return out, nil
}

func (s *Store) LoadManifest(id string) (domain.BundleManifest, error) {
	var m domain.BundleManifest
	err := s.db.View(func(tx *bbolt.Tx) error { return getJSON(tx, bucketManifests, id, &m) })
	return m, err
}

func listJSON[T any](s *Store, bucket []byte) ([]T, error) {
	var out []T
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucket)
		if b == nil {
			return fmt.Errorf("bucket missing")
		}
		keys := make([]string, 0)
		err := b.ForEach(func(k, v []byte) error { keys = append(keys, string(k)); return nil })
		if err != nil {
			return err
		}
		sort.Strings(keys)
		for _, key := range keys {
			var item T
			if err := getJSON(tx, bucket, key, &item); err != nil {
				return err
			}
			out = append(out, item)
		}
		return nil
	})
	return out, err
}

func (s *Store) ListPages(jobID string) ([]domain.DocumentPage, error) {
	all, err := listJSON[domain.DocumentPage](s, bucketPages)
	if err != nil {
		return nil, err
	}
	out := make([]domain.DocumentPage, 0)
	for _, p := range all {
		if p.JobID == jobID {
			out = append(out, p)
		}
	}
	return out, nil
}

func (s *Store) ListAssets(jobID string) ([]domain.AssetResource, error) {
	all, err := listJSON[domain.AssetResource](s, bucketAssets)
	if err != nil {
		return nil, err
	}
	out := make([]domain.AssetResource, 0)
	for _, a := range all {
		if a.JobID == jobID {
			out = append(out, a)
		}
	}
	return out, nil
}

func (s *Store) DeleteJob(id string) error {
	return s.db.Update(func(tx *bbolt.Tx) error { return tx.Bucket(bucketJobs).Delete([]byte(id)) })
}

func (s *Store) Health() error {
	if s == nil || s.db == nil {
		return errors.New("store closed")
	}
	return s.db.View(func(*bbolt.Tx) error { return nil })
}
