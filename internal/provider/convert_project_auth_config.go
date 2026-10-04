package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/benjamin-lykins/terraform-provider-volcano/internal/client"
)

func projectAuthConfigUpdateRequestFromModel(ctx context.Context, m projectAuthConfigResourceModel, diags *diag.Diagnostics) (client.UpdateProjectAuthConfigRequest, bool) {
	in := client.UpdateProjectAuthConfigRequest{
		AllowedEmailDomainsMode:     m.AllowedEmailDomainsMode.ValueString(),
		DeviceVerificationURL:       m.DeviceVerificationURL.ValueString(),
		EmailConfirmationSubject:    m.EmailConfirmationSubject.ValueString(),
		EmailFromAddress:            m.EmailFromAddress.ValueString(),
		EmailFromName:               m.EmailFromName.ValueString(),
		EmailPasswordChangedSubject: m.EmailPasswordChangedSubject.ValueString(),
		EmailPasswordResetSubject:   m.EmailPasswordResetSubject.ValueString(),
		PostAuthRedirectURL:         m.PostAuthRedirectURL.ValueString(),
		PostLogoutRedirectURL:       m.PostLogoutRedirectURL.ValueString(),
		SMTPHost:                    m.SMTPHost.ValueString(),
		SMTPPassword:                m.SMTPPassword.ValueString(),
		SMTPUsername:                m.SMTPUsername.ValueString(),
	}

	in.AccessTokenLifetime = int64PtrFromModel(m.AccessTokenLifetime)
	in.AllowPasswordReset = boolPtrFromModel(m.AllowPasswordReset)
	in.AutoLinkVerifiedOAuth = boolPtrFromModel(m.AutoLinkVerifiedOAuth)
	in.CORSAllowCredentials = boolPtrFromModel(m.CORSAllowCredentials)
	in.CORSMaxAge = int64PtrFromModel(m.CORSMaxAge)
	in.EmailConfirmationTimeout = int64PtrFromModel(m.EmailConfirmationTimeout)
	in.EmailEnabled = boolPtrFromModel(m.EmailEnabled)
	in.EnableAnonymousSignins = boolPtrFromModel(m.EnableAnonymousSignins)
	in.EnableEmailPassword = boolPtrFromModel(m.EnableEmailPassword)
	in.EnableSignup = boolPtrFromModel(m.EnableSignup)
	in.InactivityTimeout = int64PtrFromModel(m.InactivityTimeout)
	in.ManagedAuthEnabled = boolPtrFromModel(m.ManagedAuthEnabled)
	in.MaxPasswordHistory = int64PtrFromModel(m.MaxPasswordHistory)
	in.MaxSessionDuration = int64PtrFromModel(m.MaxSessionDuration)
	in.MinPasswordLength = int64PtrFromModel(m.MinPasswordLength)
	in.PasswordResetTimeout = int64PtrFromModel(m.PasswordResetTimeout)
	in.RateLimitSignin = int64PtrFromModel(m.RateLimitSignin)
	in.RateLimitSignup = int64PtrFromModel(m.RateLimitSignup)
	in.RateLimitTokenRefresh = int64PtrFromModel(m.RateLimitTokenRefresh)
	in.RefreshTokenLifetime = int64PtrFromModel(m.RefreshTokenLifetime)
	in.RequireEmailConfirmation = boolPtrFromModel(m.RequireEmailConfirmation)
	in.RequireLowercase = boolPtrFromModel(m.RequireLowercase)
	in.RequireNumbers = boolPtrFromModel(m.RequireNumbers)
	in.RequireSpecialChars = boolPtrFromModel(m.RequireSpecialChars)
	in.RequireUppercase = boolPtrFromModel(m.RequireUppercase)
	in.SMTPPort = int64PtrFromModel(m.SMTPPort)
	in.SMTPUseTLS = boolPtrFromModel(m.SMTPUseTLS)

	if !m.AllowedEmailDomains.IsNull() && !m.AllowedEmailDomains.IsUnknown() {
		var v []string
		diags.Append(m.AllowedEmailDomains.ElementsAs(ctx, &v, false)...)
		in.AllowedEmailDomains = v
	}
	if !m.AllowedRedirectURLs.IsNull() && !m.AllowedRedirectURLs.IsUnknown() {
		var v []string
		diags.Append(m.AllowedRedirectURLs.ElementsAs(ctx, &v, false)...)
		in.AllowedRedirectURLs = v
	}
	if diags.HasError() {
		return in, false
	}
	return in, true
}

func int64PtrFromModel(v types.Int64) *int64 {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	x := v.ValueInt64()
	return &x
}

func boolPtrFromModel(v types.Bool) *bool {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	x := v.ValueBool()
	return &x
}

func projectAuthConfigModelFromAPI(ctx context.Context, projectID string, smtpPassword types.String, c *client.ProjectAuthConfig, diags *diag.Diagnostics) projectAuthConfigResourceModel {
	return projectAuthConfigResourceModel{
		ProjectID:                   types.StringValue(projectID),
		AccessTokenLifetime:         types.Int64Value(c.AccessTokenLifetime),
		AllowPasswordReset:          types.BoolValue(c.AllowPasswordReset),
		AllowedEmailDomains:         stringList(ctx, c.AllowedEmailDomains, diags),
		AllowedEmailDomainsMode:     stringOrNull(c.AllowedEmailDomainsMode),
		AllowedRedirectURLs:         stringList(ctx, c.AllowedRedirectURLs, diags),
		AutoLinkVerifiedOAuth:       types.BoolValue(c.AutoLinkVerifiedOAuth),
		CORSAllowCredentials:        types.BoolValue(c.CORSAllowCredentials),
		CORSAllowedOrigins:          stringList(ctx, c.CORSAllowedOrigins, diags),
		CORSEnabled:                 types.BoolValue(c.CORSEnabled),
		CORSMaxAge:                  types.Int64Value(c.CORSMaxAge),
		DeviceVerificationURL:       stringOrNull(c.DeviceVerificationURL),
		EmailConfirmationSubject:    stringOrNull(c.EmailConfirmationSubject),
		EmailConfirmationTimeout:    types.Int64Value(c.EmailConfirmationTimeout),
		EmailEnabled:                types.BoolValue(c.EmailEnabled),
		EmailFromAddress:            stringOrNull(c.EmailFromAddress),
		EmailFromName:               stringOrNull(c.EmailFromName),
		EmailPasswordChangedSubject: stringOrNull(c.EmailPasswordChangedSubject),
		EmailPasswordResetSubject:   stringOrNull(c.EmailPasswordResetSubject),
		EnableAnonymousSignins:      types.BoolValue(c.EnableAnonymousSignins),
		EnableEmailPassword:         types.BoolValue(c.EnableEmailPassword),
		EnableSignup:                types.BoolValue(c.EnableSignup),
		InactivityTimeout:           types.Int64Value(c.InactivityTimeout),
		ManagedAuthEnabled:          types.BoolValue(c.ManagedAuthEnabled),
		MaxPasswordHistory:          types.Int64Value(c.MaxPasswordHistory),
		MaxSessionDuration:          types.Int64Value(c.MaxSessionDuration),
		MinPasswordLength:           types.Int64Value(c.MinPasswordLength),
		PasswordResetTimeout:        types.Int64Value(c.PasswordResetTimeout),
		PlatformTokenTTL:            types.Int64Value(c.PlatformTokenTTL),
		PostAuthRedirectURL:         stringOrNull(c.PostAuthRedirectURL),
		PostLogoutRedirectURL:       stringOrNull(c.PostLogoutRedirectURL),
		RateLimitSignin:             types.Int64Value(c.RateLimitSignin),
		RateLimitSignup:             types.Int64Value(c.RateLimitSignup),
		RateLimitTokenRefresh:       types.Int64Value(c.RateLimitTokenRefresh),
		RefreshTokenLifetime:        types.Int64Value(c.RefreshTokenLifetime),
		RequireEmailConfirmation:    types.BoolValue(c.RequireEmailConfirmation),
		RequireLowercase:            types.BoolValue(c.RequireLowercase),
		RequireNumbers:              types.BoolValue(c.RequireNumbers),
		RequireSpecialChars:         types.BoolValue(c.RequireSpecialChars),
		RequireUppercase:            types.BoolValue(c.RequireUppercase),
		SMTPHost:                    stringOrNull(c.SMTPHost),
		SMTPPassword:                smtpPassword,
		SMTPPasswordConfigured:      types.BoolValue(c.SMTPPasswordConfigured),
		SMTPPort:                    types.Int64Value(c.SMTPPort),
		SMTPUseTLS:                  types.BoolValue(c.SMTPUseTLS),
		SMTPUsername:                stringOrNull(c.SMTPUsername),
	}
}
