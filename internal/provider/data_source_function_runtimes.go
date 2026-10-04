package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/benjamin-lykins/terraform-provider-volcano/internal/client"
)

var _ datasource.DataSource = &functionRuntimesDataSource{}
var _ datasource.DataSourceWithConfigure = &functionRuntimesDataSource{}

func NewFunctionRuntimesDataSource() datasource.DataSource {
	return &functionRuntimesDataSource{}
}

type functionRuntimesDataSource struct {
	client *client.Client
}

type functionRuntimeModel struct {
	Name           types.String `tfsdk:"name"`
	Label          types.String `tfsdk:"label"`
	Language       types.String `tfsdk:"language"`
	Default        types.Bool   `tfsdk:"default"`
	DurableCapable types.Bool   `tfsdk:"durable_capable"`
}

type functionRuntimesDataSourceModel struct {
	Runtimes []functionRuntimeModel `tfsdk:"runtimes"`
}

func (d *functionRuntimesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_function_runtimes"
}

func (d *functionRuntimesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Language runtimes available to functions and durable functions.",
		Attributes: map[string]schema.Attribute{
			"runtimes": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name":            schema.StringAttribute{Computed: true, Description: "Value to use as a function's `runtime`, e.g. nodejs24.x."},
						"label":           schema.StringAttribute{Computed: true},
						"language":        schema.StringAttribute{Computed: true},
						"default":         schema.BoolAttribute{Computed: true, Description: "Whether this is the CLI default for its language."},
						"durable_capable": schema.BoolAttribute{Computed: true, Description: "Whether this runtime can back a durable function."},
					},
				},
			},
		},
	}
}

func (d *functionRuntimesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *functionRuntimesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	runtimes, err := d.client.ListFunctionRuntimes(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error listing function runtimes", err.Error())
		return
	}

	var data functionRuntimesDataSourceModel
	for _, r := range runtimes {
		data.Runtimes = append(data.Runtimes, functionRuntimeModel{
			Name:           types.StringValue(r.Name),
			Label:          stringOrNull(r.Label),
			Language:       stringOrNull(r.Language),
			Default:        types.BoolValue(r.Default),
			DurableCapable: types.BoolValue(r.DurableCapable),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
