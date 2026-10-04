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
	_ resource.Resource                = &frontendDomainResource{}
	_ resource.ResourceWithConfigure   = &frontendDomainResource{}
	_ resource.ResourceWithImportState = &frontendDomainResource{}
)

func NewFrontendDomainResource() resource.Resource {
	return &frontendDomainResource{}
}

type frontendDomainResource struct {
	client *client.Client
}

type frontendDomainResourceModel struct {
	ProjectID             types.String `tfsdk:"project_id"`
	FrontendID            types.String `tfsdk:"frontend_id"`
	Domain                types.String `tfsdk:"domain"`
	CertificatePEM        types.String `tfsdk:"certificate_pem"`
	PrivateKeyPEM         types.String `tfsdk:"private_key_pem"`
	CertificateChainPEM   types.String `tfsdk:"certificate_chain_pem"`
	DomainStatus          types.String `tfsdk:"domain_status"`
	VerificationStatus    types.String `tfsdk:"verification_status"`
	RoutingTargetHostname types.String `tfsdk:"routing_target_hostname"`
	EffectiveURLs         types.List   `tfsdk:"effective_urls"`
	CreatedAt             types.String `tfsdk:"created_at"`
	UpdatedAt             types.String `tfsdk:"updated_at"`
}

func (r *frontendDomainResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_frontend_domain"
}

func (r *frontendDomainResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A custom domain attached to a frontend, with a bring-your-own-certificate TLS configuration. There is no update endpoint - any attribute change replaces it.",
		Attributes: map[string]schema.Attribute{
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
			"domain": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"certificate_pem": schema.StringAttribute{
				Required:  true,
				Sensitive: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"private_key_pem": schema.StringAttribute{
				Required:  true,
				Sensitive: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"certificate_chain_pem": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"domain_status": schema.StringAttribute{
				Computed: true,
				Validators: []validator.String{
					stringvalidator.OneOf("pending_verification", "provisioning", "active", "detaching", "failed", "deleted"),
				},
			},
			"verification_status": schema.StringAttribute{
				Computed: true,
			},
			"routing_target_hostname": schema.StringAttribute{
				Computed:    true,
				Description: "DNS routing target hostname to point the domain at.",
			},
			"effective_urls": schema.ListAttribute{
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

func (r *frontendDomainResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *frontendDomainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan frontendDomainResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	d, err := r.client.CreateFrontendDomain(ctx, plan.ProjectID.ValueString(), plan.FrontendID.ValueString(), client.CreateFrontendDomainRequest{
		Domain: plan.Domain.ValueString(),
		TLS: client.FrontendDomainTLS{
			Mode:                "byoc",
			CertificatePEM:      plan.CertificatePEM.ValueString(),
			PrivateKeyPEM:       plan.PrivateKeyPEM.ValueString(),
			CertificateChainPEM: plan.CertificateChainPEM.ValueString(),
		},
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating frontend domain", err.Error())
		return
	}

	model := frontendDomainModelFromAPI(ctx, plan, d, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *frontendDomainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state frontendDomainResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	d, err := r.client.GetFrontendDomain(ctx, state.ProjectID.ValueString(), state.FrontendID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading frontend domain", err.Error())
		return
	}
	if d == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	model := frontendDomainModelFromAPI(ctx, state, d, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *frontendDomainResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Frontend domains cannot be updated in place",
		"Every attribute forces replacement; Update should never be invoked. This is a provider bug if it was.",
	)
}

func (r *frontendDomainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state frontendDomainResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteFrontendDomain(ctx, state.ProjectID.ValueString(), state.FrontendID.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting frontend domain", err.Error())
	}
}

func (r *frontendDomainResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importStateCompositeID(ctx, req, resp, "project_id", "frontend_id")
}
