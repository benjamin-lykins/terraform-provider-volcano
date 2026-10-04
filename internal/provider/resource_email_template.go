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
	_ resource.Resource                = &emailTemplateResource{}
	_ resource.ResourceWithConfigure   = &emailTemplateResource{}
	_ resource.ResourceWithImportState = &emailTemplateResource{}
)

func NewEmailTemplateResource() resource.Resource {
	return &emailTemplateResource{}
}

type emailTemplateResource struct {
	client *client.Client
}

type emailTemplateResourceModel struct {
	ID           types.String `tfsdk:"id"`
	ProjectID    types.String `tfsdk:"project_id"`
	TemplateType types.String `tfsdk:"template_type"`
	Subject      types.String `tfsdk:"subject"`
	HTMLBody     types.String `tfsdk:"html_body"`
	TextBody     types.String `tfsdk:"text_body"`
	CreatedAt    types.String `tfsdk:"created_at"`
	UpdatedAt    types.String `tfsdk:"updated_at"`
}

func (r *emailTemplateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_email_template"
}

func (r *emailTemplateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A project's override of an auth email's subject and body. See the volcano_email_template_defaults data source for the platform defaults this replaces.",
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
			"template_type": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.OneOf("welcome", "confirmation", "password_reset", "password_changed"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"subject": schema.StringAttribute{
				Required: true,
			},
			"html_body": schema.StringAttribute{
				Required: true,
			},
			"text_body": schema.StringAttribute{
				Required: true,
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

func (r *emailTemplateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *emailTemplateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan emailTemplateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tmpl, err := r.client.CreateEmailTemplate(ctx, plan.ProjectID.ValueString(), client.CreateEmailTemplateRequest{
		TemplateType: plan.TemplateType.ValueString(),
		Subject:      plan.Subject.ValueString(),
		HTMLBody:     plan.HTMLBody.ValueString(),
		TextBody:     plan.TextBody.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating email template", err.Error())
		return
	}

	model := emailTemplateModelFromAPI(plan.ProjectID.ValueString(), tmpl)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *emailTemplateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state emailTemplateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tmpl, err := r.client.GetEmailTemplate(ctx, state.ProjectID.ValueString(), state.TemplateType.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading email template", err.Error())
		return
	}

	model := emailTemplateModelFromAPI(state.ProjectID.ValueString(), tmpl)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *emailTemplateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state emailTemplateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := client.UpdateEmailTemplateRequest{}
	if !plan.Subject.Equal(state.Subject) {
		v := plan.Subject.ValueString()
		in.Subject = &v
	}
	if !plan.HTMLBody.Equal(state.HTMLBody) {
		v := plan.HTMLBody.ValueString()
		in.HTMLBody = &v
	}
	if !plan.TextBody.Equal(state.TextBody) {
		v := plan.TextBody.ValueString()
		in.TextBody = &v
	}

	tmpl, err := r.client.UpdateEmailTemplate(ctx, state.ProjectID.ValueString(), state.TemplateType.ValueString(), in)
	if err != nil {
		resp.Diagnostics.AddError("Error updating email template", err.Error())
		return
	}

	model := emailTemplateModelFromAPI(state.ProjectID.ValueString(), tmpl)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *emailTemplateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state emailTemplateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteEmailTemplate(ctx, state.ProjectID.ValueString(), state.TemplateType.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting email template", err.Error())
	}
}

func (r *emailTemplateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importStateCompositeID(ctx, req, resp, "project_id", "template_type")
}
