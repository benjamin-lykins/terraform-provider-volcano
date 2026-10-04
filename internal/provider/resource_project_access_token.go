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
	_ resource.Resource                = &projectAccessTokenResource{}
	_ resource.ResourceWithConfigure   = &projectAccessTokenResource{}
	_ resource.ResourceWithImportState = &projectAccessTokenResource{}
)

func NewProjectAccessTokenResource() resource.Resource {
	return &projectAccessTokenResource{}
}

type projectAccessTokenResource struct {
	client *client.Client
}

type projectAccessTokenResourceModel struct {
	ID        types.String `tfsdk:"id"`
	ProjectID types.String `tfsdk:"project_id"`
	Name      types.String `tfsdk:"name"`
	Scope     types.String `tfsdk:"scope"`
	ExpiresAt types.String `tfsdk:"expires_at"`
	Status    types.String `tfsdk:"status"`
	Token     types.String `tfsdk:"token"`
	CreatedAt types.String `tfsdk:"created_at"`
}

func (r *projectAccessTokenResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_access_token"
}

func (r *projectAccessTokenResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A project access token: a control-plane credential scoped to a single project, suitable for CI, scripts, or agents. The plaintext secret (`token`) is returned only on creation and cannot be recovered afterwards.",
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
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Unique per project. Forces replacement if changed.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"scope": schema.StringAttribute{
				Required:    true,
				Description: "`full` (everything except managing access tokens) or `read_only`. Forces replacement if changed.",
				Validators: []validator.String{
					stringvalidator.OneOf("full", "read_only"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"expires_at": schema.StringAttribute{
				Optional:    true,
				Description: "RFC 3339 expiry timestamp. Omit for a token that never expires. Forces replacement if changed.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "`active`, `revoked`, or `expired`.",
			},
			"token": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "The plaintext secret. Only ever populated from the create response; the API never returns it again, so this value is preserved as-is across Read.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *projectAccessTokenResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *projectAccessTokenResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan projectAccessTokenResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateProjectAccessToken(ctx, plan.ProjectID.ValueString(), client.CreateProjectAccessTokenRequest{
		Name:      plan.Name.ValueString(),
		Scope:     plan.Scope.ValueString(),
		ExpiresAt: plan.ExpiresAt.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating project access token", err.Error())
		return
	}

	plan.ID = types.StringValue(created.ID)
	plan.Status = types.StringValue(created.Status)
	plan.Token = types.StringValue(created.Token)
	plan.CreatedAt = types.StringValue(created.CreatedAt)
	plan.ExpiresAt = stringOrNull(created.ExpiresAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *projectAccessTokenResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state projectAccessTokenResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	token, err := r.client.GetProjectAccessToken(ctx, state.ProjectID.ValueString(), state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading project access token", err.Error())
		return
	}

	state.Name = types.StringValue(token.Name)
	state.Scope = types.StringValue(token.Scope)
	state.Status = types.StringValue(token.Status)
	state.ExpiresAt = stringOrNull(token.ExpiresAt)
	state.CreatedAt = types.StringValue(token.CreatedAt)
	// token.Token is intentionally left untouched: the API never returns
	// the plaintext secret again after creation.

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *projectAccessTokenResource) Update(ctx context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Project access tokens cannot be updated in place",
		"Every mutable attribute forces replacement; Update should never be invoked. This is a provider bug if it was.",
	)
}

func (r *projectAccessTokenResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state projectAccessTokenResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteProjectAccessToken(ctx, state.ProjectID.ValueString(), state.ID.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting project access token", err.Error())
	}
}

func (r *projectAccessTokenResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importStateCompositeID(ctx, req, resp, "project_id", "id")
}
