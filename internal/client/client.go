// Package client is a thin, hand-written REST client for the Volcano
// platform API (https://api.volcano.dev). It handles authentication,
// request/response marshalling, pagination, and typed error reporting so
// that the provider and resource code never touches net/http directly.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
)

const DefaultEndpoint = "https://api.volcano.dev"

// Client is the shared handle used by every resource and data source.
type Client struct {
	endpoint   string
	token      string
	httpClient *http.Client
}

// New constructs a Client. endpoint must be an absolute base URL (no
// trailing slash required). token is sent as a Bearer credential on every
// request; it may be a platform token (pk-), project access token (pt-),
// anon key, or service key depending on which routes the caller exercises.
func New(endpoint, token string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{
		endpoint:   strings.TrimRight(endpoint, "/"),
		token:      token,
		httpClient: httpClient,
	}
}

// APIError represents a non-2xx response from the Volcano API. The body is
// documented as always being `{"error": "<description>"}` on failure.
type APIError struct {
	StatusCode int
	Message    string
	Raw        string
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("volcano API error (%d): %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("volcano API error (%d): %s", e.StatusCode, e.Raw)
}

// IsNotFound reports whether err is an APIError with a 404 status, which
// resource Read implementations use to decide whether to drop state.
func IsNotFound(err error) bool {
	apiErr, ok := err.(*APIError)
	return ok && apiErr.StatusCode == http.StatusNotFound
}

type errorBody struct {
	Error string `json:"error"`
}

// Request performs a single HTTP call against the Volcano API.
//
// path is joined to the configured endpoint as-is (callers are responsible
// for URL-encoding path segments that come from user input, e.g. resource
// names that may contain reserved characters).
//
// body, if non-nil, is marshalled as JSON. out, if non-nil, receives the
// JSON-decoded response body on success (2xx). A nil out is valid for
// requests whose success response has no body (e.g. DELETE).
func (c *Client) Request(ctx context.Context, method, path string, query url.Values, body, out interface{}) error {
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("encoding request body: %w", err)
		}
		reqBody = bytes.NewReader(b)
	}

	u := c.endpoint + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, u, reqBody)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("performing request: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		apiErr := &APIError{StatusCode: resp.StatusCode, Raw: string(respBytes)}
		var eb errorBody
		if json.Unmarshal(respBytes, &eb) == nil {
			apiErr.Message = eb.Error
		}
		return apiErr
	}

	if out != nil && len(respBytes) > 0 {
		if err := json.Unmarshal(respBytes, out); err != nil {
			return fmt.Errorf("decoding response body: %w", err)
		}
	}

	return nil
}

// MultipartField is one non-file form field in a multipart request.
type MultipartField struct {
	Name  string
	Value string
}

// MultipartFile describes the single file part of a multipart request.
type MultipartFile struct {
	FieldName string
	Filename  string
	Content   []byte
}

// MultipartRequest performs a multipart/form-data request, used by the
// handful of endpoints (function/durable-function create, project logo
// upload) that accept a file alongside ordinary form fields.
func (c *Client) MultipartRequest(ctx context.Context, method, path string, fields []MultipartField, file MultipartFile, out interface{}) error {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	for _, f := range fields {
		if err := writer.WriteField(f.Name, f.Value); err != nil {
			return fmt.Errorf("writing multipart field %q: %w", f.Name, err)
		}
	}
	if file.FieldName != "" {
		part, err := writer.CreateFormFile(file.FieldName, file.Filename)
		if err != nil {
			return fmt.Errorf("building multipart file part: %w", err)
		}
		if _, err := part.Write(file.Content); err != nil {
			return fmt.Errorf("writing multipart file content: %w", err)
		}
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("closing multipart writer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.endpoint+path, &body)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Accept", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("performing request: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		apiErr := &APIError{StatusCode: resp.StatusCode, Raw: string(respBytes)}
		var eb errorBody
		if json.Unmarshal(respBytes, &eb) == nil {
			apiErr.Message = eb.Error
		}
		return apiErr
	}

	if out != nil && len(respBytes) > 0 {
		if err := json.Unmarshal(respBytes, out); err != nil {
			return fmt.Errorf("decoding response body: %w", err)
		}
	}
	return nil
}

// Page is the common pagination envelope returned by every list endpoint.
type Page[T any] struct {
	Data       []T    `json:"data"`
	Page       int    `json:"page,omitempty"`
	Limit      int    `json:"limit,omitempty"`
	Total      int    `json:"total,omitempty"`
	HasMore    bool   `json:"has_more,omitempty"`
	NextCursor string `json:"next_cursor,omitempty"`
}

// EncodePathSegment percent-encodes a single path segment supplied by a
// caller (resource names, keys, etc.) for safe inclusion in a request path.
func EncodePathSegment(s string) string {
	return url.PathEscape(s)
}
