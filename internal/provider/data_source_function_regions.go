package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/benjamin-lykins/terraform-provider-volcano/internal/client"
)

var _ datasource.DataSource = &functionRegionsDataSource{}
var _ datasource.DataSourceWithConfigure = &functionRegionsDataSource{}

func NewFunctionRegionsDataSource() datasource.DataSource {
	return &functionRegionsDataSource{}
}

type functionRegionsDataSource struct {
	client *client.Client
}

type functionRegionModel struct {
	Code  types.String `tfsdk:"code"`
	Label types.String `tfsdk:"label"`
	Flag  types.String `tfsdk:"flag"`
}

type functionRegionsDataSourceModel struct {
	Regions []functionRegionModel `tfsdk:"regions"`
}

func (d *functionRegionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_function_regions"
}

func (d *functionRegionsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Regions that Volcano functions can deploy to.",
		Attributes: map[string]schema.Attribute{
			"regions": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"code":  schema.StringAttribute{Computed: true, Description: "Region identifier accepted by function APIs."},
						"label": schema.StringAttribute{Computed: true},
						"flag":  schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *functionRegionsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *functionRegionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	regions, err := d.client.ListFunctionRegions(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error listing function regions", err.Error())
		return
	}

	var data functionRegionsDataSourceModel
	for _, r := range regions {
		data.Regions = append(data.Regions, functionRegionModel{
			Code:  types.StringValue(r.Code),
			Label: types.StringValue(r.Label),
			Flag:  stringOrNull(r.Flag),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
