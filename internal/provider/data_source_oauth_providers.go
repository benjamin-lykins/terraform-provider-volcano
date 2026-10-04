package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/benjamin-lykins/terraform-provider-volcano/internal/client"
)

var _ datasource.DataSource = &oauthProvidersDataSource{}
var _ datasource.DataSourceWithConfigure = &oauthProvidersDataSource{}

func NewOAuthProvidersDataSource() datasource.DataSource {
	return &oauthProvidersDataSource{}
}

type oauthProvidersDataSource struct {
	client *client.Client
}

type oauthProviderInfoModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	DefaultScopes types.List   `tfsdk:"default_scopes"`
}

type oauthProvidersDataSourceModel struct {
	ProjectID types.String             `tfsdk:"project_id"`
	Providers []oauthProviderInfoModel `tfsdk:"providers"`
}

func (d *oauthProvidersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_oauth_providers"
}

func (d *oauthProvidersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "OAuth provider types a project can configure via volcano_oauth_config, with their default scopes. This lists what's available, not what's currently configured.",
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{Required: true},
			"providers": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":             schema.StringAttribute{Computed: true},
						"name":           schema.StringAttribute{Computed: true},
						"default_scopes": schema.ListAttribute{Computed: true, ElementType: types.StringType},
					},
				},
			},
		},
	}
}

func (d *oauthProvidersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *oauthProvidersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data oauthProvidersDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	providers, err := d.client.ListOAuthProviderInfo(ctx, data.ProjectID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error listing OAuth providers", err.Error())
		return
	}

	data.Providers = make([]oauthProviderInfoModel, 0, len(providers))
	for _, p := range providers {
		data.Providers = append(data.Providers, oauthProviderInfoModel{
			ID:            types.StringValue(p.ID),
			Name:          types.StringValue(p.Name),
			DefaultScopes: stringList(ctx, p.DefaultScopes, &resp.Diagnostics),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
