package client

import "context"

// PasswordPolicy is the platform-wide password policy baseline (a
// project's auth config may tighten it further).
type PasswordPolicy struct {
	EffectiveMinLength           int64 `json:"effective_min_length"`
	MinConfigurableLength        int64 `json:"min_configurable_length"`
	MaxLength                    int64 `json:"max_length"`
	RequireLowercase             bool  `json:"require_lowercase"`
	RequireUppercase             bool  `json:"require_uppercase"`
	RequireNumbers               bool  `json:"require_numbers"`
	RequireSpecialChars          bool  `json:"require_special_chars"`
	CompromisedPasswordsRejected bool  `json:"compromised_passwords_rejected"`
}

func (c *Client) GetPasswordPolicy(ctx context.Context) (*PasswordPolicy, error) {
	var out PasswordPolicy
	if err := c.Request(ctx, "GET", "/auth/password-policy", nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
