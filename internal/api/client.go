package api

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client is the YouTrack REST API client.
type Client struct {
	BaseURL    string
	token      string
	httpClient *http.Client
	maxRetries int
	retryDelay time.Duration
}

// APIError represents an error from the YouTrack API.
type APIError struct {
	StatusCode int
	Message    string
	Body       string
}

func (e *APIError) Error() string {
	if e.Body != "" {
		return fmt.Sprintf("API error %d: %s — %s", e.StatusCode, e.Message, e.Body)
	}
	return fmt.Sprintf("API error %d: %s", e.StatusCode, e.Message)
}

// NewClient creates a new API client.
func NewClient(baseURL, token string, verifySSL bool) *Client {
	// Clean trailing slash
	baseURL = strings.TrimRight(baseURL, "/")

	transport := &http.Transport{}
	if !verifySSL {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec
	}

	return &Client{
		BaseURL: baseURL,
		token:   token,
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   30 * time.Second,
		},
		maxRetries: 3,
		retryDelay: 1 * time.Second,
	}
}

// doRequest executes an HTTP request with retry logic.
func (c *Client) doRequest(method, endpoint string, body interface{}) ([]byte, error) {
	url := fmt.Sprintf("%s/api/%s", c.BaseURL, strings.TrimLeft(endpoint, "/"))

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshaling request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff with jitter
			backoff := c.retryDelay * time.Duration(math.Pow(2, float64(attempt)))
			jitter := time.Duration(rand.Int63n(int64(backoff) / 2)) //nolint:gosec
			time.Sleep(backoff + jitter)
		}

		// Reset body reader for retries
		if body != nil {
			data, _ := json.Marshal(body)
			bodyReader = bytes.NewReader(data)
		}

		req, err := http.NewRequest(method, url, bodyReader)
		if err != nil {
			return nil, fmt.Errorf("creating request: %w", err)
		}

		req.Header.Set("Authorization", "Bearer "+c.token)
		req.Header.Set("Accept", "application/json")
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			continue
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("reading response: %w", err)
			continue
		}

		// Success
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return respBody, nil
		}

		// Retry on 429 and 5xx
		if resp.StatusCode == 429 || resp.StatusCode >= 500 {
			lastErr = &APIError{
				StatusCode: resp.StatusCode,
				Message:    http.StatusText(resp.StatusCode),
				Body:       string(respBody),
			}
			continue
		}

		// Non-retryable error
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    http.StatusText(resp.StatusCode),
			Body:       string(respBody),
		}
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// Get performs a GET request.
func (c *Client) Get(endpoint string, params map[string]string) ([]byte, error) {
	if len(params) > 0 {
		v := url.Values{}
		for k, val := range params {
			v.Set(k, val)
		}
		endpoint = endpoint + "?" + v.Encode()
	}
	return c.doRequest("GET", endpoint, nil)
}

// Post performs a POST request.
func (c *Client) Post(endpoint string, body interface{}) ([]byte, error) {
	return c.doRequest("POST", endpoint, body)
}

// Delete performs a DELETE request.
func (c *Client) Delete(endpoint string) ([]byte, error) {
	return c.doRequest("DELETE", endpoint, nil)
}

// TestConnection tests the API connection by fetching current user info.
func (c *Client) TestConnection() (string, error) {
	data, err := c.Get("users/me", map[string]string{"fields": "id,login,name,email"})
	if err != nil {
		return "", err
	}
	var user struct {
		Login string `json:"login"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	if err := json.Unmarshal(data, &user); err != nil {
		return "", fmt.Errorf("parsing user info: %w", err)
	}
	if user.Name != "" {
		return fmt.Sprintf("%s (%s)", user.Name, user.Login), nil
	}
	return user.Login, nil
}
