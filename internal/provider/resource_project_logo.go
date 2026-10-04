package provider

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/benjamin-lykins/terraform-provider-volcano/internal/client"
)

var (
	_ resource.Resource                = &projectLogoResource{}
	_ resource.ResourceWithConfigure   = &projectLogoResource{}
	_ resource.ResourceWithImportState = &projectLogoResource{}
)

func NewProjectLogoResource() resource.Resource {
	return &projectLogoResource{}
}

type projectLogoResource struct {
	client *client.Client
}

type projectLogoResourceModel struct {
	ProjectID types.String `tfsdk:"project_id"`
	Source    types.String `tfsdk:"source"`
	LogoURL   types.String `tfsdk:"logo_url"`
}

func (r *projectLogoResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_logo"
}

func (r *projectLogoResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A project's logo image (PNG, JPEG, GIF, WebP, or SVG; max 2 MB). The API has no way to compare existing logo bytes, so drift detection is limited to whether a logo is present at all - if it is removed out of band, this resource is recreated on the next apply.",
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"source": schema.StringAttribute{
				Required:    true,
				Description: "Path to the local image file to upload.",
			},
			"logo_url": schema.StringAttribute{
				Computed:    true,
				Description: "Relative API path serving the uploaded logo.",
			},
		},
	}
}

func (r *projectLogoResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *projectLogoResource) upload(ctx context.Context, projectID, source string) (*client.Project, error) {
	content, err := os.ReadFile(source)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", source, err)
	}
	return r.client.UploadProjectLogo(ctx, projectID, filepath.Base(source), content)
}

func (r *projectLogoResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan projectLogoResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	project, err := r.upload(ctx, plan.ProjectID.ValueString(), plan.Source.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error uploading project logo", err.Error())
		return
	}

	plan.LogoURL = stringOrNull(project.LogoURL)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *projectLogoResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state projectLogoResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	project, err := r.client.GetProject(ctx, state.ProjectID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading project", err.Error())
		return
	}

	if project.LogoURL == "" {
		// Logo was removed out of band; nothing to reconcile onto but a
		// fresh upload, so drop state and let the next apply recreate it.
		resp.State.RemoveResource(ctx)
		return
	}

	state.LogoURL = types.StringValue(project.LogoURL)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *projectLogoResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan projectLogoResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	project, err := r.upload(ctx, plan.ProjectID.ValueString(), plan.Source.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error uploading project logo", err.Error())
		return
	}

	plan.LogoURL = stringOrNull(project.LogoURL)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *projectLogoResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state projectLogoResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteProjectLogo(ctx, state.ProjectID.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting project logo", err.Error())
	}
}

func (r *projectLogoResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, pathProjectID, req.ID)...)
}
