package service

import (
	"net/http"
	"offlinebundle/internal/fetcher"
	"time"
)

type Options struct {
	Client         fetcher.HTTPClient
	MaxPages       int
	RequestTimeout time.Duration
}

func DefaultOptions() Options {
	return Options{Client: &http.Client{Timeout: 5 * time.Second}, MaxPages: 32, RequestTimeout: 5 * time.Second}
}
func (o Options) WithClient(client fetcher.HTTPClient) Options { o.Client = client; return o }
func (o Options) WithMaxPages(limit int) Options {
	if limit > 0 {
		o.MaxPages = limit
	}
	return o
}
func NewWithOptions(store interface{ SaveJob(any) error }, options Options) *Service {
	return &Service{MaxPages: options.MaxPages, Client: options.Client}
}
