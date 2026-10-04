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
	_ resource.Resource                = &storageBucketPolicyResource{}
	_ resource.ResourceWithConfigure   = &storageBucketPolicyResource{}
	_ resource.ResourceWithImportState = &storageBucketPolicyResource{}
)

func NewStorageBucketPolicyResource() resource.Resource {
	return &storageBucketPolicyResource{}
}

type storageBucketPolicyResource struct {
	client *client.Client
}

type storageBucketPolicyResourceModel struct {
	ID         types.String `tfsdk:"id"`
	ProjectID  types.String `tfsdk:"project_id"`
	BucketName types.String `tfsdk:"bucket_name"`
	Name       types.String `tfsdk:"name"`
	Operation  types.String `tfsdk:"operation"`
	Definition types.String `tfsdk:"definition"`
	CreatedAt  types.String `tfsdk:"created_at"`
}

func (r *storageBucketPolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_storage_bucket_policy"
}

func (r *storageBucketPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A row-level access policy for a storage bucket, evaluated at request time for one operation. There is no update endpoint - every attribute change replaces it. There is also no single-policy GET; Read reconciles by listing the bucket's policies and matching on id.",
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
			"bucket_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Unique within the bucket.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"operation": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.OneOf("SELECT", "INSERT", "UPDATE", "DELETE"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"definition": schema.StringAttribute{
				Required:    true,
				Description: "Policy expression evaluated at request time, e.g. \"auth.uid() = owner_id\".",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
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

func (r *storageBucketPolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *storageBucketPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan storageBucketPolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policy, err := r.client.CreateStorageBucketPolicy(ctx, plan.ProjectID.ValueString(), plan.BucketName.ValueString(), client.CreateStorageBucketPolicyRequest{
		Name:       plan.Name.ValueString(),
		Operation:  plan.Operation.ValueString(),
		Definition: plan.Definition.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating storage bucket policy", err.Error())
		return
	}

	model := storageBucketPolicyModelFromAPI(plan.ProjectID.ValueString(), plan.BucketName.ValueString(), policy)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *storageBucketPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state storageBucketPolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policies, err := r.client.ListStorageBucketPolicies(ctx, state.ProjectID.ValueString(), state.BucketName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error listing storage bucket policies", err.Error())
		return
	}

	for _, p := range policies {
		if p.ID == state.ID.ValueString() {
			model := storageBucketPolicyModelFromAPI(state.ProjectID.ValueString(), state.BucketName.ValueString(), &p)
			resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
			return
		}
	}
	resp.State.RemoveResource(ctx)
}

func (r *storageBucketPolicyResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Storage bucket policies cannot be updated in place",
		"Every attribute forces replacement; Update should never be invoked. This is a provider bug if it was.",
	)
}

func (r *storageBucketPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state storageBucketPolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteStorageBucketPolicy(ctx, state.ProjectID.ValueString(), state.BucketName.ValueString(), state.ID.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting storage bucket policy", err.Error())
	}
}

func (r *storageBucketPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importStateCompositeID(ctx, req, resp, "project_id", "bucket_name", "id")
}
