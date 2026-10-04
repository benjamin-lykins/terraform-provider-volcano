package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

// UploadProjectLogo uploads (or replaces, since there is no dedicated PUT)
// the project's logo image. filename is used only to derive the multipart
// part's content type; content is the raw image bytes.
func (c *Client) UploadProjectLogo(ctx context.Context, projectID, filename string, content []byte) (*Project, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("logo", filename)
	if err != nil {
		return nil, fmt.Errorf("building multipart request: %w", err)
	}
	if _, err := part.Write(content); err != nil {
		return nil, fmt.Errorf("writing multipart body: %w", err)
	}
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("closing multipart writer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.endpoint+"/projects/"+EncodePathSegment(projectID)+"/logo", &body)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Accept", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("performing request: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		apiErr := &APIError{StatusCode: resp.StatusCode, Raw: string(respBytes)}
		var eb errorBody
		if json.Unmarshal(respBytes, &eb) == nil {
			apiErr.Message = eb.Error
		}
		return nil, apiErr
	}

	var out Project
	if err := json.Unmarshal(respBytes, &out); err != nil {
		return nil, fmt.Errorf("decoding response body: %w", err)
	}
	return &out, nil
}

func (c *Client) DeleteProjectLogo(ctx context.Context, projectID string) error {
	return c.Request(ctx, "DELETE", "/projects/"+EncodePathSegment(projectID)+"/logo", nil, nil, nil)
}
