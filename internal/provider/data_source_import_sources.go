package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/benjamin-lykins/terraform-provider-volcano/internal/client"
)

var _ datasource.DataSource = &importSourcesDataSource{}
var _ datasource.DataSourceWithConfigure = &importSourcesDataSource{}

func NewImportSourcesDataSource() datasource.DataSource {
	return &importSourcesDataSource{}
}

type importSourcesDataSource struct {
	client *client.Client
}

type importSourceModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Framework types.String `tfsdk:"framework"`
	AccountID types.String `tfsdk:"account_id"`
}

type importSourcesDataSourceModel struct {
	Provider types.String        `tfsdk:"import_provider"`
	Sources  []importSourceModel `tfsdk:"sources"`
}

func (d *importSourcesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_import_sources"
}

func (d *importSourcesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Projects available to import from a connected third-party platform (see volcano_user_import_connections).",
		Attributes: map[string]schema.Attribute{
			"import_provider": schema.StringAttribute{Required: true, Description: "The import source provider, e.g. a hosting platform identifier."},
			"sources": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":         schema.StringAttribute{Computed: true},
						"name":       schema.StringAttribute{Computed: true},
						"framework":  schema.StringAttribute{Computed: true},
						"account_id": schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *importSourcesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *importSourcesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data importSourcesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sources, err := d.client.ListImportSources(ctx, data.Provider.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error listing import sources", err.Error())
		return
	}

	data.Sources = make([]importSourceModel, 0, len(sources))
	for _, s := range sources {
		data.Sources = append(data.Sources, importSourceModel{
			ID:        types.StringValue(s.ID),
			Name:      types.StringValue(s.Name),
			Framework: stringOrNull(s.Framework),
			AccountID: stringOrNull(s.AccountID),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
