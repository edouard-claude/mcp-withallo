package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const defaultBaseURL = "https://api.withallo.com"

type AlloClient struct {
	http    *http.Client
	apiKey  string
	baseURL string
}

func NewAlloClient(apiKey string) *AlloClient {
	return &AlloClient{
		http:    &http.Client{Timeout: 30 * time.Second},
		apiKey:  apiKey,
		baseURL: defaultBaseURL,
	}
}

func (c *AlloClient) do(req *http.Request) ([]byte, int, error) {
	req.Header.Set("Authorization", c.apiKey)
	req.Header.Set("Accept", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return body, resp.StatusCode, nil
}

func (c *AlloClient) Get(ctx context.Context, path string, query url.Values) ([]byte, int, error) {
	u := c.baseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, 0, err
	}
	return c.do(req)
}

func (c *AlloClient) PostJSON(ctx context.Context, path string, body any) ([]byte, int, error) {
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			return nil, 0, fmt.Errorf("encode body: %w", err)
		}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, &buf)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	return c.do(req)
}
