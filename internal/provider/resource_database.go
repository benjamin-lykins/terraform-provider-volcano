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
	_ resource.Resource                = &databaseResource{}
	_ resource.ResourceWithConfigure   = &databaseResource{}
	_ resource.ResourceWithImportState = &databaseResource{}
)

func NewDatabaseResource() resource.Resource {
	return &databaseResource{}
}

type databaseResource struct {
	client *client.Client
}

type databaseResourceModel struct {
	ID               types.String `tfsdk:"id"`
	ProjectID        types.String `tfsdk:"project_id"`
	Name             types.String `tfsdk:"name"`
	Region           types.String `tfsdk:"region"`
	PGVersion        types.String `tfsdk:"pg_version"`
	DatabaseType     types.String `tfsdk:"database_type"`
	Status           types.String `tfsdk:"status"`
	ConnectionString types.String `tfsdk:"connection_string"`
	CreatedAt        types.String `tfsdk:"created_at"`
	UpdatedAt        types.String `tfsdk:"updated_at"`
}

func (r *databaseResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_database"
}

func (r *databaseResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A provisioned PostgreSQL database within a Volcano project.",
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
				Description: "Lowercase letters, numbers, and underscores only. Unique within the project. Forces replacement if changed.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"region": schema.StringAttribute{
				Required:    true,
				Description: "Hosting region, e.g. us-east-1. See the volcano_database_regions data source for valid values. Forces replacement if changed.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"pg_version": schema.StringAttribute{
				Required:    true,
				Description: "PostgreSQL major version (\"15\" or \"16\"). Forces replacement if changed.",
				Validators: []validator.String{
					stringvalidator.OneOf("15", "16"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"database_type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Compute size tier. Defaults to volcano-db-xs. Mutable in place via the dedicated resize endpoint.",
				Validators: []validator.String{
					stringvalidator.OneOf("volcano-db-xs", "volcano-db-s", "volcano-db-m", "volcano-db-l", "volcano-db-xl", "volcano-db-2xl"),
				},
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "provisioning, active, failed, restoring, or deleting.",
			},
			"connection_string": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "PostgreSQL connection URI with full admin access.",
			},
			"created_at": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (r *databaseResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *databaseResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan databaseResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	db, err := r.client.CreateDatabase(ctx, plan.ProjectID.ValueString(), client.CreateDatabaseRequest{
		Name:         plan.Name.ValueString(),
		Region:       plan.Region.ValueString(),
		PGVersion:    plan.PGVersion.ValueString(),
		DatabaseType: plan.DatabaseType.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating database", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, databaseModelFromAPI(db))...)
}

func (r *databaseResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state databaseResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	db, err := r.client.GetDatabase(ctx, state.ProjectID.ValueString(), state.Name.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading database", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, databaseModelFromAPI(db))...)
}

func (r *databaseResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state databaseResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var db *client.Database
	var err error
	if !plan.DatabaseType.Equal(state.DatabaseType) {
		db, err = r.client.UpdateDatabaseType(ctx, state.ProjectID.ValueString(), state.Name.ValueString(), client.UpdateDatabaseTypeRequest{
			DatabaseType: plan.DatabaseType.ValueString(),
		})
	} else {
		db, err = r.client.GetDatabase(ctx, state.ProjectID.ValueString(), state.Name.ValueString())
	}
	if err != nil {
		resp.Diagnostics.AddError("Error updating database", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, databaseModelFromAPI(db))...)
}

func (r *databaseResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state databaseResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteDatabase(ctx, state.ProjectID.ValueString(), state.Name.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting database", err.Error())
	}
}

func (r *databaseResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importStateCompositeID(ctx, req, resp, "project_id", "name")
}
