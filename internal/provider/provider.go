// Package provider implements the Terraform provider for Volcano
// (https://docs.volcano.dev). Each resource and data source lives in its
// own file, grouped loosely by the API domain it belongs to (projects,
// databases, functions, frontends, storage, auth, ...).
package provider

import (
	"context"
	"net/http"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/benjamin-lykins/terraform-provider-volcano/internal/client"
)

const (
	envToken    = "VOLCANO_TOKEN"
	envEndpoint = "VOLCANO_API_URL"
)

// Ensure the implementation satisfies the expected interfaces.
var _ provider.Provider = &volcanoProvider{}

type volcanoProvider struct {
	// version is set from the GoReleaser build and reported via Metadata.
	version string
}

// New returns a provider server constructor, as required by
// providerserver.Serve.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &volcanoProvider{version: version}
	}
}

type volcanoProviderModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
	Token    types.String `tfsdk:"token"`
}

func (p *volcanoProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "volcano"
	resp.Version = p.version
}

func (p *volcanoProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Interact with the Volcano platform API (projects, databases, functions, frontends, storage, and related configuration).",
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				Optional:    true,
				Description: "Base URL of the Volcano API. Defaults to https://api.volcano.dev. May also be set via the VOLCANO_API_URL environment variable.",
			},
			"token": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Bearer credential used to authenticate to the Volcano API: a platform token (pk-), project access token (pt-), or service key, depending on which resources are used. May also be set via the VOLCANO_TOKEN environment variable.",
			},
		},
	}
}

func (p *volcanoProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data volcanoProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := os.Getenv(envEndpoint)
	if !data.Endpoint.IsNull() && data.Endpoint.ValueString() != "" {
		endpoint = data.Endpoint.ValueString()
	}
	if endpoint == "" {
		endpoint = client.DefaultEndpoint
	}

	token := os.Getenv(envToken)
	if !data.Token.IsNull() && data.Token.ValueString() != "" {
		token = data.Token.ValueString()
	}
	if token == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Missing Volcano API Token",
			"The provider requires an API token. Set the `token` attribute in the provider configuration or the VOLCANO_TOKEN environment variable.",
		)
		return
	}

	c := client.New(endpoint, token, http.DefaultClient)
	resp.ResourceData = c
	resp.DataSourceData = c
}

func (p *volcanoProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewProjectResource,
		NewProjectAccessTokenResource,
		NewAnonKeyResource,
		NewServiceKeyResource,
		NewProjectVariableResource,
		NewProjectLogoResource,
		NewDatabaseResource,
		NewDatabaseBranchResource,
		NewDatabaseBackupResource,
		NewDatabaseBackupScheduleResource,
		NewDatabaseRestoreResource,
	}
}

func (p *volcanoProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewProjectDataSource,
		NewProjectsDataSource,
		NewDatabaseRegionsDataSource,
		NewPostgresVersionsDataSource,
	}
}
