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
	_ resource.Resource                = &projectVariableResource{}
	_ resource.ResourceWithConfigure   = &projectVariableResource{}
	_ resource.ResourceWithImportState = &projectVariableResource{}
)

func NewProjectVariableResource() resource.Resource {
	return &projectVariableResource{}
}

type projectVariableResource struct {
	client *client.Client
}

type projectVariableResourceModel struct {
	ID        types.String `tfsdk:"id"`
	ProjectID types.String `tfsdk:"project_id"`
	Name      types.String `tfsdk:"name"`
	Value     types.String `tfsdk:"value"`
	Shared    types.Bool   `tfsdk:"shared"`
	Status    types.String `tfsdk:"status"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

func (r *projectVariableResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_variable"
}

func (r *projectVariableResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A project environment variable, propagated to functions and (when shared) to frontends.",
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
				Description: "Must start with a letter or underscore; letters, numbers, and underscores only. Forces replacement if changed.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"value": schema.StringAttribute{
				Required:  true,
				Sensitive: true,
			},
			"shared": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether this variable is included in the project's shared frontend-variable list.",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "Latest propagation status (`provisioning`, `active`, or `failed`).",
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

func (r *projectVariableResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *projectVariableResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan projectVariableResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := client.CreateVariableRequest{
		Name:  plan.Name.ValueString(),
		Value: plan.Value.ValueString(),
	}
	if !plan.Shared.IsNull() && !plan.Shared.IsUnknown() {
		v := plan.Shared.ValueBool()
		in.Shared = &v
	}

	v, err := r.client.CreateVariable(ctx, plan.ProjectID.ValueString(), in)
	if err != nil {
		resp.Diagnostics.AddError("Error creating project variable", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, projectVariableModelFromAPI(plan.ProjectID.ValueString(), v))...)
}

func (r *projectVariableResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state projectVariableResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	v, err := r.client.GetVariable(ctx, state.ProjectID.ValueString(), state.Name.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading project variable", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, projectVariableModelFromAPI(state.ProjectID.ValueString(), v))...)
}

func (r *projectVariableResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state projectVariableResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := client.UpdateVariableRequest{}
	if !plan.Value.Equal(state.Value) {
		v := plan.Value.ValueString()
		in.Value = &v
	}
	if !plan.Shared.Equal(state.Shared) {
		v := plan.Shared.ValueBool()
		in.Shared = &v
	}

	v, err := r.client.UpdateVariable(ctx, state.ProjectID.ValueString(), state.Name.ValueString(), in)
	if err != nil {
		resp.Diagnostics.AddError("Error updating project variable", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, projectVariableModelFromAPI(state.ProjectID.ValueString(), v))...)
}

func (r *projectVariableResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state projectVariableResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteVariable(ctx, state.ProjectID.ValueString(), state.Name.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting project variable", err.Error())
	}
}

func (r *projectVariableResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importStateCompositeID(ctx, req, resp, "project_id", "name")
}
