package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/benjamin-lykins/terraform-provider-volcano/internal/client"
)

var (
	_ resource.Resource                = &projectAuthMethodsResource{}
	_ resource.ResourceWithConfigure   = &projectAuthMethodsResource{}
	_ resource.ResourceWithImportState = &projectAuthMethodsResource{}
)

func NewProjectAuthMethodsResource() resource.Resource {
	return &projectAuthMethodsResource{}
}

type projectAuthMethodsResource struct {
	client *client.Client
}

type oauthProviderToggleModel struct {
	Provider types.String `tfsdk:"provider"`
	Enabled  types.Bool   `tfsdk:"enabled"`
}

type projectAuthMethodsResourceModel struct {
	ProjectID           types.String               `tfsdk:"project_id"`
	EnableAnonymous     types.Bool                 `tfsdk:"enable_anonymous"`
	EnableEmailPassword types.Bool                 `tfsdk:"enable_email_password"`
	OAuthProviders      []oauthProviderToggleModel `tfsdk:"oauth_providers"`
	AvailableMethods    types.List                 `tfsdk:"available_methods"`
}

func (r *projectAuthMethodsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_auth_methods"
}

func (r *projectAuthMethodsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Which sign-in methods are enabled for a project. Singleton - no create/delete endpoint; Create performs the first PUT and Delete is a local no-op.",
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"enable_anonymous": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"enable_email_password": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"oauth_providers": schema.ListNestedAttribute{
				Optional: true,
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"provider": schema.StringAttribute{Required: true, Description: "e.g. google, github, microsoft, apple."},
						"enabled":  schema.BoolAttribute{Required: true},
					},
				},
			},
			"available_methods": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Every sign-in method the project could enable.",
			},
		},
	}
}

func (r *projectAuthMethodsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *projectAuthMethodsResource) apply(ctx context.Context, model projectAuthMethodsResourceModel, diags *diag.Diagnostics) (*client.ProjectAuthMethods, bool) {
	in := client.UpdateProjectAuthMethodsRequest{
		EnableAnonymous:     boolPtrFromModel(model.EnableAnonymous),
		EnableEmailPassword: boolPtrFromModel(model.EnableEmailPassword),
	}
	for _, p := range model.OAuthProviders {
		in.OAuthProviders = append(in.OAuthProviders, client.OAuthProviderToggle{
			Provider: p.Provider.ValueString(),
			Enabled:  p.Enabled.ValueBool(),
		})
	}

	out, err := r.client.UpdateProjectAuthMethods(ctx, model.ProjectID.ValueString(), in)
	if err != nil {
		diags.AddError("Error setting project auth methods", err.Error())
		return nil, false
	}
	return out, true
}

func (r *projectAuthMethodsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan projectAuthMethodsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, ok := r.apply(ctx, plan, &resp.Diagnostics)
	if !ok {
		return
	}

	model := projectAuthMethodsModelFromAPI(ctx, plan.ProjectID.ValueString(), out, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *projectAuthMethodsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state projectAuthMethodsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := r.client.GetProjectAuthMethods(ctx, state.ProjectID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading project auth methods", err.Error())
		return
	}

	model := projectAuthMethodsModelFromAPI(ctx, state.ProjectID.ValueString(), out, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *projectAuthMethodsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan projectAuthMethodsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, ok := r.apply(ctx, plan, &resp.Diagnostics)
	if !ok {
		return
	}

	model := projectAuthMethodsModelFromAPI(ctx, plan.ProjectID.ValueString(), out, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *projectAuthMethodsResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// No-op: no delete endpoint exists for this singleton.
}

func (r *projectAuthMethodsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, pathProjectID, req.ID)...)
}
