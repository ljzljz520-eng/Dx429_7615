package domain

import (
	"errors"
	"net/url"
	"path"
	"strings"
)

type CapturePolicy struct {
	MaxPages       int
	MaxAssets      int
	AllowFragments bool
	KeepQuery      bool
}

func DefaultCapturePolicy() CapturePolicy {
	return CapturePolicy{MaxPages: 32, MaxAssets: 128, AllowFragments: false, KeepQuery: true}
}

func (p CapturePolicy) Validate() error {
	if p.MaxPages < 1 {
		return errors.New("max pages must be positive")
	}
	if p.MaxAssets < 1 {
		return errors.New("max assets must be positive")
	}
	return nil
}

func (p CapturePolicy) AcceptPage(count int) bool {
	return count < p.MaxPages
}

func (p CapturePolicy) AcceptAsset(count int) bool {
	return count < p.MaxAssets
}

func (p CapturePolicy) Normalize(raw string) string {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return ""
	}
	if !p.AllowFragments {
		u.Fragment = ""
	}
	if !p.KeepQuery {
		u.RawQuery = ""
	}
	if u.Path == "" {
		u.Path = "/"
	}
	u.Path = path.Clean("/" + u.Path)
	if u.Path == "." || u.Path == "" {
		u.Path = "/"
	}
	return u.String()
}

func (p CapturePolicy) AllowsURL(root, candidate string) bool {
	if root == "" || candidate == "" {
		return false
	}
	return SameHost(root, candidate)
}

func (p CapturePolicy) Classify(raw string) string {
	lower := strings.ToLower(raw)
	switch {
	case strings.HasSuffix(lower, ".css"):
		return "stylesheet"
	case strings.HasSuffix(lower, ".png"), strings.HasSuffix(lower, ".jpg"), strings.HasSuffix(lower, ".jpeg"), strings.HasSuffix(lower, ".gif"), strings.HasSuffix(lower, ".svg"):
		return "image"
	case strings.HasSuffix(lower, ".go"), strings.HasSuffix(lower, ".json"), strings.HasSuffix(lower, ".zip"), strings.HasSuffix(lower, ".txt"):
		return "example"
	default:
		return "page"
	}
}

func IsSafeRelativePath(candidate string) bool {
	if candidate == "" || strings.HasPrefix(candidate, "/") || strings.Contains(candidate, "..") {
		return false
	}
	clean := path.Clean(candidate)
	return clean == candidate && clean != "."
}

func ValidateResourcePath(candidate string) error {
	if !IsSafeRelativePath(candidate) {
		return errors.New("unsafe relative resource path")
	}
	return nil
}

func IsPageStatusTerminal(status PageStatus) bool {
	return status == PageFetched || status == PageFailed
}

func IsAssetStatusTerminal(status AssetStatus) bool {
	return status == AssetFetched || status == AssetFailed || status == AssetExternal
}

func TransitionStatus(current JobStatus, next JobStatus) error {
	allowed := map[JobStatus][]JobStatus{
		StatusPending: {StatusRunning, StatusFailed},
		StatusRunning: {StatusComplete, StatusIncomplete, StatusFailed},
	}
	for _, candidate := range allowed[current] {
		if candidate == next {
			return nil
		}
	}
	return errors.New("invalid job status transition")
}

func StableEntityID(jobID, kind, value string) string {
	return strings.Join([]string{jobID, kind, NormalizeURL(value)}, "|")
}
