package report

import (
	"encoding/json"
	"offlinebundle/internal/domain"
)

type Snapshot struct {
	Job      domain.CaptureJob     `json:"job"`
	Manifest domain.BundleManifest `json:"manifest"`
}

func EncodeSnapshot(job domain.CaptureJob, manifest domain.BundleManifest) ([]byte, error) {
	return json.Marshal(Snapshot{job, manifest})
}
func DecodeSnapshot(data []byte) (Snapshot, error) {
	var s Snapshot
	err := json.Unmarshal(data, &s)
	return s, err
}
