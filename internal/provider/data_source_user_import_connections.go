package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/benjamin-lykins/terraform-provider-volcano/internal/client"
)

var _ datasource.DataSource = &userImportConnectionsDataSource{}
var _ datasource.DataSourceWithConfigure = &userImportConnectionsDataSource{}

func NewUserImportConnectionsDataSource() datasource.DataSource {
	return &userImportConnectionsDataSource{}
}

type userImportConnectionsDataSource struct {
	client *client.Client
}

type userImportConnectionModel struct {
	ID                  types.String `tfsdk:"id"`
	Provider            types.String `tfsdk:"provider"`
	AccountID           types.String `tfsdk:"account_id"`
	AccountName         types.String `tfsdk:"account_name"`
	Status              types.String `tfsdk:"status"`
	LastAuthenticatedAt types.String `tfsdk:"last_authenticated_at"`
}

type userImportConnectionsDataSourceModel struct {
	Connections []userImportConnectionModel `tfsdk:"connections"`
}

func (d *userImportConnectionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_import_connections"
}

func (d *userImportConnectionsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The account's connections to third-party platforms for importing projects from them. Connecting a new one requires an interactive OAuth flow this provider cannot drive - manage that through the dashboard or CLI.",
		Attributes: map[string]schema.Attribute{
			"connections": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":                    schema.StringAttribute{Computed: true},
						"provider":              schema.StringAttribute{Computed: true},
						"account_id":            schema.StringAttribute{Computed: true},
						"account_name":          schema.StringAttribute{Computed: true},
						"status":                schema.StringAttribute{Computed: true},
						"last_authenticated_at": schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *userImportConnectionsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *userImportConnectionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	connections, err := d.client.ListUserImportConnections(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error listing user import connections", err.Error())
		return
	}

	var data userImportConnectionsDataSourceModel
	for _, c := range connections {
		data.Connections = append(data.Connections, userImportConnectionModel{
			ID:                  types.StringValue(c.ID),
			Provider:            types.StringValue(c.Provider),
			AccountID:           stringOrNull(c.AccountID),
			AccountName:         stringOrNull(c.AccountName),
			Status:              stringOrNull(c.Status),
			LastAuthenticatedAt: stringOrNull(c.LastAuthenticatedAt),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
