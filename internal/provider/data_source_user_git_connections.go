package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/benjamin-lykins/terraform-provider-volcano/internal/client"
)

var _ datasource.DataSource = &userGitConnectionsDataSource{}
var _ datasource.DataSourceWithConfigure = &userGitConnectionsDataSource{}

func NewUserGitConnectionsDataSource() datasource.DataSource {
	return &userGitConnectionsDataSource{}
}

type userGitConnectionsDataSource struct {
	client *client.Client
}

type userGitConnectionModel struct {
	ID                  types.String `tfsdk:"id"`
	Provider            types.String `tfsdk:"provider"`
	ProviderLogin       types.String `tfsdk:"provider_login"`
	Status              types.String `tfsdk:"status"`
	LastAuthenticatedAt types.String `tfsdk:"last_authenticated_at"`
}

type userGitConnectionsDataSourceModel struct {
	Connections []userGitConnectionModel `tfsdk:"connections"`
}

func (d *userGitConnectionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_git_connections"
}

func (d *userGitConnectionsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The account's git provider connections (e.g. a GitHub App installation owner), used as the connection_id for volcano_project_git_connection. Connecting a new one requires an interactive OAuth/installation flow this provider cannot drive - manage that through the dashboard or CLI, then reference the result here.",
		Attributes: map[string]schema.Attribute{
			"connections": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":                    schema.StringAttribute{Computed: true},
						"provider":              schema.StringAttribute{Computed: true},
						"provider_login":        schema.StringAttribute{Computed: true},
						"status":                schema.StringAttribute{Computed: true},
						"last_authenticated_at": schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *userGitConnectionsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *userGitConnectionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	connections, err := d.client.ListUserGitConnections(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error listing user git connections", err.Error())
		return
	}

	var data userGitConnectionsDataSourceModel
	for _, c := range connections {
		data.Connections = append(data.Connections, userGitConnectionModel{
			ID:                  types.StringValue(c.ID),
			Provider:            types.StringValue(c.Provider),
			ProviderLogin:       stringOrNull(c.ProviderLogin),
			Status:              stringOrNull(c.Status),
			LastAuthenticatedAt: stringOrNull(c.LastAuthenticatedAt),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
