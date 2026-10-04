package client

import "context"

type ThemeColors struct {
	Background string `json:"background"`
	Surface    string `json:"surface"`
	Text       string `json:"text"`
	Accent     string `json:"accent"`
	AccentText string `json:"accent_text"`
}

type Theme struct {
	Version int64       `json:"version,omitempty"`
	Colors  ThemeColors `json:"colors"`
	Font    string      `json:"font,omitempty"`
	Scale   string      `json:"scale,omitempty"`
	Density string      `json:"density,omitempty"`
	Radius  string      `json:"radius,omitempty"`
}

// AuthPagesAppearance is the merged current theme and per-page-type
// layout for the project's hosted auth pages. It is the only way to read
// back the currently-applied theme/layout, since neither has its own GET.
type AuthPagesAppearance struct {
	Theme   Theme             `json:"theme"`
	Layouts map[string]string `json:"layouts"`
}

func (c *Client) GetProjectAuthPagesAppearance(ctx context.Context, projectID string) (*AuthPagesAppearance, error) {
	var out AuthPagesAppearance
	path := "/projects/" + EncodePathSegment(projectID) + "/auth/pages/appearance"
	if err := c.Request(ctx, "GET", path, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

type setThemeRequest struct {
	Theme Theme `json:"theme"`
}

func (c *Client) SetProjectAuthTheme(ctx context.Context, projectID string, theme Theme) (*Theme, error) {
	var out setThemeRequest
	path := "/projects/" + EncodePathSegment(projectID) + "/auth/pages/theme"
	if err := c.Request(ctx, "PUT", path, nil, setThemeRequest{Theme: theme}, &out); err != nil {
		return nil, err
	}
	return &out.Theme, nil
}

func (c *Client) DeleteProjectAuthTheme(ctx context.Context, projectID string) error {
	path := "/projects/" + EncodePathSegment(projectID) + "/auth/pages/theme"
	return c.Request(ctx, "DELETE", path, nil, nil, nil)
}

type setPageLayoutRequest struct {
	Layout string `json:"layout"`
}

func (c *Client) SetProjectAuthPageLayout(ctx context.Context, projectID, pageType, layout string) (string, error) {
	var out setPageLayoutRequest
	path := "/projects/" + EncodePathSegment(projectID) + "/auth/pages/" + EncodePathSegment(pageType) + "/layout"
	if err := c.Request(ctx, "PUT", path, nil, setPageLayoutRequest{Layout: layout}, &out); err != nil {
		return "", err
	}
	return out.Layout, nil
}

func (c *Client) DeleteProjectAuthPageLayout(ctx context.Context, projectID, pageType string) error {
	path := "/projects/" + EncodePathSegment(projectID) + "/auth/pages/" + EncodePathSegment(pageType) + "/layout"
	return c.Request(ctx, "DELETE", path, nil, nil, nil)
}

// AuthHostedPage is the custom HTML/CSS override for one of the project's
// hosted auth pages (login, signup, forgot-password, device, verify-email,
// reset-password).
type AuthHostedPage struct {
	HTML string `json:"html,omitempty"`
	CSS  string `json:"css,omitempty"`
}

type authHostedPageResponse struct {
	Page AuthHostedPage `json:"page"`
}

func (c *Client) GetProjectAuthHostedPage(ctx context.Context, projectID, pageType string) (*AuthHostedPage, error) {
	var out authHostedPageResponse
	path := "/projects/" + EncodePathSegment(projectID) + "/auth/hosted-pages/" + EncodePathSegment(pageType)
	if err := c.Request(ctx, "GET", path, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out.Page, nil
}

func (c *Client) SetProjectAuthHostedPage(ctx context.Context, projectID, pageType, html, css string) (*AuthHostedPage, error) {
	var out authHostedPageResponse
	path := "/projects/" + EncodePathSegment(projectID) + "/auth/hosted-pages/" + EncodePathSegment(pageType)
	body := AuthHostedPage{HTML: html, CSS: css}
	if err := c.Request(ctx, "PUT", path, nil, body, &out); err != nil {
		return nil, err
	}
	return &out.Page, nil
}
