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
	_ resource.Resource                = &projectAuthHostedPageResource{}
	_ resource.ResourceWithConfigure   = &projectAuthHostedPageResource{}
	_ resource.ResourceWithImportState = &projectAuthHostedPageResource{}
)

func NewProjectAuthHostedPageResource() resource.Resource {
	return &projectAuthHostedPageResource{}
}

type projectAuthHostedPageResource struct {
	client *client.Client
}

type projectAuthHostedPageResourceModel struct {
	ProjectID types.String `tfsdk:"project_id"`
	PageType  types.String `tfsdk:"page_type"`
	HTML      types.String `tfsdk:"html"`
	CSS       types.String `tfsdk:"css"`
}

func (r *projectAuthHostedPageResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_auth_hosted_page"
}

func (r *projectAuthHostedPageResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Custom HTML/CSS override for one of a project's hosted auth pages. There is no delete endpoint; Delete only drops it from Terraform state, it does not revert the page to the platform default.",
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"page_type": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.OneOf("login", "signup", "forgot-password", "device", "verify-email", "reset-password"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"html": schema.StringAttribute{
				Required: true,
			},
			"css": schema.StringAttribute{
				Optional: true,
			},
		},
	}
}

func (r *projectAuthHostedPageResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *projectAuthHostedPageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan projectAuthHostedPageResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	page, err := r.client.SetProjectAuthHostedPage(ctx, plan.ProjectID.ValueString(), plan.PageType.ValueString(), plan.HTML.ValueString(), plan.CSS.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error setting project auth hosted page", err.Error())
		return
	}

	model := projectAuthHostedPageModelFromAPI(plan.ProjectID.ValueString(), plan.PageType.ValueString(), page)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *projectAuthHostedPageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state projectAuthHostedPageResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	page, err := r.client.GetProjectAuthHostedPage(ctx, state.ProjectID.ValueString(), state.PageType.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading project auth hosted page", err.Error())
		return
	}

	model := projectAuthHostedPageModelFromAPI(state.ProjectID.ValueString(), state.PageType.ValueString(), page)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *projectAuthHostedPageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan projectAuthHostedPageResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	page, err := r.client.SetProjectAuthHostedPage(ctx, plan.ProjectID.ValueString(), plan.PageType.ValueString(), plan.HTML.ValueString(), plan.CSS.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error updating project auth hosted page", err.Error())
		return
	}

	model := projectAuthHostedPageModelFromAPI(plan.ProjectID.ValueString(), plan.PageType.ValueString(), page)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *projectAuthHostedPageResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// No-op: there is no delete endpoint for a hosted page override.
}

func (r *projectAuthHostedPageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importStateCompositeID(ctx, req, resp, "project_id", "page_type")
}
