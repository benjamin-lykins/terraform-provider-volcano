package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/benjamin-lykins/terraform-provider-volcano/internal/client"
)

var _ datasource.DataSource = &userGitInstallationsDataSource{}
var _ datasource.DataSourceWithConfigure = &userGitInstallationsDataSource{}

func NewUserGitInstallationsDataSource() datasource.DataSource {
	return &userGitInstallationsDataSource{}
}

type userGitInstallationsDataSource struct {
	client *client.Client
}

type userGitInstallationModel struct {
	ID                  types.Int64  `tfsdk:"id"`
	AccountLogin        types.String `tfsdk:"account_login"`
	AccountType         types.String `tfsdk:"account_type"`
	RepositorySelection types.String `tfsdk:"repository_selection"`
}

type userGitInstallationsDataSourceModel struct {
	ConnectionID  types.String               `tfsdk:"connection_id"`
	Installations []userGitInstallationModel `tfsdk:"installations"`
}

func (d *userGitInstallationsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_git_installations"
}

func (d *userGitInstallationsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "GitHub App installations reachable through an account git connection, used as installation_id for volcano_project_git_connection.",
		Attributes: map[string]schema.Attribute{
			"connection_id": schema.StringAttribute{Required: true},
			"installations": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":                   schema.Int64Attribute{Computed: true},
						"account_login":        schema.StringAttribute{Computed: true},
						"account_type":         schema.StringAttribute{Computed: true},
						"repository_selection": schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *userGitInstallationsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *userGitInstallationsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data userGitInstallationsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	installations, err := d.client.ListUserGitInstallations(ctx, data.ConnectionID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error listing user git installations", err.Error())
		return
	}

	data.Installations = make([]userGitInstallationModel, 0, len(installations))
	for _, i := range installations {
		data.Installations = append(data.Installations, userGitInstallationModel{
			ID:                  types.Int64Value(i.ID),
			AccountLogin:        stringOrNull(i.AccountLogin),
			AccountType:         stringOrNull(i.AccountType),
			RepositorySelection: stringOrNull(i.RepositorySelection),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
