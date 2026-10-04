package provider

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/benjamin-lykins/terraform-provider-volcano/internal/client"
)

var (
	_ resource.Resource                = &durableFunctionResource{}
	_ resource.ResourceWithConfigure   = &durableFunctionResource{}
	_ resource.ResourceWithImportState = &durableFunctionResource{}
)

func NewDurableFunctionResource() resource.Resource {
	return &durableFunctionResource{}
}

type durableFunctionResource struct {
	client *client.Client
}

type durableFunctionResourceModel struct {
	ID              types.String `tfsdk:"id"`
	ProjectID       types.String `tfsdk:"project_id"`
	Name            types.String `tfsdk:"name"`
	Runtime         types.String `tfsdk:"runtime"`
	Source          types.String `tfsdk:"source"`
	Handler         types.String `tfsdk:"handler"`
	IsPublic        types.Bool   `tfsdk:"is_public"`
	VariableScope   types.String `tfsdk:"variable_scope"`
	Variables       types.List   `tfsdk:"variables"`
	Status          types.String `tfsdk:"status"`
	DeployedRegions types.List   `tfsdk:"deployed_regions"`
	CreatedAt       types.String `tfsdk:"created_at"`
	UpdatedAt       types.String `tfsdk:"updated_at"`
}

func (r *durableFunctionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_durable_function"
}

func (r *durableFunctionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A long-running Volcano function that checkpoints its progress and resumes from the last checkpoint. The API has no update endpoint at all for this resource - every attribute change replaces it.",
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
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"runtime": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"source": schema.StringAttribute{
				Required:    true,
				Description: "Path to a local zip or tar.gz archive containing the function's source code and dependency manifests.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"handler": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"is_public": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"variable_scope": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.OneOf("all", "scoped"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"variables": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
					listplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				Computed: true,
			},
			"deployed_regions": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
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

func (r *durableFunctionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *durableFunctionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan durableFunctionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	content, err := os.ReadFile(plan.Source.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading durable function source archive", err.Error())
		return
	}

	in := client.CreateDurableFunctionRequest{
		Name:          plan.Name.ValueString(),
		Runtime:       plan.Runtime.ValueString(),
		CodeFilename:  filepath.Base(plan.Source.ValueString()),
		CodeContent:   content,
		Handler:       plan.Handler.ValueString(),
		VariableScope: plan.VariableScope.ValueString(),
	}
	if !plan.IsPublic.IsNull() && !plan.IsPublic.IsUnknown() {
		v := plan.IsPublic.ValueBool()
		in.IsPublic = &v
	}
	if !plan.Variables.IsNull() && !plan.Variables.IsUnknown() {
		var vars []string
		resp.Diagnostics.Append(plan.Variables.ElementsAs(ctx, &vars, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		in.Variables = vars
	}

	fn, err := r.client.CreateDurableFunction(ctx, plan.ProjectID.ValueString(), in)
	if err != nil {
		resp.Diagnostics.AddError("Error creating durable function", err.Error())
		return
	}

	model := durableFunctionModelFromAPI(ctx, plan.ProjectID.ValueString(), plan.Source, plan.VariableScope, plan.Variables, fn, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *durableFunctionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state durableFunctionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	fn, err := r.client.GetDurableFunction(ctx, state.ProjectID.ValueString(), state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading durable function", err.Error())
		return
	}

	model := durableFunctionModelFromAPI(ctx, state.ProjectID.ValueString(), state.Source, state.VariableScope, state.Variables, fn, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *durableFunctionResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Durable functions cannot be updated in place",
		"Every attribute forces replacement; Update should never be invoked. This is a provider bug if it was.",
	)
}

func (r *durableFunctionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state durableFunctionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteDurableFunction(ctx, state.ProjectID.ValueString(), state.ID.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting durable function", err.Error())
	}
}

func (r *durableFunctionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importStateCompositeID(ctx, req, resp, "project_id", "id")
}
