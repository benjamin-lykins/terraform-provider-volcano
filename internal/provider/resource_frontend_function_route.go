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
	_ resource.Resource                = &frontendFunctionRouteResource{}
	_ resource.ResourceWithConfigure   = &frontendFunctionRouteResource{}
	_ resource.ResourceWithImportState = &frontendFunctionRouteResource{}
)

func NewFrontendFunctionRouteResource() resource.Resource {
	return &frontendFunctionRouteResource{}
}

type frontendFunctionRouteResource struct {
	client *client.Client
}

type frontendFunctionRouteResourceModel struct {
	ID          types.String `tfsdk:"id"`
	ProjectID   types.String `tfsdk:"project_id"`
	FrontendID  types.String `tfsdk:"frontend_id"`
	FunctionID  types.String `tfsdk:"function_id"`
	PathPrefix  types.String `tfsdk:"path_prefix"`
	StripPrefix types.Bool   `tfsdk:"strip_prefix"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

func (r *frontendFunctionRouteResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_frontend_function_route"
}

func (r *frontendFunctionRouteResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Forwards requests under a path prefix on a frontend to a function. There is no single-route GET; Read reconciles by listing the frontend's routes and matching on id.",
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
			"frontend_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"function_id": schema.StringAttribute{
				Required: true,
			},
			"path_prefix": schema.StringAttribute{
				Required: true,
			},
			"strip_prefix": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
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

func (r *frontendFunctionRouteResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *frontendFunctionRouteResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan frontendFunctionRouteResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	strip := plan.StripPrefix.ValueBool()
	route, err := r.client.CreateFrontendFunctionRoute(ctx, plan.ProjectID.ValueString(), plan.FrontendID.ValueString(), client.FrontendFunctionRouteRequest{
		FunctionID:  plan.FunctionID.ValueString(),
		PathPrefix:  plan.PathPrefix.ValueString(),
		StripPrefix: &strip,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating frontend function route", err.Error())
		return
	}

	model := frontendFunctionRouteModelFromAPI(plan.ProjectID.ValueString(), plan.FrontendID.ValueString(), route)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *frontendFunctionRouteResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state frontendFunctionRouteResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	routes, err := r.client.ListFrontendFunctionRoutes(ctx, state.ProjectID.ValueString(), state.FrontendID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error listing frontend function routes", err.Error())
		return
	}

	for _, route := range routes {
		if route.ID == state.ID.ValueString() {
			model := frontendFunctionRouteModelFromAPI(state.ProjectID.ValueString(), state.FrontendID.ValueString(), &route)
			resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
			return
		}
	}
	resp.State.RemoveResource(ctx)
}

func (r *frontendFunctionRouteResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state frontendFunctionRouteResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	strip := plan.StripPrefix.ValueBool()
	route, err := r.client.UpdateFrontendFunctionRoute(ctx, state.ProjectID.ValueString(), state.FrontendID.ValueString(), state.ID.ValueString(), client.FrontendFunctionRouteRequest{
		FunctionID:  plan.FunctionID.ValueString(),
		PathPrefix:  plan.PathPrefix.ValueString(),
		StripPrefix: &strip,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating frontend function route", err.Error())
		return
	}

	model := frontendFunctionRouteModelFromAPI(state.ProjectID.ValueString(), state.FrontendID.ValueString(), route)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *frontendFunctionRouteResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state frontendFunctionRouteResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteFrontendFunctionRoute(ctx, state.ProjectID.ValueString(), state.FrontendID.ValueString(), state.ID.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting frontend function route", err.Error())
	}
}

func (r *frontendFunctionRouteResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importStateCompositeID(ctx, req, resp, "project_id", "frontend_id", "id")
}
