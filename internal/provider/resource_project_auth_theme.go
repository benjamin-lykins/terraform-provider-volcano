package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/benjamin-lykins/terraform-provider-volcano/internal/client"
)

var (
	_ resource.Resource                = &projectAuthThemeResource{}
	_ resource.ResourceWithConfigure   = &projectAuthThemeResource{}
	_ resource.ResourceWithImportState = &projectAuthThemeResource{}
)

func NewProjectAuthThemeResource() resource.Resource {
	return &projectAuthThemeResource{}
}

type projectAuthThemeResource struct {
	client *client.Client
}

type themeColorsModel struct {
	Background types.String `tfsdk:"background"`
	Surface    types.String `tfsdk:"surface"`
	Text       types.String `tfsdk:"text"`
	Accent     types.String `tfsdk:"accent"`
	AccentText types.String `tfsdk:"accent_text"`
}

type projectAuthThemeResourceModel struct {
	ProjectID types.String     `tfsdk:"project_id"`
	Colors    themeColorsModel `tfsdk:"colors"`
	Font      types.String     `tfsdk:"font"`
	Scale     types.String     `tfsdk:"scale"`
	Density   types.String     `tfsdk:"density"`
	Radius    types.String     `tfsdk:"radius"`
}

func (r *projectAuthThemeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_auth_theme"
}

func (r *projectAuthThemeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	colorAttr := schema.StringAttribute{
		Required:   true,
		Validators: []validator.String{stringvalidator.RegexMatches(hexColorRegexp, "must be a 6-digit hex color, e.g. #112233")},
	}
	resp.Schema = schema.Schema{
		Description: "The visual theme for a project's hosted auth pages. Singleton - read back via the appearance endpoint since theme has no GET of its own.",
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"colors": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"background":  colorAttr,
					"surface":     colorAttr,
					"text":        colorAttr,
					"accent":      colorAttr,
					"accent_text": colorAttr,
				},
			},
			"font": schema.StringAttribute{
				Optional:   true,
				Computed:   true,
				Validators: []validator.String{stringvalidator.OneOf("system", "humanist", "geometric", "slab", "mono")},
			},
			"scale": schema.StringAttribute{
				Optional:   true,
				Computed:   true,
				Validators: []validator.String{stringvalidator.OneOf("small", "default", "large")},
			},
			"density": schema.StringAttribute{
				Optional:   true,
				Computed:   true,
				Validators: []validator.String{stringvalidator.OneOf("compact", "comfortable", "spacious")},
			},
			"radius": schema.StringAttribute{
				Optional:   true,
				Computed:   true,
				Validators: []validator.String{stringvalidator.OneOf("none", "small", "medium", "large")},
			},
		},
	}
}

func (r *projectAuthThemeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *projectAuthThemeResource) apply(ctx context.Context, plan projectAuthThemeResourceModel, diags *diag.Diagnostics) (*client.Theme, bool) {
	theme := client.Theme{
		Colors: client.ThemeColors{
			Background: plan.Colors.Background.ValueString(),
			Surface:    plan.Colors.Surface.ValueString(),
			Text:       plan.Colors.Text.ValueString(),
			Accent:     plan.Colors.Accent.ValueString(),
			AccentText: plan.Colors.AccentText.ValueString(),
		},
		Font:    plan.Font.ValueString(),
		Scale:   plan.Scale.ValueString(),
		Density: plan.Density.ValueString(),
		Radius:  plan.Radius.ValueString(),
	}
	out, err := r.client.SetProjectAuthTheme(ctx, plan.ProjectID.ValueString(), theme)
	if err != nil {
		diags.AddError("Error setting project auth theme", err.Error())
		return nil, false
	}
	return out, true
}

func (r *projectAuthThemeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan projectAuthThemeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	theme, ok := r.apply(ctx, plan, &resp.Diagnostics)
	if !ok {
		return
	}

	model := projectAuthThemeModelFromAPI(plan.ProjectID.ValueString(), theme)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *projectAuthThemeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state projectAuthThemeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	appearance, err := r.client.GetProjectAuthPagesAppearance(ctx, state.ProjectID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading project auth theme", err.Error())
		return
	}

	model := projectAuthThemeModelFromAPI(state.ProjectID.ValueString(), &appearance.Theme)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *projectAuthThemeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan projectAuthThemeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	theme, ok := r.apply(ctx, plan, &resp.Diagnostics)
	if !ok {
		return
	}

	model := projectAuthThemeModelFromAPI(plan.ProjectID.ValueString(), theme)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *projectAuthThemeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state projectAuthThemeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteProjectAuthTheme(ctx, state.ProjectID.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting project auth theme", err.Error())
	}
}

func (r *projectAuthThemeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, pathProjectID, req.ID)...)
}
