package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/benjamin-lykins/terraform-provider-volcano/internal/client"
)

var _ datasource.DataSource = &postgresVersionsDataSource{}
var _ datasource.DataSourceWithConfigure = &postgresVersionsDataSource{}

func NewPostgresVersionsDataSource() datasource.DataSource {
	return &postgresVersionsDataSource{}
}

type postgresVersionsDataSource struct {
	client *client.Client
}

type postgresVersionModel struct {
	Version    types.String `tfsdk:"version"`
	Name       types.String `tfsdk:"name"`
	Default    types.Bool   `tfsdk:"default"`
	Deprecated types.Bool   `tfsdk:"deprecated"`
}

type postgresVersionsDataSourceModel struct {
	Versions []postgresVersionModel `tfsdk:"versions"`
}

func (d *postgresVersionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_database_postgres_versions"
}

func (d *postgresVersionsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "PostgreSQL major versions available for new Volcano databases.",
		Attributes: map[string]schema.Attribute{
			"versions": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"version":    schema.StringAttribute{Computed: true},
						"name":       schema.StringAttribute{Computed: true},
						"default":    schema.BoolAttribute{Computed: true},
						"deprecated": schema.BoolAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *postgresVersionsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *postgresVersionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	versions, err := d.client.ListPostgresVersions(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error listing postgres versions", err.Error())
		return
	}

	var data postgresVersionsDataSourceModel
	for _, v := range versions {
		data.Versions = append(data.Versions, postgresVersionModel{
			Version:    types.StringValue(v.Version),
			Name:       types.StringValue(v.Name),
			Default:    types.BoolValue(v.Default),
			Deprecated: types.BoolValue(v.Deprecated),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
