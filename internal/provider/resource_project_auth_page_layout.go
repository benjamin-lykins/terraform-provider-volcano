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
	_ resource.Resource                = &projectAuthPageLayoutResource{}
	_ resource.ResourceWithConfigure   = &projectAuthPageLayoutResource{}
	_ resource.ResourceWithImportState = &projectAuthPageLayoutResource{}
)

func NewProjectAuthPageLayoutResource() resource.Resource {
	return &projectAuthPageLayoutResource{}
}

type projectAuthPageLayoutResource struct {
	client *client.Client
}

type projectAuthPageLayoutResourceModel struct {
	ProjectID types.String `tfsdk:"project_id"`
	PageType  types.String `tfsdk:"page_type"`
	Layout    types.String `tfsdk:"layout"`
}

func (r *projectAuthPageLayoutResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_auth_page_layout"
}

func (r *projectAuthPageLayoutResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The layout for one of a project's hosted auth pages. Read back via the appearance endpoint since this has no GET of its own.",
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
			"layout": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.OneOf("centered", "split-left", "split-right"),
				},
			},
		},
	}
}

func (r *projectAuthPageLayoutResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *projectAuthPageLayoutResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan projectAuthPageLayoutResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	layout, err := r.client.SetProjectAuthPageLayout(ctx, plan.ProjectID.ValueString(), plan.PageType.ValueString(), plan.Layout.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error setting project auth page layout", err.Error())
		return
	}

	plan.Layout = types.StringValue(layout)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *projectAuthPageLayoutResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state projectAuthPageLayoutResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	appearance, err := r.client.GetProjectAuthPagesAppearance(ctx, state.ProjectID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading project auth page layout", err.Error())
		return
	}

	layout, ok := appearance.Layouts[state.PageType.ValueString()]
	if !ok {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Layout = types.StringValue(layout)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *projectAuthPageLayoutResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan projectAuthPageLayoutResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	layout, err := r.client.SetProjectAuthPageLayout(ctx, plan.ProjectID.ValueString(), plan.PageType.ValueString(), plan.Layout.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error updating project auth page layout", err.Error())
		return
	}

	plan.Layout = types.StringValue(layout)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *projectAuthPageLayoutResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state projectAuthPageLayoutResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteProjectAuthPageLayout(ctx, state.ProjectID.ValueString(), state.PageType.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting project auth page layout", err.Error())
	}
}

func (r *projectAuthPageLayoutResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importStateCompositeID(ctx, req, resp, "project_id", "page_type")
}
