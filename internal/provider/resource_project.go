package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/benjamin-lykins/terraform-provider-volcano/internal/client"
)

var (
	_ resource.Resource                = &projectResource{}
	_ resource.ResourceWithConfigure   = &projectResource{}
	_ resource.ResourceWithImportState = &projectResource{}
)

func NewProjectResource() resource.Resource {
	return &projectResource{}
}

type projectResource struct {
	client *client.Client
}

type projectResourceModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Status          types.String `tfsdk:"status"`
	Plan            types.String `tfsdk:"plan"`
	AllRegions      types.Bool   `tfsdk:"all_regions"`
	SelectedRegions types.List   `tfsdk:"selected_regions"`
	TemplateID      types.String `tfsdk:"template_id"`
	InitialPrompt   types.String `tfsdk:"initial_prompt"`
	LogoURL         types.String `tfsdk:"logo_url"`
	CreatedAt       types.String `tfsdk:"created_at"`
	UpdatedAt       types.String `tfsdk:"updated_at"`
}

func (r *projectResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (r *projectResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A Volcano project: the top-level container for databases, functions, frontends, storage, and auth configuration.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Project UUID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Project name. Must be unique on the account and may only contain letters, numbers, underscores, and hyphens.",
			},
			"all_regions": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "If true (default), project functions deploy to every configured platform region. If false, only to `selected_regions`.",
			},
			"selected_regions": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Region subset to deploy to when `all_regions` is false.",
			},
			"template_id": schema.StringAttribute{
				Optional:    true,
				Description: "Starter template to install on creation (`pixel-board`, `trellini`, `collab-pad`, or `official-starter`). Cannot be combined with `initial_prompt`. Forces replacement if changed.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"initial_prompt": schema.StringAttribute{
				Optional:    true,
				Description: "Initial builder prompt used on creation. Cannot be combined with `template_id`. Forces replacement if changed.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "Project status (`active`, `deleting`, or `failed`).",
			},
			"plan": schema.StringAttribute{
				Computed:    true,
				Description: "Billing plan applied to the project, when available.",
			},
			"logo_url": schema.StringAttribute{
				Computed:    true,
				Description: "Relative API path serving the project logo, when one has been uploaded.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Creation timestamp (RFC 3339).",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "Last update timestamp (RFC 3339).",
			},
		},
	}
}

func (r *projectResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *projectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan projectResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := client.CreateProjectRequest{
		Name:          plan.Name.ValueString(),
		TemplateID:    plan.TemplateID.ValueString(),
		InitialPrompt: plan.InitialPrompt.ValueString(),
	}
	if !plan.AllRegions.IsNull() && !plan.AllRegions.IsUnknown() {
		v := plan.AllRegions.ValueBool()
		in.AllRegions = &v
	}
	if !plan.SelectedRegions.IsNull() && !plan.SelectedRegions.IsUnknown() {
		var regions []string
		resp.Diagnostics.Append(plan.SelectedRegions.ElementsAs(ctx, &regions, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		in.SelectedRegions = regions
	}

	project, err := r.client.CreateProject(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError("Error creating project", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, projectModelFromAPI(ctx, project, plan.TemplateID, plan.InitialPrompt, &resp.Diagnostics))...)
}

func (r *projectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state projectResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	project, err := r.client.GetProject(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading project", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, projectModelFromAPI(ctx, project, state.TemplateID, state.InitialPrompt, &resp.Diagnostics))...)
}

func (r *projectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state projectResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := client.UpdateProjectRequest{}
	if !plan.Name.Equal(state.Name) {
		v := plan.Name.ValueString()
		in.Name = &v
	}
	if !plan.AllRegions.Equal(state.AllRegions) {
		v := plan.AllRegions.ValueBool()
		in.AllRegions = &v
	}
	if !plan.SelectedRegions.Equal(state.SelectedRegions) {
		var regions []string
		resp.Diagnostics.Append(plan.SelectedRegions.ElementsAs(ctx, &regions, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		in.SelectedRegions = regions
	}

	project, err := r.client.UpdateProject(ctx, state.ID.ValueString(), in)
	if err != nil {
		resp.Diagnostics.AddError("Error updating project", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, projectModelFromAPI(ctx, project, plan.TemplateID, plan.InitialPrompt, &resp.Diagnostics))...)
}

func (r *projectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state projectResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteProject(ctx, state.ID.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting project", err.Error())
	}
}

func (r *projectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, pathRootID, req, resp)
}
