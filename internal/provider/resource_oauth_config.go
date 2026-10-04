package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/benjamin-lykins/terraform-provider-volcano/internal/client"
)

var (
	_ resource.Resource                = &oauthConfigResource{}
	_ resource.ResourceWithConfigure   = &oauthConfigResource{}
	_ resource.ResourceWithImportState = &oauthConfigResource{}
)

func NewOAuthConfigResource() resource.Resource {
	return &oauthConfigResource{}
}

type oauthConfigResource struct {
	client *client.Client
}

type oauthConfigResourceModel struct {
	ID           types.String `tfsdk:"id"`
	ProjectID    types.String `tfsdk:"project_id"`
	Provider     types.String `tfsdk:"oauth_provider"`
	ClientID     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
	RedirectURL  types.String `tfsdk:"redirect_url"`
	Scopes       types.List   `tfsdk:"scopes"`
	Enabled      types.Bool   `tfsdk:"enabled"`
	CreatedAt    types.String `tfsdk:"created_at"`
	UpdatedAt    types.String `tfsdk:"updated_at"`
}

func (r *oauthConfigResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_oauth_config"
}

func (r *oauthConfigResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A project's OAuth provider configuration (Google, GitHub, Microsoft, Apple, or device-code).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"oauth_provider": schema.StringAttribute{
				Required:    true,
				Description: "google, github, microsoft, apple, or device.",
				Validators: []validator.String{
					stringvalidator.OneOf("google", "github", "microsoft", "apple", "device"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"client_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"client_secret": schema.StringAttribute{
				Optional:  true,
				Computed:  true,
				Sensitive: true,
			},
			"redirect_url": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"scopes": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"created_at": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (r *oauthConfigResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *oauthConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan oauthConfigResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := client.CreateOAuthConfigRequest{
		Provider:     plan.Provider.ValueString(),
		ClientID:     plan.ClientID.ValueString(),
		ClientSecret: plan.ClientSecret.ValueString(),
		RedirectURL:  plan.RedirectURL.ValueString(),
	}
	if !plan.Scopes.IsNull() && !plan.Scopes.IsUnknown() {
		var v []string
		resp.Diagnostics.Append(plan.Scopes.ElementsAs(ctx, &v, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		in.Scopes = v
	}

	cfg, err := r.client.CreateOAuthConfig(ctx, plan.ProjectID.ValueString(), in)
	if err != nil {
		resp.Diagnostics.AddError("Error creating OAuth config", err.Error())
		return
	}

	model := oauthConfigModelFromAPI(ctx, plan.ProjectID.ValueString(), cfg, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *oauthConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state oauthConfigResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg, err := r.client.GetOAuthConfig(ctx, state.ProjectID.ValueString(), state.Provider.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading OAuth config", err.Error())
		return
	}

	model := oauthConfigModelFromAPI(ctx, state.ProjectID.ValueString(), cfg, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *oauthConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state oauthConfigResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := client.UpdateOAuthConfigRequest{
		ClientID:     plan.ClientID.ValueString(),
		ClientSecret: plan.ClientSecret.ValueString(),
		RedirectURL:  plan.RedirectURL.ValueString(),
	}
	if !plan.Enabled.IsNull() && !plan.Enabled.IsUnknown() {
		v := plan.Enabled.ValueBool()
		in.Enabled = &v
	}
	if !plan.Scopes.IsNull() && !plan.Scopes.IsUnknown() {
		var v []string
		resp.Diagnostics.Append(plan.Scopes.ElementsAs(ctx, &v, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		in.Scopes = v
	}

	cfg, err := r.client.UpdateOAuthConfig(ctx, state.ProjectID.ValueString(), state.Provider.ValueString(), in)
	if err != nil {
		resp.Diagnostics.AddError("Error updating OAuth config", err.Error())
		return
	}

	model := oauthConfigModelFromAPI(ctx, state.ProjectID.ValueString(), cfg, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *oauthConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state oauthConfigResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteOAuthConfig(ctx, state.ProjectID.ValueString(), state.Provider.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting OAuth config", err.Error())
	}
}

func (r *oauthConfigResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importStateCompositeID(ctx, req, resp, "project_id", "oauth_provider")
}
