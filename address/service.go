package address

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type Service struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

type AddressSuggestion struct {
	Number   string `json:"number"`
	Street   string `json:"street"`
	City     string `json:"city"`
	Postcode string `json:"postcode"`
	Region   string `json:"region"`
}

type Option func(*Service)

func WithBaseURL(baseURL string) Option {
	return func(s *Service) {
		s.baseURL = baseURL
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(s *Service) {
		s.client.Timeout = timeout
	}
}

func NewService(apiKey string, opts ...Option) *Service {
	s := &Service{
		baseURL: "https://address-lookup.sudox.io",
		apiKey:  apiKey,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func (s *Service) Lookup(ctx context.Context, query string) ([]AddressSuggestion, error) {
	if query == "" {
		return []AddressSuggestion{}, nil
	}

	reqURL := fmt.Sprintf("%s?query=%s", s.baseURL, url.QueryEscape(query))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if s.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+s.apiKey)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	var suggestions []AddressSuggestion
	if err := json.NewDecoder(resp.Body).Decode(&suggestions); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return suggestions, nil
}
