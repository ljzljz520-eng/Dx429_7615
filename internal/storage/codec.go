package storage

import (
	"encoding/json"
	"offlinebundle/internal/domain"
)

func EncodeJob(j domain.CaptureJob) ([]byte, error) { return json.Marshal(j) }
func DecodeJob(data []byte) (domain.CaptureJob, error) {
	var j domain.CaptureJob
	err := json.Unmarshal(data, &j)
	return j, err
}
func EncodePage(p domain.DocumentPage) ([]byte, error) { return json.Marshal(p) }
func DecodePage(data []byte) (domain.DocumentPage, error) {
	var p domain.DocumentPage
	err := json.Unmarshal(data, &p)
	return p, err
}
func EncodeAsset(a domain.AssetResource) ([]byte, error) { return json.Marshal(a) }
func DecodeAsset(data []byte) (domain.AssetResource, error) {
	var a domain.AssetResource
	err := json.Unmarshal(data, &a)
	return a, err
}
func EncodeManifest(m domain.BundleManifest) ([]byte, error) { return json.Marshal(m) }
func DecodeManifest(data []byte) (domain.BundleManifest, error) {
	var m domain.BundleManifest
	err := json.Unmarshal(data, &m)
	return m, err
}

func EncodeExternalLinkNotice(n domain.ExternalLinkNotice) ([]byte, error) { return json.Marshal(n) }

func DecodeExternalLinkNotice(data []byte) (domain.ExternalLinkNotice, error) {
	var n domain.ExternalLinkNotice
	err := json.Unmarshal(data, &n)
	return n, err
}
