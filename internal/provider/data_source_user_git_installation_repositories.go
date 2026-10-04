package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/benjamin-lykins/terraform-provider-volcano/internal/client"
)

var _ datasource.DataSource = &userGitInstallationRepositoriesDataSource{}
var _ datasource.DataSourceWithConfigure = &userGitInstallationRepositoriesDataSource{}

func NewUserGitInstallationRepositoriesDataSource() datasource.DataSource {
	return &userGitInstallationRepositoriesDataSource{}
}

type userGitInstallationRepositoriesDataSource struct {
	client *client.Client
}

type userGitRepositoryModel struct {
	ID            types.Int64  `tfsdk:"id"`
	FullName      types.String `tfsdk:"full_name"`
	DefaultBranch types.String `tfsdk:"default_branch"`
	Private       types.Bool   `tfsdk:"private"`
	IsEmpty       types.Bool   `tfsdk:"is_empty"`
}

type userGitInstallationRepositoriesDataSourceModel struct {
	ConnectionID   types.String             `tfsdk:"connection_id"`
	InstallationID types.Int64              `tfsdk:"installation_id"`
	Repositories   []userGitRepositoryModel `tfsdk:"repositories"`
}

func (d *userGitInstallationRepositoriesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_git_installation_repositories"
}

func (d *userGitInstallationRepositoriesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Repositories reachable through a GitHub App installation, used to find repository_id/repo_full_name for volcano_project_git_connection.",
		Attributes: map[string]schema.Attribute{
			"connection_id":   schema.StringAttribute{Required: true},
			"installation_id": schema.Int64Attribute{Required: true},
			"repositories": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":             schema.Int64Attribute{Computed: true},
						"full_name":      schema.StringAttribute{Computed: true},
						"default_branch": schema.StringAttribute{Computed: true},
						"private":        schema.BoolAttribute{Computed: true},
						"is_empty":       schema.BoolAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *userGitInstallationRepositoriesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *userGitInstallationRepositoriesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data userGitInstallationRepositoriesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	repos, err := d.client.ListUserGitInstallationRepositories(ctx, data.ConnectionID.ValueString(), data.InstallationID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Error listing user git installation repositories", err.Error())
		return
	}

	data.Repositories = make([]userGitRepositoryModel, 0, len(repos))
	for _, r := range repos {
		data.Repositories = append(data.Repositories, userGitRepositoryModel{
			ID:            types.Int64Value(r.ID),
			FullName:      types.StringValue(r.FullName),
			DefaultBranch: stringOrNull(r.DefaultBranch),
			Private:       types.BoolValue(r.Private),
			IsEmpty:       types.BoolValue(r.IsEmpty),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
