package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/benjamin-lykins/terraform-provider-volcano/internal/client"
)

var (
	_ resource.Resource                = &databaseBranchResource{}
	_ resource.ResourceWithConfigure   = &databaseBranchResource{}
	_ resource.ResourceWithImportState = &databaseBranchResource{}
)

func NewDatabaseBranchResource() resource.Resource {
	return &databaseBranchResource{}
}

type databaseBranchResource struct {
	client *client.Client
}

type databaseBranchResourceModel struct {
	ID               types.String `tfsdk:"id"`
	ProjectID        types.String `tfsdk:"project_id"`
	DatabaseName     types.String `tfsdk:"database_name"`
	Name             types.String `tfsdk:"name"`
	TTLSeconds       types.Int64  `tfsdk:"ttl_seconds"`
	Status           types.String `tfsdk:"status"`
	ConnectionString types.String `tfsdk:"connection_string"`
	ExpiresAt        types.String `tfsdk:"expires_at"`
	CreatedAt        types.String `tfsdk:"created_at"`
}

func (r *databaseBranchResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_database_branch"
}

func (r *databaseBranchResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A short-lived fork of a database, for development or testing. Expires automatically after ttl_seconds.",
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
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ttl_seconds": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Lifetime of the branch in seconds. Mutable in place.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"status": schema.StringAttribute{
				Computed: true,
			},
			"connection_string": schema.StringAttribute{
				Computed:  true,
				Sensitive: true,
			},
			"expires_at": schema.StringAttribute{
				Computed: true,
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

func (r *databaseBranchResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *databaseBranchResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan databaseBranchResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := client.CreateDatabaseBranchRequest{Name: plan.Name.ValueString()}
	if !plan.TTLSeconds.IsNull() && !plan.TTLSeconds.IsUnknown() {
		v := plan.TTLSeconds.ValueInt64()
		in.TTLSeconds = &v
	}

	branch, err := r.client.CreateDatabaseBranch(ctx, plan.ProjectID.ValueString(), plan.DatabaseName.ValueString(), in)
	if err != nil {
		resp.Diagnostics.AddError("Error creating database branch", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, databaseBranchModelFromAPI(plan.ProjectID.ValueString(), plan.DatabaseName.ValueString(), branch))...)
}

func (r *databaseBranchResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state databaseBranchResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	branch, err := r.client.GetDatabaseBranch(ctx, state.ProjectID.ValueString(), state.DatabaseName.ValueString(), state.Name.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading database branch", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, databaseBranchModelFromAPI(state.ProjectID.ValueString(), state.DatabaseName.ValueString(), branch))...)
}

func (r *databaseBranchResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan databaseBranchResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	branch, err := r.client.UpdateDatabaseBranch(ctx, plan.ProjectID.ValueString(), plan.DatabaseName.ValueString(), plan.Name.ValueString(), client.UpdateDatabaseBranchRequest{
		TTLSeconds: plan.TTLSeconds.ValueInt64(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating database branch", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, databaseBranchModelFromAPI(plan.ProjectID.ValueString(), plan.DatabaseName.ValueString(), branch))...)
}

func (r *databaseBranchResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state databaseBranchResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteDatabaseBranch(ctx, state.ProjectID.ValueString(), state.DatabaseName.ValueString(), state.Name.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting database branch", err.Error())
	}
}

func (r *databaseBranchResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importStateCompositeID(ctx, req, resp, "project_id", "database_name", "name")
}
