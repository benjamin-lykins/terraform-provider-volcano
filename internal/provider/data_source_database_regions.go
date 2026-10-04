package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/benjamin-lykins/terraform-provider-volcano/internal/client"
)

var _ datasource.DataSource = &databaseRegionsDataSource{}
var _ datasource.DataSourceWithConfigure = &databaseRegionsDataSource{}

func NewDatabaseRegionsDataSource() datasource.DataSource {
	return &databaseRegionsDataSource{}
}

type databaseRegionsDataSource struct {
	client *client.Client
}

type regionModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type databaseRegionsDataSourceModel struct {
	Regions []regionModel `tfsdk:"regions"`
}

func (d *databaseRegionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_database_regions"
}

func (d *databaseRegionsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Hosting regions available for Volcano databases in this environment.",
		Attributes: map[string]schema.Attribute{
			"regions": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":   schema.StringAttribute{Computed: true, Description: "Region identifier for API usage, e.g. us-east-1."},
						"name": schema.StringAttribute{Computed: true, Description: "Human-readable region location."},
					},
				},
			},
		},
	}
}

func (d *databaseRegionsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *databaseRegionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	regions, err := d.client.ListDatabaseRegions(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error listing database regions", err.Error())
		return
	}

	var data databaseRegionsDataSourceModel
	for _, r := range regions {
		data.Regions = append(data.Regions, regionModel{ID: types.StringValue(r.ID), Name: types.StringValue(r.Name)})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
