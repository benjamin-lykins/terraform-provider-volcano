package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/benjamin-lykins/terraform-provider-volcano/internal/client"
)

var (
	_ resource.Resource                = &anonKeyResource{}
	_ resource.ResourceWithConfigure   = &anonKeyResource{}
	_ resource.ResourceWithImportState = &anonKeyResource{}
)

func NewAnonKeyResource() resource.Resource {
	return &anonKeyResource{}
}

type anonKeyResource struct {
	client *client.Client
}

type anonKeyResourceModel struct {
	ID          types.String `tfsdk:"id"`
	ProjectID   types.String `tfsdk:"project_id"`
	Name        types.String `tfsdk:"name"`
	Permissions types.List   `tfsdk:"permissions"`
	IsDefault   types.Bool   `tfsdk:"is_default"`
	KeyValue    types.String `tfsdk:"key_value"`
	CreatedAt   types.String `tfsdk:"created_at"`
}

func (r *anonKeyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_anon_key"
}

func (r *anonKeyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A project anon key: a public JWT safe to embed in frontend code, with a per-key permission set.",
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
				Description: "Letters, numbers, underscores, and hyphens only. Forces replacement if changed.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"permissions": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Permissions granted to this key. Defaults to auth-only permissions when omitted. Forces replacement if changed.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
					listplanmodifier.RequiresReplace(),
				},
			},
			"is_default": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether this is the project's default anon key. Only one key per project may be default; setting this to true demotes the previous default.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"key_value": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "The anon key JWT for use in an Authorization header.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
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

func (r *anonKeyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *anonKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan anonKeyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := client.CreateAnonKeyRequest{Name: plan.Name.ValueString()}
	if !plan.Permissions.IsNull() && !plan.Permissions.IsUnknown() {
		var perms []string
		resp.Diagnostics.Append(plan.Permissions.ElementsAs(ctx, &perms, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		in.Permissions = perms
	}

	key, err := r.client.CreateAnonKey(ctx, plan.ProjectID.ValueString(), in)
	if err != nil {
		resp.Diagnostics.AddError("Error creating anon key", err.Error())
		return
	}

	if plan.IsDefault.ValueBool() && !key.IsDefault {
		if err := r.client.SetDefaultAnonKey(ctx, plan.ProjectID.ValueString(), key.ID); err != nil {
			resp.Diagnostics.AddError("Error setting anon key as default", err.Error())
			return
		}
		key.IsDefault = true
	}

	model := anonKeyModelFromAPI(ctx, plan.ProjectID.ValueString(), key, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *anonKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state anonKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	key, err := r.client.GetAnonKey(ctx, state.ProjectID.ValueString(), state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading anon key", err.Error())
		return
	}

	model := anonKeyModelFromAPI(ctx, state.ProjectID.ValueString(), key, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *anonKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state anonKeyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// The only in-place mutation the API supports is promoting a key to default;
	// demotion happens implicitly when another key is promoted.
	if plan.IsDefault.ValueBool() && !state.IsDefault.ValueBool() {
		if err := r.client.SetDefaultAnonKey(ctx, state.ProjectID.ValueString(), state.ID.ValueString()); err != nil {
			resp.Diagnostics.AddError("Error setting anon key as default", err.Error())
			return
		}
	}

	key, err := r.client.GetAnonKey(ctx, state.ProjectID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading anon key after update", err.Error())
		return
	}

	model := anonKeyModelFromAPI(ctx, state.ProjectID.ValueString(), key, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *anonKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state anonKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteAnonKey(ctx, state.ProjectID.ValueString(), state.ID.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting anon key", err.Error())
	}
}

func (r *anonKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importStateCompositeID(ctx, req, resp, "project_id", "id")
}
