package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/benjamin-lykins/terraform-provider-volcano/internal/client"
)

var (
	_ resource.Resource                = &functionSchedulerResource{}
	_ resource.ResourceWithConfigure   = &functionSchedulerResource{}
	_ resource.ResourceWithImportState = &functionSchedulerResource{}
)

func NewFunctionSchedulerResource() resource.Resource {
	return &functionSchedulerResource{}
}

type functionSchedulerResource struct {
	client *client.Client
}

type functionSchedulerResourceModel struct {
	ID             types.String `tfsdk:"id"`
	ProjectID      types.String `tfsdk:"project_id"`
	FunctionID     types.String `tfsdk:"function_id"`
	Name           types.String `tfsdk:"name"`
	Enabled        types.Bool   `tfsdk:"enabled"`
	CronExpression types.String `tfsdk:"cron_expression"`
	Payload        types.String `tfsdk:"payload"`
	Regions        types.List   `tfsdk:"regions"`
	RunCount       types.Int64  `tfsdk:"run_count"`
	NextRunAt      types.String `tfsdk:"next_run_at"`
	CreatedAt      types.String `tfsdk:"created_at"`
}

func (r *functionSchedulerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_function_scheduler"
}

func (r *functionSchedulerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A cron trigger attached to a function.",
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
			"function_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
			},
			"cron_expression": schema.StringAttribute{
				Required:    true,
				Description: "Standard 5-field cron expression, evaluated in UTC.",
			},
			"payload": schema.StringAttribute{
				Optional:    true,
				Description: "JSON-encoded object passed to the function on each scheduled invocation, e.g. jsonencode({ foo = \"bar\" }).",
			},
			"regions": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "At most one explicit deployed region. If omitted, the scheduler picks one deployed region itself.",
			},
			"run_count": schema.Int64Attribute{
				Computed: true,
			},
			"next_run_at": schema.StringAttribute{
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

func (r *functionSchedulerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *functionSchedulerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan functionSchedulerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in, ok := schedulerCreateRequestFromModel(ctx, plan.Name, plan.Enabled, plan.CronExpression, plan.Payload, plan.Regions, &resp.Diagnostics)
	if !ok {
		return
	}

	s, err := r.client.CreateFunctionScheduler(ctx, plan.ProjectID.ValueString(), plan.FunctionID.ValueString(), in)
	if err != nil {
		resp.Diagnostics.AddError("Error creating function scheduler", err.Error())
		return
	}

	model, diags := functionSchedulerModelFromAPI(ctx, plan.ProjectID.ValueString(), plan.FunctionID.ValueString(), s)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *functionSchedulerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state functionSchedulerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	s, err := r.client.GetFunctionScheduler(ctx, state.ProjectID.ValueString(), state.FunctionID.ValueString(), state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading function scheduler", err.Error())
		return
	}

	model, diags := functionSchedulerModelFromAPI(ctx, state.ProjectID.ValueString(), state.FunctionID.ValueString(), s)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *functionSchedulerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state functionSchedulerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in, ok := schedulerUpdateRequestFromModel(ctx, plan, state, &resp.Diagnostics)
	if !ok {
		return
	}

	s, err := r.client.UpdateFunctionScheduler(ctx, state.ProjectID.ValueString(), state.FunctionID.ValueString(), state.ID.ValueString(), in)
	if err != nil {
		resp.Diagnostics.AddError("Error updating function scheduler", err.Error())
		return
	}

	model, diags := functionSchedulerModelFromAPI(ctx, state.ProjectID.ValueString(), state.FunctionID.ValueString(), s)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *functionSchedulerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state functionSchedulerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteFunctionScheduler(ctx, state.ProjectID.ValueString(), state.FunctionID.ValueString(), state.ID.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting function scheduler", err.Error())
	}
}

func (r *functionSchedulerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importStateCompositeID(ctx, req, resp, "project_id", "function_id", "id")
}
