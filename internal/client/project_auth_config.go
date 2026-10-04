package client

import "context"

// ProjectAuthConfig is the project's end-user authentication configuration.
// It is a singleton: every project has one from creation, so this
// resource only ever GETs and PUTs it, never creates or deletes it.
type ProjectAuthConfig struct {
	ProjectID                   string   `json:"project_id,omitempty"`
	AccessTokenLifetime         int64    `json:"access_token_lifetime,omitempty"`
	AllowPasswordReset          bool     `json:"allow_password_reset,omitempty"`
	AllowedEmailDomains         []string `json:"allowed_email_domains,omitempty"`
	AllowedEmailDomainsMode     string   `json:"allowed_email_domains_mode,omitempty"`
	AllowedRedirectURLs         []string `json:"allowed_redirect_urls,omitempty"`
	AutoLinkVerifiedOAuth       bool     `json:"auto_link_verified_oauth,omitempty"`
	CORSAllowCredentials        bool     `json:"cors_allow_credentials,omitempty"`
	CORSAllowedOrigins          []string `json:"cors_allowed_origins,omitempty"`
	CORSEnabled                 bool     `json:"cors_enabled,omitempty"`
	CORSMaxAge                  int64    `json:"cors_max_age,omitempty"`
	DeviceVerificationURL       string   `json:"device_verification_url,omitempty"`
	EmailConfirmationSubject    string   `json:"email_confirmation_subject,omitempty"`
	EmailConfirmationTimeout    int64    `json:"email_confirmation_timeout,omitempty"`
	EmailEnabled                bool     `json:"email_enabled,omitempty"`
	EmailFromAddress            string   `json:"email_from_address,omitempty"`
	EmailFromName               string   `json:"email_from_name,omitempty"`
	EmailPasswordChangedSubject string   `json:"email_password_changed_subject,omitempty"`
	EmailPasswordResetSubject   string   `json:"email_password_reset_subject,omitempty"`
	EnableAnonymousSignins      bool     `json:"enable_anonymous_signins,omitempty"`
	EnableEmailPassword         bool     `json:"enable_email_password,omitempty"`
	EnableSignup                bool     `json:"enable_signup,omitempty"`
	InactivityTimeout           int64    `json:"inactivity_timeout,omitempty"`
	ManagedAuthEnabled          bool     `json:"managed_auth_enabled,omitempty"`
	MaxPasswordHistory          int64    `json:"max_password_history,omitempty"`
	MaxSessionDuration          int64    `json:"max_session_duration,omitempty"`
	MinPasswordLength           int64    `json:"min_password_length,omitempty"`
	PasswordResetTimeout        int64    `json:"password_reset_timeout,omitempty"`
	PlatformTokenTTL            int64    `json:"platform_token_ttl,omitempty"`
	PostAuthRedirectURL         string   `json:"post_auth_redirect_url,omitempty"`
	PostLogoutRedirectURL       string   `json:"post_logout_redirect_url,omitempty"`
	RateLimitSignin             int64    `json:"rate_limit_signin,omitempty"`
	RateLimitSignup             int64    `json:"rate_limit_signup,omitempty"`
	RateLimitTokenRefresh       int64    `json:"rate_limit_token_refresh,omitempty"`
	RefreshTokenLifetime        int64    `json:"refresh_token_lifetime,omitempty"`
	RequireEmailConfirmation    bool     `json:"require_email_confirmation,omitempty"`
	RequireLowercase            bool     `json:"require_lowercase,omitempty"`
	RequireNumbers              bool     `json:"require_numbers,omitempty"`
	RequireSpecialChars         bool     `json:"require_special_chars,omitempty"`
	RequireUppercase            bool     `json:"require_uppercase,omitempty"`
	SMTPHost                    string   `json:"smtp_host,omitempty"`
	SMTPPasswordConfigured      bool     `json:"smtp_password_configured,omitempty"`
	SMTPPort                    int64    `json:"smtp_port,omitempty"`
	SMTPUseTLS                  bool     `json:"smtp_use_tls,omitempty"`
	SMTPUsername                string   `json:"smtp_username,omitempty"`
}

// UpdateProjectAuthConfigRequest mirrors the mutable subset of
// ProjectAuthConfig; cors_allowed_origins, cors_enabled, and
// platform_token_ttl are server-derived and not settable.
type UpdateProjectAuthConfigRequest struct {
	AccessTokenLifetime         *int64   `json:"access_token_lifetime,omitempty"`
	AllowPasswordReset          *bool    `json:"allow_password_reset,omitempty"`
	AllowedEmailDomains         []string `json:"allowed_email_domains,omitempty"`
	AllowedEmailDomainsMode     string   `json:"allowed_email_domains_mode,omitempty"`
	AllowedRedirectURLs         []string `json:"allowed_redirect_urls,omitempty"`
	AutoLinkVerifiedOAuth       *bool    `json:"auto_link_verified_oauth,omitempty"`
	CORSAllowCredentials        *bool    `json:"cors_allow_credentials,omitempty"`
	CORSMaxAge                  *int64   `json:"cors_max_age,omitempty"`
	DeviceVerificationURL       string   `json:"device_verification_url,omitempty"`
	EmailConfirmationSubject    string   `json:"email_confirmation_subject,omitempty"`
	EmailConfirmationTimeout    *int64   `json:"email_confirmation_timeout,omitempty"`
	EmailEnabled                *bool    `json:"email_enabled,omitempty"`
	EmailFromAddress            string   `json:"email_from_address,omitempty"`
	EmailFromName               string   `json:"email_from_name,omitempty"`
	EmailPasswordChangedSubject string   `json:"email_password_changed_subject,omitempty"`
	EmailPasswordResetSubject   string   `json:"email_password_reset_subject,omitempty"`
	EnableAnonymousSignins      *bool    `json:"enable_anonymous_signins,omitempty"`
	EnableEmailPassword         *bool    `json:"enable_email_password,omitempty"`
	EnableSignup                *bool    `json:"enable_signup,omitempty"`
	InactivityTimeout           *int64   `json:"inactivity_timeout,omitempty"`
	ManagedAuthEnabled          *bool    `json:"managed_auth_enabled,omitempty"`
	MaxPasswordHistory          *int64   `json:"max_password_history,omitempty"`
	MaxSessionDuration          *int64   `json:"max_session_duration,omitempty"`
	MinPasswordLength           *int64   `json:"min_password_length,omitempty"`
	PasswordResetTimeout        *int64   `json:"password_reset_timeout,omitempty"`
	PostAuthRedirectURL         string   `json:"post_auth_redirect_url,omitempty"`
	PostLogoutRedirectURL       string   `json:"post_logout_redirect_url,omitempty"`
	RateLimitSignin             *int64   `json:"rate_limit_signin,omitempty"`
	RateLimitSignup             *int64   `json:"rate_limit_signup,omitempty"`
	RateLimitTokenRefresh       *int64   `json:"rate_limit_token_refresh,omitempty"`
	RefreshTokenLifetime        *int64   `json:"refresh_token_lifetime,omitempty"`
	RequireEmailConfirmation    *bool    `json:"require_email_confirmation,omitempty"`
	RequireLowercase            *bool    `json:"require_lowercase,omitempty"`
	RequireNumbers              *bool    `json:"require_numbers,omitempty"`
	RequireSpecialChars         *bool    `json:"require_special_chars,omitempty"`
	RequireUppercase            *bool    `json:"require_uppercase,omitempty"`
	SMTPHost                    string   `json:"smtp_host,omitempty"`
	SMTPPassword                string   `json:"smtp_password,omitempty"`
	SMTPPort                    *int64   `json:"smtp_port,omitempty"`
	SMTPUseTLS                  *bool    `json:"smtp_use_tls,omitempty"`
	SMTPUsername                string   `json:"smtp_username,omitempty"`
}

func (c *Client) GetProjectAuthConfig(ctx context.Context, projectID string) (*ProjectAuthConfig, error) {
	var out ProjectAuthConfig
	path := "/projects/" + EncodePathSegment(projectID) + "/auth/config"
	if err := c.Request(ctx, "GET", path, nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) UpdateProjectAuthConfig(ctx context.Context, projectID string, in UpdateProjectAuthConfigRequest) (*ProjectAuthConfig, error) {
	var out ProjectAuthConfig
	path := "/projects/" + EncodePathSegment(projectID) + "/auth/config"
	if err := c.Request(ctx, "PUT", path, nil, in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
