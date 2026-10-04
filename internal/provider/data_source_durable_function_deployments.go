package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	"github.com/benjamin-lykins/terraform-provider-volcano/internal/client"
)

var _ datasource.DataSource = &durableFunctionDeploymentsDataSource{}
var _ datasource.DataSourceWithConfigure = &durableFunctionDeploymentsDataSource{}

func NewDurableFunctionDeploymentsDataSource() datasource.DataSource {
	return &durableFunctionDeploymentsDataSource{}
}

type durableFunctionDeploymentsDataSource struct {
	client *client.Client
}

func (d *durableFunctionDeploymentsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_durable_function_deployments"
}

func (d *durableFunctionDeploymentsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Deployment history for a durable function.",
		Attributes: map[string]schema.Attribute{
			"project_id":  schema.StringAttribute{Required: true},
			"function_id": schema.StringAttribute{Required: true},
			"deployments": schema.ListNestedAttribute{Computed: true, NestedObject: deploymentNestedObject},
		},
	}
}

func (d *durableFunctionDeploymentsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *durableFunctionDeploymentsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data functionDeploymentsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deployments, err := d.client.ListDurableFunctionDeployments(ctx, data.ProjectID.ValueString(), data.FunctionID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error listing durable function deployments", err.Error())
		return
	}

	data.Deployments = deploymentModelsFromAPI(deployments)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
