package client

import "context"

// EmailTemplate overrides the subject/body of one auth email type for a
// project (e.g. confirmation, password-reset).
type EmailTemplate struct {
	ID           string `json:"id"`
	ProjectID    string `json:"project_id"`
	TemplateType string `json:"template_type"`
	Subject      string `json:"subject"`
	HTMLBody     string `json:"html_body"`
	TextBody     string `json:"text_body"`
	CreatedAt    string `json:"created_at,omitempty"`
	UpdatedAt    string `json:"updated_at,omitempty"`
}

type CreateEmailTemplateRequest struct {
	TemplateType string `json:"template_type"`
	Subject      string `json:"subject"`
	HTMLBody     string `json:"html_body"`
	TextBody     string `json:"text_body"`
}

type UpdateEmailTemplateRequest struct {
	Subject  *string `json:"subject,omitempty"`
	HTMLBody *string `json:"html_body,omitempty"`
	TextBody *string `json:"text_body,omitempty"`
}

func (c *Client) CreateEmailTemplate(ctx context.Context, projectID string, in CreateEmailTemplateRequest) (*EmailTemplate, error) {
	var out EmailTemplate
	path := "/projects/" + EncodePathSegment(projectID) + "/email-templates"
	if err := c.Request(ctx, "POST", path, nil, in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetEmailTemplate(ctx context.Context, projectID, templateType string) (*EmailTemplate, error) {
	var out EmailTemplate
	path := "/projects/" + EncodePathSegment(projectID) + "/email-templates/" + EncodePathSegment(templateType)
	if err := c.Request(ctx, "GET", path, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) UpdateEmailTemplate(ctx context.Context, projectID, templateType string, in UpdateEmailTemplateRequest) (*EmailTemplate, error) {
	var out EmailTemplate
	path := "/projects/" + EncodePathSegment(projectID) + "/email-templates/" + EncodePathSegment(templateType)
	if err := c.Request(ctx, "PUT", path, nil, in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteEmailTemplate(ctx context.Context, projectID, templateType string) error {
	path := "/projects/" + EncodePathSegment(projectID) + "/email-templates/" + EncodePathSegment(templateType)
	return c.Request(ctx, "DELETE", path, nil, nil, nil)
}

// EmailTemplateDefault is the platform's built-in default content for a
// template type, before any project override.
type EmailTemplateDefault struct {
	TemplateType string `json:"template_type"`
	Subject      string `json:"subject"`
	HTMLBody     string `json:"html_body"`
	TextBody     string `json:"text_body"`
}

func (c *Client) ListEmailTemplateDefaults(ctx context.Context) ([]EmailTemplateDefault, error) {
	var out struct {
		Data []EmailTemplateDefault `json:"data"`
	}
	if err := c.Request(ctx, "GET", "/email-templates/defaults", nil, nil, &out); err != nil {
		return nil, err
	}
	return out.Data, nil
}
