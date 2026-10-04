package client

import "context"

// RoutingRecord is a DNS record the caller must create to route a custom
// domain to the frontend.
type RoutingRecord struct {
	RecordType string `json:"record_type,omitempty"`
	Name       string `json:"name,omitempty"`
	Value      string `json:"value,omitempty"`
	Type       string `json:"type,omitempty"`
}

// FrontendDomain is a custom domain attached to a frontend.
type FrontendDomain struct {
	Domain                string          `json:"domain"`
	TLSMode               string          `json:"tls_mode"`
	DomainStatus          string          `json:"domain_status"`
	VerificationStatus    string          `json:"verification_status"`
	EffectiveURLs         []string        `json:"effective_urls,omitempty"`
	RequiredRoutingRecord *RoutingRecord  `json:"required_routing_record,omitempty"`
	RoutingTargetHostname string          `json:"routing_target_hostname,omitempty"`
	VerificationRecords   []RoutingRecord `json:"verification_records,omitempty"`
	CreatedAt             string          `json:"created_at,omitempty"`
	UpdatedAt             string          `json:"updated_at,omitempty"`
}

type CreateFrontendDomainRequest struct {
	Domain string            `json:"domain"`
	TLS    FrontendDomainTLS `json:"tls"`
}

type FrontendDomainTLS struct {
	Mode                string `json:"mode"`
	CertificatePEM      string `json:"certificate_pem"`
	PrivateKeyPEM       string `json:"private_key_pem"`
	CertificateChainPEM string `json:"certificate_chain_pem,omitempty"`
}

func (c *Client) GetFrontendDomain(ctx context.Context, projectID, frontendID string) (*FrontendDomain, error) {
	var out *FrontendDomain
	path := "/projects/" + EncodePathSegment(projectID) + "/frontends/" + EncodePathSegment(frontendID) + "/domain"
	if err := c.Request(ctx, "GET", path, nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) CreateFrontendDomain(ctx context.Context, projectID, frontendID string, in CreateFrontendDomainRequest) (*FrontendDomain, error) {
	var out FrontendDomain
	path := "/projects/" + EncodePathSegment(projectID) + "/frontends/" + EncodePathSegment(frontendID) + "/domain"
	if err := c.Request(ctx, "POST", path, nil, in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteFrontendDomain(ctx context.Context, projectID, frontendID string) error {
	path := "/projects/" + EncodePathSegment(projectID) + "/frontends/" + EncodePathSegment(frontendID) + "/domain"
	return c.Request(ctx, "DELETE", path, nil, nil, nil)
}
