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
	_ resource.Resource                = &projectGitConnectionResource{}
	_ resource.ResourceWithConfigure   = &projectGitConnectionResource{}
	_ resource.ResourceWithImportState = &projectGitConnectionResource{}
)

func NewProjectGitConnectionResource() resource.Resource {
	return &projectGitConnectionResource{}
}

type projectGitConnectionResource struct {
	client *client.Client
}

type projectGitConnectionResourceModel struct {
	ProjectID        types.String `tfsdk:"project_id"`
	ConnectionID     types.String `tfsdk:"connection_id"`
	InstallationID   types.Int64  `tfsdk:"installation_id"`
	RepositoryID     types.Int64  `tfsdk:"repository_id"`
	RepoFullName     types.String `tfsdk:"repo_full_name"`
	RootDirectory    types.String `tfsdk:"root_directory"`
	ProductionBranch types.String `tfsdk:"production_branch"`
	UpdatedAt        types.String `tfsdk:"updated_at"`
}

func (r *projectGitConnectionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_git_connection"
}

func (r *projectGitConnectionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Ties a project to a repository reachable through one of the account's git connections (see the volcano_user_git_connections and volcano_user_git_installations data sources). Singleton per project - no create/delete endpoint; Create performs the first PUT and Delete calls the dedicated DELETE to unlink.",
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"connection_id": schema.StringAttribute{
				Required: true,
			},
			"installation_id": schema.Int64Attribute{
				Required: true,
			},
			"repository_id": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"repo_full_name": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"root_directory": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"production_branch": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Mutable in place via the dedicated production-branch endpoint.",
			},
			"updated_at": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (r *projectGitConnectionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *projectGitConnectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan projectGitConnectionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	gc, err := r.client.SetProjectGitConnection(ctx, plan.ProjectID.ValueString(), client.SetProjectGitConnectionRequest{
		ConnectionID:     plan.ConnectionID.ValueString(),
		InstallationID:   plan.InstallationID.ValueInt64(),
		RepositoryID:     plan.RepositoryID.ValueInt64(),
		RepoFullName:     plan.RepoFullName.ValueString(),
		RootDirectory:    plan.RootDirectory.ValueString(),
		ProductionBranch: plan.ProductionBranch.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error setting project git connection", err.Error())
		return
	}

	model := projectGitConnectionModelFromAPI(plan.ProjectID.ValueString(), plan.ConnectionID, gc)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *projectGitConnectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state projectGitConnectionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	gc, err := r.client.GetProjectGitConnection(ctx, state.ProjectID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading project git connection", err.Error())
		return
	}

	model := projectGitConnectionModelFromAPI(state.ProjectID.ValueString(), state.ConnectionID, gc)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *projectGitConnectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state projectGitConnectionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var gc *client.ProjectGitConnection
	var err error
	if !plan.ProductionBranch.Equal(state.ProductionBranch) {
		gc, err = r.client.SetProjectGitProductionBranch(ctx, state.ProjectID.ValueString(), plan.ProductionBranch.ValueString())
	} else {
		gc, err = r.client.SetProjectGitConnection(ctx, state.ProjectID.ValueString(), client.SetProjectGitConnectionRequest{
			ConnectionID:     plan.ConnectionID.ValueString(),
			InstallationID:   plan.InstallationID.ValueInt64(),
			RepositoryID:     plan.RepositoryID.ValueInt64(),
			RepoFullName:     plan.RepoFullName.ValueString(),
			RootDirectory:    plan.RootDirectory.ValueString(),
			ProductionBranch: plan.ProductionBranch.ValueString(),
		})
	}
	if err != nil {
		resp.Diagnostics.AddError("Error updating project git connection", err.Error())
		return
	}

	model := projectGitConnectionModelFromAPI(state.ProjectID.ValueString(), plan.ConnectionID, gc)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *projectGitConnectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state projectGitConnectionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteProjectGitConnection(ctx, state.ProjectID.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting project git connection", err.Error())
	}
}

func (r *projectGitConnectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, pathProjectID, req.ID)...)
}
