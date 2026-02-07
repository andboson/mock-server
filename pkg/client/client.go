// Package client provides a Go client for interacting with the Mock Server API.
// It allows users to programmatically create expectations, check for matches,
// and manage mock rules within the Mock Server.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Client is a client for the Mock Server API.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// New creates a new Client.
// baseURL should include the scheme and host, e.g., "http://localhost:8081".
// If httpClient is nil, http.DefaultClient is used.
func New(baseURL string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: httpClient,
	}
}

// CreateExpectation adds a new expectation to the server.
func (c *Client) CreateExpectation(ctx context.Context, exp ExpectationCreate) (*ExpectationID, error) {
	var resp ExpectationID
	err := c.do(ctx, http.MethodPost, "/api/expectation", exp, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to create expectation: %w", err)
	}
	return &resp, nil
}

// UpdateExpectation updates an existing expectation.
func (c *Client) UpdateExpectation(ctx context.Context, id string, exp ExpectationCreate) (*ExpectationID, error) {
	var resp ExpectationID
	err := c.do(ctx, http.MethodPut, fmt.Sprintf("/api/expectation/%s", id), exp, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to update expectation: %w", err)
	}
	return &resp, nil
}

// CheckExpectation checks if an expectation was matched.
func (c *Client) CheckExpectation(ctx context.Context, id string) (*MatchStatus, error) {
	var resp MatchStatus
	err := c.do(ctx, http.MethodGet, fmt.Sprintf("/api/expectation/%s", id), nil, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to check expectation: %w", err)
	}
	return &resp, nil
}

// RemoveExpectation removes an expectation.
func (c *Client) RemoveExpectation(ctx context.Context, id string) error {
	if err := c.do(ctx, http.MethodDelete, fmt.Sprintf("/api/expectation/%s", id), nil, nil); err != nil {
		return fmt.Errorf("failed to remove expectation: %w", err)
	}
	return nil
}

// GetExpectations gets all expectations.
func (c *Client) GetExpectations(ctx context.Context) ([]Expectation, error) {
	var resp []Expectation
	err := c.do(ctx, http.MethodGet, "/api/expectations", nil, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to get expectation list: %w", err)
	}
	return resp, nil
}

// do simplifies making HTTP requests and decoding responses.
func (c *Client) do(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	req, err := c.newRequest(ctx, method, path, body)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to perform request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	return c.handleResponse(resp, result)
}

func (c *Client) newRequest(ctx context.Context, method, path string, body interface{}) (*http.Request, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}

func (c *Client) handleResponse(resp *http.Response, result interface{}) error {
	if resp.StatusCode >= 400 {
		return fmt.Errorf("server returned error status: %d", resp.StatusCode)
	}

	if result != nil && resp.StatusCode != http.StatusNoContent {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}
