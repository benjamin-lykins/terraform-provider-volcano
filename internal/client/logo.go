package client

import "context"

// UploadProjectLogo uploads (or replaces, since there is no dedicated PUT)
// the project's logo image. filename is used only to derive the multipart
// part's content type; content is the raw image bytes.
func (c *Client) UploadProjectLogo(ctx context.Context, projectID, filename string, content []byte) (*Project, error) {
	var out Project
	path := "/projects/" + EncodePathSegment(projectID) + "/logo"
	file := MultipartFile{FieldName: "logo", Filename: filename, Content: content}
	if err := c.MultipartRequest(ctx, "POST", path, nil, file, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteProjectLogo(ctx context.Context, projectID string) error {
	return c.Request(ctx, "DELETE", "/projects/"+EncodePathSegment(projectID)+"/logo", nil, nil, nil)
}
