package domain

import (
	"crypto/sha1"
	"encoding/hex"
	"net/url"
	"path"
	"strings"
)

func NormalizeURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	u.Fragment = ""
	if u.Path == "" {
		u.Path = "/"
	}
	return u.String()
}

func SameHost(root, candidate string) bool {
	a, ea := url.Parse(root)
	b, eb := url.Parse(candidate)
	if ea != nil || eb != nil {
		return false
	}
	return strings.EqualFold(a.Host, b.Host)
}

func StableFileName(raw, ext string) string {
	h := sha1.Sum([]byte(raw))
	return hex.EncodeToString(h[:8]) + ext
}

func PagePath(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return StableFileName(raw, ".html")
	}
	p := strings.Trim(u.Path, "/")
	if p == "" {
		return "index.html"
	}
	p = path.Clean(p)
	p = strings.ReplaceAll(p, "/", "_")
	if !strings.HasSuffix(p, ".html") {
		p += ".html"
	}
	return p
}

func AssetPath(raw, kind string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return StableFileName(raw, ".bin")
	}
	ext := path.Ext(u.Path)
	if ext == "" {
		switch kind {
		case "stylesheet":
			ext = ".css"
		case "image":
			ext = ".img"
		default:
			ext = ".bin"
		}
	}
	return StableFileName(raw, ext)
}
