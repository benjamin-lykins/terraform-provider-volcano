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
	_ resource.Resource                = &frontendSharedVariablesResource{}
	_ resource.ResourceWithConfigure   = &frontendSharedVariablesResource{}
	_ resource.ResourceWithImportState = &frontendSharedVariablesResource{}
)

func NewFrontendSharedVariablesResource() resource.Resource {
	return &frontendSharedVariablesResource{}
}

type frontendSharedVariablesResource struct {
	client *client.Client
}

type frontendSharedVariablesResourceModel struct {
	ProjectID     types.String `tfsdk:"project_id"`
	VariableNames types.List   `tfsdk:"variable_names"`
}

func (r *frontendSharedVariablesResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_frontend_shared_variables"
}

func (r *frontendSharedVariablesResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The project's complete list of shared frontend variable names (replaces it wholesale). This is a singleton per project, write-only - the API has no GET for it, so Read cannot detect drift and simply trusts the last value Terraform applied.",
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"variable_names": schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "Project variable names to share with every frontend whose variable_scope is \"all\". An empty list clears the shared set.",
			},
		},
	}
}

func (r *frontendSharedVariablesResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *frontendSharedVariablesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan frontendSharedVariablesResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var names []string
	resp.Diagnostics.Append(plan.VariableNames.ElementsAs(ctx, &names, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.SetFrontendSharedVariables(ctx, plan.ProjectID.ValueString(), names); err != nil {
		resp.Diagnostics.AddError("Error setting frontend shared variables", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *frontendSharedVariablesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state frontendSharedVariablesResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *frontendSharedVariablesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan frontendSharedVariablesResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var names []string
	resp.Diagnostics.Append(plan.VariableNames.ElementsAs(ctx, &names, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.SetFrontendSharedVariables(ctx, plan.ProjectID.ValueString(), names); err != nil {
		resp.Diagnostics.AddError("Error updating frontend shared variables", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *frontendSharedVariablesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state frontendSharedVariablesResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.SetFrontendSharedVariables(ctx, state.ProjectID.ValueString(), nil); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Error clearing frontend shared variables", err.Error())
	}
}

func (r *frontendSharedVariablesResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, pathProjectID, req.ID)...)
}
