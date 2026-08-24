package fetcher

import (
	"net/http"
	"time"
)

func DeterministicClient() *http.Client {
	return &http.Client{Timeout: 3 * time.Second, CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) > 4 {
			return http.ErrUseLastResponse
		}
		return nil
	}}
}

func IsSuccessful(status int) bool { return status >= 200 && status < 300 }

func RetryableStatus(status int) bool { return status == 408 || status == 429 || status >= 500 }

func NewRequest(raw string) (*http.Request, error) { return http.NewRequest(http.MethodGet, raw, nil) }
