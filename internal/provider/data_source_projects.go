package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/benjamin-lykins/terraform-provider-volcano/internal/client"
)

var _ datasource.DataSource = &projectsDataSource{}
var _ datasource.DataSourceWithConfigure = &projectsDataSource{}

func NewProjectsDataSource() datasource.DataSource {
	return &projectsDataSource{}
}

type projectsDataSource struct {
	client *client.Client
}

type projectsDataSourceModel struct {
	Search   types.String             `tfsdk:"search"`
	Projects []projectDataSourceModel `tfsdk:"projects"`
}

func (d *projectsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_projects"
}

func (d *projectsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "List Volcano projects on the account, optionally filtered by search term.",
		Attributes: map[string]schema.Attribute{
			"search": schema.StringAttribute{
				Optional:    true,
				Description: "Optional free-text filter on project name.",
			},
			"projects": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":               schema.StringAttribute{Computed: true},
						"name":             schema.StringAttribute{Computed: true},
						"status":           schema.StringAttribute{Computed: true},
						"plan":             schema.StringAttribute{Computed: true},
						"all_regions":      schema.BoolAttribute{Computed: true},
						"selected_regions": schema.ListAttribute{Computed: true, ElementType: types.StringType},
						"logo_url":         schema.StringAttribute{Computed: true},
						"created_at":       schema.StringAttribute{Computed: true},
						"updated_at":       schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *projectsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *projectsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data projectsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projects, err := d.client.ListProjects(ctx, client.ListProjectsOptions{Search: data.Search.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError("Error listing projects", err.Error())
		return
	}

	data.Projects = make([]projectDataSourceModel, 0, len(projects))
	for _, p := range projects {
		data.Projects = append(data.Projects, projectDataSourceModel{
			ID:              types.StringValue(p.ID),
			Name:            types.StringValue(p.Name),
			Status:          types.StringValue(p.Status),
			Plan:            stringOrNull(p.Plan),
			AllRegions:      types.BoolValue(p.AllRegions),
			SelectedRegions: stringList(ctx, p.SelectedRegions, &resp.Diagnostics),
			LogoURL:         stringOrNull(p.LogoURL),
			CreatedAt:       types.StringValue(p.CreatedAt),
			UpdatedAt:       types.StringValue(p.UpdatedAt),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
