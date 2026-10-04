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
	_ resource.Resource                = &realtimeConfigResource{}
	_ resource.ResourceWithConfigure   = &realtimeConfigResource{}
	_ resource.ResourceWithImportState = &realtimeConfigResource{}
)

func NewRealtimeConfigResource() resource.Resource {
	return &realtimeConfigResource{}
}

type realtimeConfigResource struct {
	client *client.Client
}

type realtimeConfigResourceModel struct {
	ProjectID              types.String `tfsdk:"project_id"`
	Enabled                types.Bool   `tfsdk:"enabled"`
	BroadcastEnabled       types.Bool   `tfsdk:"broadcast_enabled"`
	PresenceEnabled        types.Bool   `tfsdk:"presence_enabled"`
	PostgresChangesEnabled types.Bool   `tfsdk:"postgres_changes_enabled"`
}

func (r *realtimeConfigResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_realtime_config"
}

func (r *realtimeConfigResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A project's realtime (WebSocket) configuration. Singleton - no create/delete endpoint; Create performs the first PUT and Delete is a local no-op.",
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"enabled":                  schema.BoolAttribute{Optional: true, Computed: true},
			"broadcast_enabled":        schema.BoolAttribute{Optional: true, Computed: true},
			"presence_enabled":         schema.BoolAttribute{Optional: true, Computed: true},
			"postgres_changes_enabled": schema.BoolAttribute{Optional: true, Computed: true},
		},
	}
}

func (r *realtimeConfigResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func realtimeConfigUpdateRequestFromModel(plan realtimeConfigResourceModel) client.UpdateRealtimeConfigRequest {
	return client.UpdateRealtimeConfigRequest{
		Enabled:                boolPtrFromModel(plan.Enabled),
		BroadcastEnabled:       boolPtrFromModel(plan.BroadcastEnabled),
		PresenceEnabled:        boolPtrFromModel(plan.PresenceEnabled),
		PostgresChangesEnabled: boolPtrFromModel(plan.PostgresChangesEnabled),
	}
}

func (r *realtimeConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan realtimeConfigResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg, err := r.client.UpdateRealtimeConfig(ctx, plan.ProjectID.ValueString(), realtimeConfigUpdateRequestFromModel(plan))
	if err != nil {
		resp.Diagnostics.AddError("Error setting realtime config", err.Error())
		return
	}

	model := realtimeConfigModelFromAPI(plan.ProjectID.ValueString(), cfg)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *realtimeConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state realtimeConfigResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg, err := r.client.GetRealtimeConfig(ctx, state.ProjectID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading realtime config", err.Error())
		return
	}

	model := realtimeConfigModelFromAPI(state.ProjectID.ValueString(), cfg)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *realtimeConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan realtimeConfigResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg, err := r.client.UpdateRealtimeConfig(ctx, plan.ProjectID.ValueString(), realtimeConfigUpdateRequestFromModel(plan))
	if err != nil {
		resp.Diagnostics.AddError("Error updating realtime config", err.Error())
		return
	}

	model := realtimeConfigModelFromAPI(plan.ProjectID.ValueString(), cfg)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *realtimeConfigResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// No-op: no delete endpoint exists for this singleton.
}

func (r *realtimeConfigResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, pathProjectID, req.ID)...)
}
