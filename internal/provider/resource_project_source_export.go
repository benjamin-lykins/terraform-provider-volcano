package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/benjamin-lykins/terraform-provider-volcano/internal/client"
)

var (
	_ resource.Resource                = &projectSourceExportResource{}
	_ resource.ResourceWithConfigure   = &projectSourceExportResource{}
	_ resource.ResourceWithImportState = &projectSourceExportResource{}
)

func NewProjectSourceExportResource() resource.Resource {
	return &projectSourceExportResource{}
}

type projectSourceExportResource struct {
	client *client.Client
}

type projectSourceExportResourceModel struct {
	ProjectID           types.String `tfsdk:"project_id"`
	ProductionBranch    types.String `tfsdk:"production_branch"`
	RepoFullName        types.String `tfsdk:"repo_full_name"`
	Branch              types.String `tfsdk:"branch"`
	CommitSHA           types.String `tfsdk:"commit_sha"`
	FileCount           types.Int64  `tfsdk:"file_count"`
	Skipped             types.List   `tfsdk:"skipped"`
	Omitted             types.List   `tfsdk:"omitted"`
	Mode                types.String `tfsdk:"mode"`
	ExportedAt          types.String `tfsdk:"exported_at"`
	TransitionStartedAt types.String `tfsdk:"transition_started_at"`
	HandedOverAt        types.String `tfsdk:"handed_over_at"`
}

func (r *projectSourceExportResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_source_export"
}

func (r *projectSourceExportResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Triggers a one-shot export of a project's source to a git repository (a git takeover). This is a one-shot operation modeled as create+read+delete: Delete cancels/clears the export record via the API's own DELETE, it does not undo an already-completed handover.",
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"production_branch": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"repo_full_name": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"branch": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"commit_sha": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"file_count": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"skipped": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"omitted": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"mode": schema.StringAttribute{
				Computed: true,
			},
			"exported_at": schema.StringAttribute{
				Computed: true,
			},
			"transition_started_at": schema.StringAttribute{
				Computed: true,
			},
			"handed_over_at": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (r *projectSourceExportResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *projectSourceExportResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan projectSourceExportResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.CreateProjectSourceExport(ctx, plan.ProjectID.ValueString(), client.CreateProjectSourceExportRequest{
		ProductionBranch: plan.ProductionBranch.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating project source export", err.Error())
		return
	}

	status, err := r.client.GetProjectSourceExportStatus(ctx, plan.ProjectID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading project source export status", err.Error())
		return
	}

	model := projectSourceExportModelFromAPI(ctx, plan.ProjectID.ValueString(), plan.ProductionBranch, result, status, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *projectSourceExportResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state projectSourceExportResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	status, err := r.client.GetProjectSourceExportStatus(ctx, state.ProjectID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading project source export status", err.Error())
		return
	}

	model := projectSourceExportModelFromAPI(ctx, state.ProjectID.ValueString(), state.ProductionBranch, nil, status, &resp.Diagnostics)
	// Preserve the one-time create result fields (not re-derivable from the status endpoint).
	model.RepoFullName = state.RepoFullName
	model.Branch = state.Branch
	model.CommitSHA = state.CommitSHA
	model.FileCount = state.FileCount
	model.Skipped = state.Skipped
	model.Omitted = state.Omitted
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *projectSourceExportResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Project source exports cannot be updated in place",
		"Every attribute forces replacement; Update should never be invoked. This is a provider bug if it was.",
	)
}

func (r *projectSourceExportResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state projectSourceExportResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteProjectSourceExport(ctx, state.ProjectID.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting project source export", err.Error())
	}
}

func (r *projectSourceExportResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, pathProjectID, req.ID)...)
}
