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
	_ resource.Resource                = &databaseRestoreResource{}
	_ resource.ResourceWithConfigure   = &databaseRestoreResource{}
	_ resource.ResourceWithImportState = &databaseRestoreResource{}
)

func NewDatabaseRestoreResource() resource.Resource {
	return &databaseRestoreResource{}
}

type databaseRestoreResource struct {
	client *client.Client
}

type databaseRestoreResourceModel struct {
	ID           types.String `tfsdk:"id"`
	ProjectID    types.String `tfsdk:"project_id"`
	DatabaseName types.String `tfsdk:"database_name"`
	BackupName   types.String `tfsdk:"backup_name"`
	RestoreTo    types.String `tfsdk:"restore_to"`
	Kind         types.String `tfsdk:"kind"`
	Status       types.String `tfsdk:"status"`
	Error        types.String `tfsdk:"error"`
	CreatedAt    types.String `tfsdk:"created_at"`
	CompletedAt  types.String `tfsdk:"completed_at"`
}

func (r *databaseRestoreResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_database_restore"
}

func (r *databaseRestoreResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Triggers a one-shot restore of a database from a named backup or a point in time. This is a one-shot operation modeled as create+read only: removing it from configuration only drops it from state, it does not undo the restore (the API has no way to reverse one).",
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
			"database_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"backup_name": schema.StringAttribute{
				Optional:    true,
				Description: "Exactly one of backup_name or restore_to must be set. Forces replacement if changed.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"restore_to": schema.StringAttribute{
				Optional:    true,
				Description: "RFC 3339 point in time within the backup retention window. Forces replacement if changed.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"kind": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"status": schema.StringAttribute{
				Computed: true,
			},
			"error": schema.StringAttribute{
				Computed: true,
			},
			"created_at": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"completed_at": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (r *databaseRestoreResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *databaseRestoreResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan databaseRestoreResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	restore, err := r.client.CreateDatabaseRestore(ctx, plan.ProjectID.ValueString(), plan.DatabaseName.ValueString(), client.CreateDatabaseRestoreRequest{
		BackupName: plan.BackupName.ValueString(),
		RestoreTo:  plan.RestoreTo.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating database restore", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, databaseRestoreModelFromAPI(plan.ProjectID.ValueString(), plan.DatabaseName.ValueString(), restore))...)
}

func (r *databaseRestoreResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state databaseRestoreResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	restore, err := r.client.GetDatabaseRestore(ctx, state.ProjectID.ValueString(), state.DatabaseName.ValueString(), state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading database restore", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, databaseRestoreModelFromAPI(state.ProjectID.ValueString(), state.DatabaseName.ValueString(), restore))...)
}

func (r *databaseRestoreResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Database restores cannot be updated in place",
		"Every attribute forces replacement; Update should never be invoked. This is a provider bug if it was.",
	)
}

func (r *databaseRestoreResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// No-op: the API has no way to undo a restore. Removing this resource
	// from configuration only drops it from Terraform state.
}

func (r *databaseRestoreResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importStateCompositeID(ctx, req, resp, "project_id", "database_name", "id")
}
