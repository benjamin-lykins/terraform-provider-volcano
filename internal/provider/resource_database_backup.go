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
	_ resource.Resource                = &databaseBackupResource{}
	_ resource.ResourceWithConfigure   = &databaseBackupResource{}
	_ resource.ResourceWithImportState = &databaseBackupResource{}
)

func NewDatabaseBackupResource() resource.Resource {
	return &databaseBackupResource{}
}

type databaseBackupResource struct {
	client *client.Client
}

type databaseBackupResourceModel struct {
	ProjectID    types.String `tfsdk:"project_id"`
	DatabaseName types.String `tfsdk:"database_name"`
	Name         types.String `tfsdk:"name"`
	Source       types.String `tfsdk:"source"`
	SizeBytes    types.Int64  `tfsdk:"size_bytes"`
	CreatedAt    types.String `tfsdk:"created_at"`
	ExpiresAt    types.String `tfsdk:"expires_at"`
}

func (r *databaseBackupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_database_backup"
}

func (r *databaseBackupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "An on-demand named backup (snapshot) of a database.",
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"database_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"source": schema.StringAttribute{
				Computed: true,
			},
			"size_bytes": schema.Int64Attribute{
				Computed: true,
			},
			"created_at": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"expires_at": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (r *databaseBackupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *databaseBackupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan databaseBackupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	backup, err := r.client.CreateDatabaseBackup(ctx, plan.ProjectID.ValueString(), plan.DatabaseName.ValueString(), client.CreateDatabaseBackupRequest{
		Name: plan.Name.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating database backup", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, databaseBackupModelFromAPI(plan.ProjectID.ValueString(), plan.DatabaseName.ValueString(), backup))...)
}

func (r *databaseBackupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state databaseBackupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	backup, err := r.client.GetDatabaseBackup(ctx, state.ProjectID.ValueString(), state.DatabaseName.ValueString(), state.Name.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading database backup", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, databaseBackupModelFromAPI(state.ProjectID.ValueString(), state.DatabaseName.ValueString(), backup))...)
}

func (r *databaseBackupResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Database backups cannot be updated in place",
		"Every attribute forces replacement; Update should never be invoked. This is a provider bug if it was.",
	)
}

func (r *databaseBackupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state databaseBackupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteDatabaseBackup(ctx, state.ProjectID.ValueString(), state.DatabaseName.ValueString(), state.Name.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting database backup", err.Error())
	}
}

func (r *databaseBackupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importStateCompositeID(ctx, req, resp, "project_id", "database_name", "name")
}
