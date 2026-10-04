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
	_ resource.Resource                = &databaseBackupScheduleResource{}
	_ resource.ResourceWithConfigure   = &databaseBackupScheduleResource{}
	_ resource.ResourceWithImportState = &databaseBackupScheduleResource{}
)

func NewDatabaseBackupScheduleResource() resource.Resource {
	return &databaseBackupScheduleResource{}
}

type databaseBackupScheduleResource struct {
	client *client.Client
}

type backupScheduleEntryModel struct {
	Frequency        types.String `tfsdk:"frequency"`
	Hour             types.Int64  `tfsdk:"hour"`
	Day              types.Int64  `tfsdk:"day"`
	RetentionSeconds types.Int64  `tfsdk:"retention_seconds"`
}

type databaseBackupScheduleResourceModel struct {
	ProjectID    types.String               `tfsdk:"project_id"`
	DatabaseName types.String               `tfsdk:"database_name"`
	Entries      []backupScheduleEntryModel `tfsdk:"entries"`
}

func (r *databaseBackupScheduleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_database_backup_schedule"
}

func (r *databaseBackupScheduleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A database's automated backup schedule. This is a singleton per database: creating it sets the schedule, deleting it clears it (an empty entry list).",
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
			"entries": schema.ListNestedAttribute{
				Required:    true,
				Description: "One or more recurrences. An empty list clears the schedule.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"frequency": schema.StringAttribute{
							Required:    true,
							Description: "daily, weekly, or monthly.",
						},
						"hour": schema.Int64Attribute{
							Required:    true,
							Description: "Hour of the day in UTC (0-23).",
						},
						"day": schema.Int64Attribute{
							Optional:    true,
							Description: "Day of week (1-7, Monday-Sunday) for weekly, or day of month (1-28) for monthly. Required for both; ignored for daily.",
						},
						"retention_seconds": schema.Int64Attribute{
							Optional:    true,
							Computed:    true,
							Description: "How long backups from this recurrence are kept. Defaults to the plan's retention when omitted.",
						},
					},
				},
			},
		},
	}
}

func (r *databaseBackupScheduleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func entriesToAPI(entries []backupScheduleEntryModel) []client.BackupScheduleEntry {
	out := make([]client.BackupScheduleEntry, 0, len(entries))
	for _, e := range entries {
		entry := client.BackupScheduleEntry{
			Frequency: e.Frequency.ValueString(),
			Hour:      e.Hour.ValueInt64(),
		}
		if !e.Day.IsNull() && !e.Day.IsUnknown() {
			v := e.Day.ValueInt64()
			entry.Day = &v
		}
		if !e.RetentionSeconds.IsNull() && !e.RetentionSeconds.IsUnknown() {
			v := e.RetentionSeconds.ValueInt64()
			entry.RetentionSeconds = &v
		}
		out = append(out, entry)
	}
	return out
}

func entriesFromAPI(entries []client.BackupScheduleEntry) []backupScheduleEntryModel {
	out := make([]backupScheduleEntryModel, 0, len(entries))
	for _, e := range entries {
		m := backupScheduleEntryModel{
			Frequency: types.StringValue(e.Frequency),
			Hour:      types.Int64Value(e.Hour),
		}
		if e.Day != nil {
			m.Day = types.Int64Value(*e.Day)
		} else {
			m.Day = types.Int64Null()
		}
		if e.RetentionSeconds != nil {
			m.RetentionSeconds = types.Int64Value(*e.RetentionSeconds)
		} else {
			m.RetentionSeconds = types.Int64Null()
		}
		out = append(out, m)
	}
	return out
}

func (r *databaseBackupScheduleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan databaseBackupScheduleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	schedule, err := r.client.SetDatabaseBackupSchedule(ctx, plan.ProjectID.ValueString(), plan.DatabaseName.ValueString(), client.BackupSchedule{
		Entries: entriesToAPI(plan.Entries),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error setting database backup schedule", err.Error())
		return
	}

	plan.Entries = entriesFromAPI(schedule.Entries)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *databaseBackupScheduleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state databaseBackupScheduleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	schedule, err := r.client.GetDatabaseBackupSchedule(ctx, state.ProjectID.ValueString(), state.DatabaseName.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading database backup schedule", err.Error())
		return
	}

	state.Entries = entriesFromAPI(schedule.Entries)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *databaseBackupScheduleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan databaseBackupScheduleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	schedule, err := r.client.SetDatabaseBackupSchedule(ctx, plan.ProjectID.ValueString(), plan.DatabaseName.ValueString(), client.BackupSchedule{
		Entries: entriesToAPI(plan.Entries),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating database backup schedule", err.Error())
		return
	}

	plan.Entries = entriesFromAPI(schedule.Entries)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *databaseBackupScheduleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state databaseBackupScheduleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := r.client.SetDatabaseBackupSchedule(ctx, state.ProjectID.ValueString(), state.DatabaseName.ValueString(), client.BackupSchedule{Entries: nil}); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Error clearing database backup schedule", err.Error())
	}
}

func (r *databaseBackupScheduleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importStateCompositeID(ctx, req, resp, "project_id", "database_name")
}
