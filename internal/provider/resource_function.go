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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/benjamin-lykins/terraform-provider-volcano/internal/client"
)

var (
	_ resource.Resource                = &functionResource{}
	_ resource.ResourceWithConfigure   = &functionResource{}
	_ resource.ResourceWithImportState = &functionResource{}
)

func NewFunctionResource() resource.Resource {
	return &functionResource{}
}

type functionResource struct {
	client *client.Client
}

type functionResourceModel struct {
	ID              types.String `tfsdk:"id"`
	ProjectID       types.String `tfsdk:"project_id"`
	Name            types.String `tfsdk:"name"`
	Runtime         types.String `tfsdk:"runtime"`
	Source          types.String `tfsdk:"source"`
	Handler         types.String `tfsdk:"handler"`
	HTTPAuthMode    types.String `tfsdk:"http_auth_mode"`
	InvocationMode  types.String `tfsdk:"invocation_mode"`
	IsPublic        types.Bool   `tfsdk:"is_public"`
	OpenAPISpec     types.String `tfsdk:"openapi_spec"`
	VariableScope   types.String `tfsdk:"variable_scope"`
	Variables       types.List   `tfsdk:"variables"`
	Status          types.String `tfsdk:"status"`
	InvokeURL       types.String `tfsdk:"invoke_url"`
	DeployedRegions types.List   `tfsdk:"deployed_regions"`
	CreatedAt       types.String `tfsdk:"created_at"`
	UpdatedAt       types.String `tfsdk:"updated_at"`
}

func (r *functionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_function"
}

func (r *functionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A deployed Volcano function. Code, handler, runtime, and variable scoping are set once at creation - the API has no way to redeploy new code in place, so changing any of them replaces the function.",
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
				Description: "DNS-safe name (lowercase letters, numbers, hyphens). Forces replacement if changed.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"runtime": schema.StringAttribute{
				Required:    true,
				Description: "e.g. nodejs24.x, python3.13, ruby3.4. See the volcano_function_runtimes data source for valid values. Forces replacement if changed.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"source": schema.StringAttribute{
				Required:    true,
				Description: "Path to a local zip or tar.gz archive containing the function's source code and dependency manifests. Forces replacement if changed.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"handler": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Exported function name to invoke. Defaults to \"handler\". Forces replacement if changed.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"http_auth_mode": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "volcano (default) or none. none is valid only for public HTTP-mode functions.",
				Validators: []validator.String{
					stringvalidator.OneOf("volcano", "none"),
				},
			},
			"invocation_mode": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "rpc (default, POST {payload}) or http (forwards HTTP request semantics).",
				Validators: []validator.String{
					stringvalidator.OneOf("rpc", "http"),
				},
			},
			"is_public": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether the function is reachable through public invocation ingress.",
			},
			"openapi_spec": schema.StringAttribute{
				Optional:    true,
				Description: "JSON-encoded OpenAPI 3.0/3.1 metadata, for an http-mode function.",
			},
			"variable_scope": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "all (default; shared project variables only) or scoped (only variables listed in `variables`). Forces replacement if changed.",
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
				Description: "Project variable names this function requires, on top of ones detected in its source. Only used when variable_scope is scoped. Forces replacement if changed.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
					listplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				Computed: true,
			},
			"invoke_url": schema.StringAttribute{
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

func (r *functionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *functionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan functionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	content, err := os.ReadFile(plan.Source.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading function source archive", err.Error())
		return
	}

	in := client.CreateFunctionRequest{
		Name:           plan.Name.ValueString(),
		Runtime:        plan.Runtime.ValueString(),
		CodeFilename:   filepath.Base(plan.Source.ValueString()),
		CodeContent:    content,
		Handler:        plan.Handler.ValueString(),
		HTTPAuthMode:   plan.HTTPAuthMode.ValueString(),
		InvocationMode: plan.InvocationMode.ValueString(),
		OpenAPISpec:    plan.OpenAPISpec.ValueString(),
		VariableScope:  plan.VariableScope.ValueString(),
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

	fn, err := r.client.CreateFunction(ctx, plan.ProjectID.ValueString(), in)
	if err != nil {
		resp.Diagnostics.AddError("Error creating function", err.Error())
		return
	}

	model := functionModelFromAPI(ctx, plan.ProjectID.ValueString(), plan.Source, plan.VariableScope, plan.Variables, fn, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *functionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state functionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	fn, err := r.client.GetFunction(ctx, state.ProjectID.ValueString(), state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading function", err.Error())
		return
	}

	model := functionModelFromAPI(ctx, state.ProjectID.ValueString(), state.Source, state.VariableScope, state.Variables, fn, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *functionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state functionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := client.UpdateFunctionRequest{}
	if !plan.HTTPAuthMode.Equal(state.HTTPAuthMode) {
		v := plan.HTTPAuthMode.ValueString()
		in.HTTPAuthMode = &v
	}
	if !plan.InvocationMode.Equal(state.InvocationMode) {
		v := plan.InvocationMode.ValueString()
		in.InvocationMode = &v
	}
	if !plan.IsPublic.Equal(state.IsPublic) {
		v := plan.IsPublic.ValueBool()
		in.IsPublic = &v
	}
	if !plan.OpenAPISpec.Equal(state.OpenAPISpec) {
		v := plan.OpenAPISpec.ValueString()
		in.OpenAPISpec = &v
	}

	fn, err := r.client.UpdateFunction(ctx, state.ProjectID.ValueString(), state.ID.ValueString(), in)
	if err != nil {
		resp.Diagnostics.AddError("Error updating function", err.Error())
		return
	}

	model := functionModelFromAPI(ctx, state.ProjectID.ValueString(), state.Source, state.VariableScope, state.Variables, fn, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *functionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state functionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteFunction(ctx, state.ProjectID.ValueString(), state.ID.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting function", err.Error())
	}
}

func (r *functionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importStateCompositeID(ctx, req, resp, "project_id", "id")
}
