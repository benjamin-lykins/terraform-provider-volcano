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
	_ resource.Resource                = &sharedVariablesResource{}
	_ resource.ResourceWithConfigure   = &sharedVariablesResource{}
	_ resource.ResourceWithImportState = &sharedVariablesResource{}
)

func NewSharedVariablesResource() resource.Resource {
	return &sharedVariablesResource{}
}

type sharedVariablesResource struct {
	client *client.Client
}

type sharedVariablesResourceModel struct {
	ProjectID     types.String `tfsdk:"project_id"`
	VariableNames types.List   `tfsdk:"variable_names"`
}

func (r *sharedVariablesResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_shared_variables"
}

func (r *sharedVariablesResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The project's complete list of shared function variable names (membership only, not values; replaces it wholesale). Singleton per project, write-only - the API has no GET for it, so Read cannot detect drift and simply trusts the last value Terraform applied. See volcano_frontend_shared_variables for the equivalent list scoped to frontends.",
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
				Description: "Project variable names to share with every function whose variable_scope is \"all\". An empty list clears the shared set.",
			},
		},
	}
}

func (r *sharedVariablesResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *sharedVariablesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan sharedVariablesResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var names []string
	resp.Diagnostics.Append(plan.VariableNames.ElementsAs(ctx, &names, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.SetSharedVariables(ctx, plan.ProjectID.ValueString(), names); err != nil {
		resp.Diagnostics.AddError("Error setting shared variables", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *sharedVariablesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state sharedVariablesResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *sharedVariablesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan sharedVariablesResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var names []string
	resp.Diagnostics.Append(plan.VariableNames.ElementsAs(ctx, &names, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.SetSharedVariables(ctx, plan.ProjectID.ValueString(), names); err != nil {
		resp.Diagnostics.AddError("Error updating shared variables", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *sharedVariablesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state sharedVariablesResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.SetSharedVariables(ctx, state.ProjectID.ValueString(), nil); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Error clearing shared variables", err.Error())
	}
}

func (r *sharedVariablesResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, pathProjectID, req.ID)...)
}
