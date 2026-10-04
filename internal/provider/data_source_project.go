package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/benjamin-lykins/terraform-provider-volcano/internal/client"
)

var _ datasource.DataSource = &projectDataSource{}
var _ datasource.DataSourceWithConfigure = &projectDataSource{}

func NewProjectDataSource() datasource.DataSource {
	return &projectDataSource{}
}

type projectDataSource struct {
	client *client.Client
}

type projectDataSourceModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Status          types.String `tfsdk:"status"`
	Plan            types.String `tfsdk:"plan"`
	AllRegions      types.Bool   `tfsdk:"all_regions"`
	SelectedRegions types.List   `tfsdk:"selected_regions"`
	LogoURL         types.String `tfsdk:"logo_url"`
	CreatedAt       types.String `tfsdk:"created_at"`
	UpdatedAt       types.String `tfsdk:"updated_at"`
}

func (d *projectDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (d *projectDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Look up a single Volcano project by ID.",
		Attributes: map[string]schema.Attribute{
			"id":               schema.StringAttribute{Required: true, Description: "Project UUID."},
			"name":             schema.StringAttribute{Computed: true},
			"status":           schema.StringAttribute{Computed: true},
			"plan":             schema.StringAttribute{Computed: true},
			"all_regions":      schema.BoolAttribute{Computed: true},
			"selected_regions": schema.ListAttribute{Computed: true, ElementType: types.StringType},
			"logo_url":         schema.StringAttribute{Computed: true},
			"created_at":       schema.StringAttribute{Computed: true},
			"updated_at":       schema.StringAttribute{Computed: true},
		},
	}
}

func (d *projectDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type", fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData))
		return
	}
	d.client = c
}

func (d *projectDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data projectDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	p, err := d.client.GetProject(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading project", err.Error())
		return
	}

	data.Name = types.StringValue(p.Name)
	data.Status = types.StringValue(p.Status)
	data.Plan = stringOrNull(p.Plan)
	data.AllRegions = types.BoolValue(p.AllRegions)
	data.SelectedRegions = stringList(ctx, p.SelectedRegions, &resp.Diagnostics)
	data.LogoURL = stringOrNull(p.LogoURL)
	data.CreatedAt = types.StringValue(p.CreatedAt)
	data.UpdatedAt = types.StringValue(p.UpdatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
