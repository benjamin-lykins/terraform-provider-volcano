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
	_ resource.Resource                = &projectGitDeploySettingsResource{}
	_ resource.ResourceWithConfigure   = &projectGitDeploySettingsResource{}
	_ resource.ResourceWithImportState = &projectGitDeploySettingsResource{}
)

func NewProjectGitDeploySettingsResource() resource.Resource {
	return &projectGitDeploySettingsResource{}
}

type projectGitDeploySettingsResource struct {
	client *client.Client
}

type projectGitDeploySettingsResourceModel struct {
	ProjectID         types.String `tfsdk:"project_id"`
	AutoDeployEnabled types.Bool   `tfsdk:"auto_deploy_enabled"`
	DeployFunctions   types.Bool   `tfsdk:"deploy_functions"`
	FrontendName      types.String `tfsdk:"frontend_name"`
	FrontendAppRoot   types.String `tfsdk:"frontend_app_root"`
	UpdatedAt         types.String `tfsdk:"updated_at"`
}

func (r *projectGitDeploySettingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_git_deploy_settings"
}

func (r *projectGitDeploySettingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Controls what a project's git-connected deploys build. Singleton - no create/delete endpoint; Create performs the first PUT and Delete is a local no-op.",
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"auto_deploy_enabled": schema.BoolAttribute{
				Required: true,
			},
			"deploy_functions": schema.BoolAttribute{
				Required: true,
			},
			"frontend_name": schema.StringAttribute{
				Optional:    true,
				Description: "Name of the frontend to build from this repository, if any.",
			},
			"frontend_app_root": schema.StringAttribute{
				Optional: true,
			},
			"updated_at": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (r *projectGitDeploySettingsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *projectGitDeploySettingsResource) apply(ctx context.Context, plan projectGitDeploySettingsResourceModel) client.ProjectGitDeploySettings {
	return client.ProjectGitDeploySettings{
		AutoDeployEnabled: plan.AutoDeployEnabled.ValueBool(),
		DeployFunctions:   plan.DeployFunctions.ValueBool(),
		FrontendName:      plan.FrontendName.ValueString(),
		FrontendAppRoot:   plan.FrontendAppRoot.ValueString(),
	}
}

func (r *projectGitDeploySettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan projectGitDeploySettingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	settings, err := r.client.SetProjectGitDeploySettings(ctx, plan.ProjectID.ValueString(), r.apply(ctx, plan))
	if err != nil {
		resp.Diagnostics.AddError("Error setting project git deploy settings", err.Error())
		return
	}

	model := projectGitDeploySettingsModelFromAPI(plan.ProjectID.ValueString(), settings)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *projectGitDeploySettingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state projectGitDeploySettingsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	settings, err := r.client.GetProjectGitDeploySettings(ctx, state.ProjectID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading project git deploy settings", err.Error())
		return
	}

	model := projectGitDeploySettingsModelFromAPI(state.ProjectID.ValueString(), settings)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *projectGitDeploySettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan projectGitDeploySettingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	settings, err := r.client.SetProjectGitDeploySettings(ctx, plan.ProjectID.ValueString(), r.apply(ctx, plan))
	if err != nil {
		resp.Diagnostics.AddError("Error updating project git deploy settings", err.Error())
		return
	}

	model := projectGitDeploySettingsModelFromAPI(plan.ProjectID.ValueString(), settings)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *projectGitDeploySettingsResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// No-op: no delete endpoint exists for this singleton.
}

func (r *projectGitDeploySettingsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, pathProjectID, req.ID)...)
}
