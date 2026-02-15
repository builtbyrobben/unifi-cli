package api

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"time"
)

var (
	errLoginFailed   = errors.New("login failed")
	errUnauthorized  = errors.New("unauthorized")
	errEmptyHost     = errors.New("host URL is required")
	errEmptyUsername = errors.New("username is required")
	errEmptyPassword = errors.New("password is required")
)

// Credentials holds UniFi controller login credentials.
type Credentials struct {
	Host     string
	Username string
	Password string
	Site     string
}

// Client is an HTTP client with cookie-based session auth for UniFi controllers.
type Client struct {
	httpClient    *http.Client
	creds         Credentials
	userAgent     string
	authenticated bool
}

// ClientOption configures the Client.
type ClientOption func(*Client)

// WithInsecure skips TLS certificate verification.
func WithInsecure(insecure bool) ClientOption {
	return func(c *Client) {
		if insecure {
			transport := c.httpClient.Transport.(*http.Transport)
			transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec // user-requested via --insecure flag
		}
	}
}

// WithUserAgent sets the User-Agent header.
func WithUserAgent(ua string) ClientOption {
	return func(c *Client) {
		c.userAgent = ua
	}
}

// WithTimeout sets the HTTP client timeout.
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}

// NewClient creates a new UniFi API client with cookie session management.
func NewClient(creds Credentials, opts ...ClientOption) (*Client, error) {
	if creds.Host == "" {
		return nil, errEmptyHost
	}

	if creds.Username == "" {
		return nil, errEmptyUsername
	}

	if creds.Password == "" {
		return nil, errEmptyPassword
	}

	if creds.Site == "" {
		creds.Site = "default"
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("create cookie jar: %w", err)
	}

	c := &Client{
		httpClient: &http.Client{
			Timeout:   30 * time.Second,
			Jar:       jar,
			Transport: &http.Transport{},
		},
		creds:     creds,
		userAgent: "unifi-cli/1.0",
	}

	for _, opt := range opts {
		opt(c)
	}

	return c, nil
}

// Login authenticates with the UniFi controller and stores the session cookie.
func (c *Client) Login(ctx context.Context) error {
	body := map[string]string{
		"username": c.creds.Username,
		"password": c.creds.Password,
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal login body: %w", err)
	}

	url := c.creds.Host + "/api/auth/login"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("create login request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("execute login request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: status %d", errLoginFailed, resp.StatusCode)
	}

	c.authenticated = true

	return nil
}

// Request describes an HTTP request to the UniFi API.
type Request struct {
	Method  string
	Path    string
	Body    any
	Headers map[string]string
}

// Do performs an HTTP request, logging in first if needed, and retrying on 401.
func (c *Client) Do(ctx context.Context, req Request) (*http.Response, error) {
	if !c.authenticated {
		if err := c.Login(ctx); err != nil {
			return nil, err
		}
	}

	resp, err := c.doOnce(ctx, req)
	if err != nil {
		return nil, err
	}

	// Auto re-login on 401
	if resp.StatusCode == http.StatusUnauthorized {
		_ = resp.Body.Close()
		c.authenticated = false

		if err := c.Login(ctx); err != nil {
			return nil, fmt.Errorf("%w: re-login failed: %w", errUnauthorized, err)
		}

		return c.doOnce(ctx, req)
	}

	return resp, nil
}

func (c *Client) doOnce(ctx context.Context, req Request) (*http.Response, error) {
	var bodyReader io.Reader

	if req.Body != nil {
		bodyBytes, err := json.Marshal(req.Body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}

		bodyReader = bytes.NewReader(bodyBytes)
	}

	url := c.creds.Host + req.Path

	httpReq, err := http.NewRequestWithContext(ctx, req.Method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", c.userAgent)

	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}

	return resp, nil
}

// Get performs a GET request and decodes the response into result.
func (c *Client) Get(ctx context.Context, path string, result any) error {
	return c.doJSON(ctx, Request{Method: http.MethodGet, Path: path}, result)
}

// Post performs a POST request and decodes the response into result.
func (c *Client) Post(ctx context.Context, path string, body, result any) error {
	return c.doJSON(ctx, Request{Method: http.MethodPost, Path: path, Body: body}, result)
}

func (c *Client) doJSON(ctx context.Context, req Request, result any) error {
	resp, err := c.Do(ctx, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return parseAPIError(resp)
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}

	return nil
}

// APIError represents a UniFi API error.
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error (%d): %s", e.StatusCode, e.Message)
}

func parseAPIError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)

	var apiErr struct {
		Meta struct {
			RC  string `json:"rc"`
			Msg string `json:"msg"`
		} `json:"meta"`
		Message string `json:"message"`
		Error   string `json:"error"`
	}

	if json.Unmarshal(body, &apiErr) == nil {
		msg := apiErr.Meta.Msg

		if msg == "" {
			msg = apiErr.Message
		}

		if msg == "" {
			msg = apiErr.Error
		}

		if msg != "" {
			return &APIError{StatusCode: resp.StatusCode, Message: msg}
		}
	}

	return &APIError{
		StatusCode: resp.StatusCode,
		Message:    http.StatusText(resp.StatusCode),
	}
}

// SitePath returns the API path prefix for the configured site.
func (c *Client) SitePath() string {
	return "/proxy/network/api/s/" + c.creds.Site
}
