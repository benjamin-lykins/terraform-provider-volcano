package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/benjamin-lykins/terraform-provider-volcano/internal/client"
)

var _ datasource.DataSource = &frontendDeploymentsDataSource{}
var _ datasource.DataSourceWithConfigure = &frontendDeploymentsDataSource{}

func NewFrontendDeploymentsDataSource() datasource.DataSource {
	return &frontendDeploymentsDataSource{}
}

type frontendDeploymentsDataSource struct {
	client *client.Client
}

type frontendDeploymentsDataSourceModel struct {
	ProjectID   types.String              `tfsdk:"project_id"`
	FrontendID  types.String              `tfsdk:"frontend_id"`
	Deployments []functionDeploymentModel `tfsdk:"deployments"`
}

func (d *frontendDeploymentsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_frontend_deployments"
}

func (d *frontendDeploymentsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Deployment history for a frontend.",
		Attributes: map[string]schema.Attribute{
			"project_id":  schema.StringAttribute{Required: true},
			"frontend_id": schema.StringAttribute{Required: true},
			"deployments": schema.ListNestedAttribute{Computed: true, NestedObject: deploymentNestedObject},
		},
	}
}

func (d *frontendDeploymentsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *frontendDeploymentsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data frontendDeploymentsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deployments, err := d.client.ListFrontendDeployments(ctx, data.ProjectID.ValueString(), data.FrontendID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error listing frontend deployments", err.Error())
		return
	}

	data.Deployments = deploymentModelsFromAPI(deployments)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
