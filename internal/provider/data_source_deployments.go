package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	"github.com/benjamin-lykins/terraform-provider-volcano/internal/client"
)

var _ datasource.DataSource = &accountDeploymentsDataSource{}
var _ datasource.DataSourceWithConfigure = &accountDeploymentsDataSource{}

func NewAccountDeploymentsDataSource() datasource.DataSource {
	return &accountDeploymentsDataSource{}
}

type accountDeploymentsDataSource struct {
	client *client.Client
}

type accountDeploymentsDataSourceModel struct {
	Deployments []deploymentModel `tfsdk:"deployments"`
}

func (d *accountDeploymentsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_deployments"
}

func (d *accountDeploymentsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Deployment history across every project in the account reachable by the configured token.",
		Attributes: map[string]schema.Attribute{
			"deployments": schema.ListNestedAttribute{Computed: true, NestedObject: deploymentFullNestedObject},
		},
	}
}

func (d *accountDeploymentsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *accountDeploymentsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	deployments, err := d.client.ListAccountDeployments(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error listing account deployments", err.Error())
		return
	}

	data := accountDeploymentsDataSourceModel{Deployments: deploymentFullModelsFromAPI(deployments)}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
