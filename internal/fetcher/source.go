package fetcher

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
)

type ResourceSource interface {
	Fetch(rawURL string) (Response, error)
}

type Response struct {
	URL        string
	StatusCode int
	Body       []byte
	Headers    map[string]string
}

func (r Response) Successful() bool {
	return r.StatusCode >= http.StatusOK && r.StatusCode < http.StatusMultipleChoices
}

func (r Response) ContentType() string {
	if r.Headers == nil {
		return ""
	}
	return r.Headers["Content-Type"]
}

type HTTPSource struct {
	Client HTTPClient
}

func NewHTTPSource(client HTTPClient) *HTTPSource {
	if client == nil {
		client = DeterministicClient()
	}
	return &HTTPSource{Client: client}
}

func (s *HTTPSource) Fetch(rawURL string) (Response, error) {
	if s == nil || s.Client == nil {
		return Response{}, errors.New("HTTP source is not configured")
	}
	req, err := NewRequest(rawURL)
	if err != nil {
		return Response{}, err
	}
	resp, err := s.Client.Do(req)
	if err != nil {
		return Response{}, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Response{URL: rawURL, StatusCode: resp.StatusCode}, err
	}
	headers := map[string]string{}
	for key, values := range resp.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}
	return Response{URL: rawURL, StatusCode: resp.StatusCode, Body: body, Headers: headers}, nil
}

type MapSource struct {
	Responses map[string]Response
	Failures  map[string]error
}

func NewMapSource() *MapSource {
	return &MapSource{Responses: map[string]Response{}, Failures: map[string]error{}}
}

func (s *MapSource) Add(rawURL string, status int, body string) {
	if s.Responses == nil {
		s.Responses = map[string]Response{}
	}
	s.Responses[rawURL] = Response{URL: rawURL, StatusCode: status, Body: []byte(body), Headers: map[string]string{"Content-Type": "text/html"}}
}

func (s *MapSource) Fail(rawURL string, err error) {
	if s.Failures == nil {
		s.Failures = map[string]error{}
	}
	if err == nil {
		err = errors.New("configured source failure")
	}
	s.Failures[rawURL] = err
}

func (s *MapSource) Fetch(rawURL string) (Response, error) {
	if s == nil {
		return Response{}, errors.New("map source is nil")
	}
	if err, ok := s.Failures[rawURL]; ok {
		return Response{URL: rawURL, StatusCode: http.StatusServiceUnavailable}, err
	}
	response, ok := s.Responses[rawURL]
	if !ok {
		return Response{URL: rawURL, StatusCode: http.StatusNotFound}, fmt.Errorf("resource %s is not configured", rawURL)
	}
	response.Body = append([]byte(nil), response.Body...)
	return response, nil
}

func FetchAll(source ResourceSource, urls []string) map[string]Response {
	out := map[string]Response{}
	ordered := append([]string(nil), urls...)
	sort.Strings(ordered)
	for _, rawURL := range ordered {
		if strings.TrimSpace(rawURL) == "" {
			continue
		}
		response, err := source.Fetch(rawURL)
		if err != nil {
			response.StatusCode = http.StatusBadGateway
			response.URL = rawURL
		}
		out[rawURL] = response
	}
	return out
}

func ResponseError(response Response, err error) error {
	if err != nil {
		return err
	}
	if !response.Successful() {
		return fmt.Errorf("HTTP %d for %s", response.StatusCode, response.URL)
	}
	return nil
}
