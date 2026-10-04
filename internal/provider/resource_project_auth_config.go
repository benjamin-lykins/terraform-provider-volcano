package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/benjamin-lykins/terraform-provider-volcano/internal/client"
)

var (
	_ resource.Resource                = &projectAuthConfigResource{}
	_ resource.ResourceWithConfigure   = &projectAuthConfigResource{}
	_ resource.ResourceWithImportState = &projectAuthConfigResource{}
)

func NewProjectAuthConfigResource() resource.Resource {
	return &projectAuthConfigResource{}
}

type projectAuthConfigResource struct {
	client *client.Client
}

type projectAuthConfigResourceModel struct {
	ProjectID                   types.String `tfsdk:"project_id"`
	AccessTokenLifetime         types.Int64  `tfsdk:"access_token_lifetime"`
	AllowPasswordReset          types.Bool   `tfsdk:"allow_password_reset"`
	AllowedEmailDomains         types.List   `tfsdk:"allowed_email_domains"`
	AllowedEmailDomainsMode     types.String `tfsdk:"allowed_email_domains_mode"`
	AllowedRedirectURLs         types.List   `tfsdk:"allowed_redirect_urls"`
	AutoLinkVerifiedOAuth       types.Bool   `tfsdk:"auto_link_verified_oauth"`
	CORSAllowCredentials        types.Bool   `tfsdk:"cors_allow_credentials"`
	CORSAllowedOrigins          types.List   `tfsdk:"cors_allowed_origins"`
	CORSEnabled                 types.Bool   `tfsdk:"cors_enabled"`
	CORSMaxAge                  types.Int64  `tfsdk:"cors_max_age"`
	DeviceVerificationURL       types.String `tfsdk:"device_verification_url"`
	EmailConfirmationSubject    types.String `tfsdk:"email_confirmation_subject"`
	EmailConfirmationTimeout    types.Int64  `tfsdk:"email_confirmation_timeout"`
	EmailEnabled                types.Bool   `tfsdk:"email_enabled"`
	EmailFromAddress            types.String `tfsdk:"email_from_address"`
	EmailFromName               types.String `tfsdk:"email_from_name"`
	EmailPasswordChangedSubject types.String `tfsdk:"email_password_changed_subject"`
	EmailPasswordResetSubject   types.String `tfsdk:"email_password_reset_subject"`
	EnableAnonymousSignins      types.Bool   `tfsdk:"enable_anonymous_signins"`
	EnableEmailPassword         types.Bool   `tfsdk:"enable_email_password"`
	EnableSignup                types.Bool   `tfsdk:"enable_signup"`
	InactivityTimeout           types.Int64  `tfsdk:"inactivity_timeout"`
	ManagedAuthEnabled          types.Bool   `tfsdk:"managed_auth_enabled"`
	MaxPasswordHistory          types.Int64  `tfsdk:"max_password_history"`
	MaxSessionDuration          types.Int64  `tfsdk:"max_session_duration"`
	MinPasswordLength           types.Int64  `tfsdk:"min_password_length"`
	PasswordResetTimeout        types.Int64  `tfsdk:"password_reset_timeout"`
	PlatformTokenTTL            types.Int64  `tfsdk:"platform_token_ttl"`
	PostAuthRedirectURL         types.String `tfsdk:"post_auth_redirect_url"`
	PostLogoutRedirectURL       types.String `tfsdk:"post_logout_redirect_url"`
	RateLimitSignin             types.Int64  `tfsdk:"rate_limit_signin"`
	RateLimitSignup             types.Int64  `tfsdk:"rate_limit_signup"`
	RateLimitTokenRefresh       types.Int64  `tfsdk:"rate_limit_token_refresh"`
	RefreshTokenLifetime        types.Int64  `tfsdk:"refresh_token_lifetime"`
	RequireEmailConfirmation    types.Bool   `tfsdk:"require_email_confirmation"`
	RequireLowercase            types.Bool   `tfsdk:"require_lowercase"`
	RequireNumbers              types.Bool   `tfsdk:"require_numbers"`
	RequireSpecialChars         types.Bool   `tfsdk:"require_special_chars"`
	RequireUppercase            types.Bool   `tfsdk:"require_uppercase"`
	SMTPHost                    types.String `tfsdk:"smtp_host"`
	SMTPPassword                types.String `tfsdk:"smtp_password"`
	SMTPPasswordConfigured      types.Bool   `tfsdk:"smtp_password_configured"`
	SMTPPort                    types.Int64  `tfsdk:"smtp_port"`
	SMTPUseTLS                  types.Bool   `tfsdk:"smtp_use_tls"`
	SMTPUsername                types.String `tfsdk:"smtp_username"`
}

func (r *projectAuthConfigResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_auth_config"
}

func (r *projectAuthConfigResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	str := func(desc string) schema.StringAttribute {
		return schema.StringAttribute{Optional: true, Computed: true, Description: desc}
	}
	b := func(desc string) schema.BoolAttribute {
		return schema.BoolAttribute{Optional: true, Computed: true, Description: desc}
	}
	i := func(desc string) schema.Int64Attribute {
		return schema.Int64Attribute{Optional: true, Computed: true, Description: desc}
	}
	list := func(desc string) schema.ListAttribute {
		return schema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType, Description: desc}
	}

	resp.Schema = schema.Schema{
		Description: "A project's end-user authentication configuration. This is a singleton every project already has, so there is no create/delete on the API side - Create performs the first PUT and Delete is a local no-op (removing this resource from configuration does not reset the project to defaults).",
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"access_token_lifetime":          i("Access token lifetime in seconds."),
			"allow_password_reset":           b("Whether users may reset their password."),
			"allowed_email_domains":          list("Email domains allowed to sign up/in. Empty means no restriction."),
			"allowed_email_domains_mode":     schema.StringAttribute{Optional: true, Computed: true, Validators: []validator.String{stringvalidator.OneOf("disabled", "signup", "signup_and_signin")}},
			"allowed_redirect_urls":          list("Redirect URLs allowed after OAuth/email flows."),
			"auto_link_verified_oauth":       b("Automatically link OAuth identities with a verified, matching email to an existing account."),
			"cors_allow_credentials":         b("Whether CORS responses include credentials."),
			"cors_allowed_origins":           schema.ListAttribute{Computed: true, ElementType: types.StringType, Description: "Derived from allowed_redirect_urls; not independently settable."},
			"cors_enabled":                   schema.BoolAttribute{Computed: true},
			"cors_max_age":                   i("CORS preflight cache duration in seconds."),
			"device_verification_url":        str("URL shown during CLI device-code login."),
			"email_confirmation_subject":     str("Subject line for the signup confirmation email."),
			"email_confirmation_timeout":     i("Seconds an email confirmation link remains valid."),
			"email_enabled":                  b("Whether the project sends auth emails at all."),
			"email_from_address":             str("From address for auth emails."),
			"email_from_name":                str("From display name for auth emails."),
			"email_password_changed_subject": str("Subject line for the password-changed notification email."),
			"email_password_reset_subject":   str("Subject line for the password-reset email."),
			"enable_anonymous_signins":       b("Whether anonymous sign-in is allowed."),
			"enable_email_password":          b("Whether email/password sign-in is allowed."),
			"enable_signup":                  b("Whether new signups are allowed."),
			"inactivity_timeout":             i("Seconds of inactivity before a session is invalidated."),
			"managed_auth_enabled":           b("Whether Volcano's hosted auth pages are enabled."),
			"max_password_history":           i("Number of previous passwords a user may not reuse."),
			"max_session_duration":           i("Maximum session duration in seconds, regardless of activity."),
			"min_password_length":            i("Minimum password length."),
			"password_reset_timeout":         i("Seconds a password-reset link remains valid."),
			"platform_token_ttl":             schema.Int64Attribute{Computed: true},
			"post_auth_redirect_url":         str("Default redirect URL after successful authentication."),
			"post_logout_redirect_url":       str("Default redirect URL after logout."),
			"rate_limit_signin":              i("Max sign-in attempts per window."),
			"rate_limit_signup":              i("Max signups per window."),
			"rate_limit_token_refresh":       i("Max token refreshes per window."),
			"refresh_token_lifetime":         i("Refresh token lifetime in seconds."),
			"require_email_confirmation":     b("Whether a new signup must confirm their email before signing in."),
			"require_lowercase":              b("Whether passwords must contain a lowercase letter."),
			"require_numbers":                b("Whether passwords must contain a digit."),
			"require_special_chars":          b("Whether passwords must contain a special character."),
			"require_uppercase":              b("Whether passwords must contain an uppercase letter."),
			"smtp_host":                      str("Custom SMTP host. Omit to use the platform's default email sender."),
			"smtp_password": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Custom SMTP password. Write-only: the API never returns it, so this value is preserved as-is across Read once set.",
			},
			"smtp_password_configured": schema.BoolAttribute{Computed: true, Description: "Whether a custom SMTP password is currently set."},
			"smtp_port":                i("Custom SMTP port."),
			"smtp_use_tls":             b("Whether to use TLS for custom SMTP."),
			"smtp_username":            str("Custom SMTP username."),
		},
	}
}

func (r *projectAuthConfigResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData))
		return
	}
	r.client = c
}

func (r *projectAuthConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan projectAuthConfigResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in, ok := projectAuthConfigUpdateRequestFromModel(ctx, plan, &resp.Diagnostics)
	if !ok {
		return
	}

	cfg, err := r.client.UpdateProjectAuthConfig(ctx, plan.ProjectID.ValueString(), in)
	if err != nil {
		resp.Diagnostics.AddError("Error setting project auth config", err.Error())
		return
	}

	model := projectAuthConfigModelFromAPI(ctx, plan.ProjectID.ValueString(), plan.SMTPPassword, cfg, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *projectAuthConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state projectAuthConfigResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg, err := r.client.GetProjectAuthConfig(ctx, state.ProjectID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading project auth config", err.Error())
		return
	}

	model := projectAuthConfigModelFromAPI(ctx, state.ProjectID.ValueString(), state.SMTPPassword, cfg, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *projectAuthConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan projectAuthConfigResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in, ok := projectAuthConfigUpdateRequestFromModel(ctx, plan, &resp.Diagnostics)
	if !ok {
		return
	}

	cfg, err := r.client.UpdateProjectAuthConfig(ctx, plan.ProjectID.ValueString(), in)
	if err != nil {
		resp.Diagnostics.AddError("Error updating project auth config", err.Error())
		return
	}

	model := projectAuthConfigModelFromAPI(ctx, plan.ProjectID.ValueString(), plan.SMTPPassword, cfg, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *projectAuthConfigResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// No-op: auth config is inherent to the project and has no delete
	// endpoint. Removing this resource only drops it from Terraform state.
}

func (r *projectAuthConfigResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, pathProjectID, req.ID)...)
}
