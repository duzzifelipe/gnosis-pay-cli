package apiclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client is an HTTP client for the Gnosis Pay API.
type Client struct {
	BaseURL    string
	JWT        string
	Origin     string
	UserAgent  string
	httpClient *http.Client
}

// New creates a new API client.
func New(baseURL, jwt string) *Client {
	return &Client{
		BaseURL:   baseURL,
		JWT:       jwt,
		UserAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// NewWithOrigin creates a new API client with an Origin header for CORS/WAF.
func NewWithOrigin(baseURL, jwt, origin string) *Client {
	c := New(baseURL, jwt)
	c.Origin = origin

	return c
}

// Response wraps a raw API response.
type Response struct {
	StatusCode int
	Body       []byte
	Headers    http.Header
}

// JSON unmarshals the response body into v.
func (r *Response) JSON(v interface{}) error {
	return json.Unmarshal(r.Body, v)
}

// String returns the response body as a string.
func (r *Response) String() string {
	return string(r.Body)
}

// Get performs an authenticated GET request.
func (c *Client) Get(path string) (*Response, error) {
	return c.do("GET", path, nil)
}

// Post performs an authenticated POST request with a JSON body.
func (c *Client) Post(path string, body interface{}) (*Response, error) {
	return c.do("POST", path, body)
}

// Put performs an authenticated PUT request with a JSON body.
func (c *Client) Put(path string, body interface{}) (*Response, error) {
	return c.do("PUT", path, body)
}

// Delete performs an authenticated DELETE request with a JSON body.
func (c *Client) Delete(path string, body interface{}) (*Response, error) {
	return c.do("DELETE", path, body)
}

func (c *Client) do(method, path string, body interface{}) (*Response, error) {
	url := c.BaseURL + path

	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "cross-site")

	if c.JWT != "" {
		req.Header.Set("Authorization", "Bearer "+c.JWT)
	}
	if c.Origin != "" {
		req.Header.Set("Origin", c.Origin)
		req.Header.Set("Referer", c.Origin+"/")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Body:       respBody,
		Headers:    resp.Header,
	}, nil
}

// PrettyJSON returns indented JSON from raw bytes, or the raw string on failure.
func PrettyJSON(data []byte) string {
	var obj interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		return string(data)
	}
	pretty, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return string(data)
	}
	return string(pretty)
}
