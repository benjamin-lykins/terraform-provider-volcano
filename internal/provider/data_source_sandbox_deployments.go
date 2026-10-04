package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/benjamin-lykins/terraform-provider-volcano/internal/client"
)

var _ datasource.DataSource = &sandboxDeploymentsDataSource{}
var _ datasource.DataSourceWithConfigure = &sandboxDeploymentsDataSource{}

func NewSandboxDeploymentsDataSource() datasource.DataSource {
	return &sandboxDeploymentsDataSource{}
}

type sandboxDeploymentsDataSource struct {
	client *client.Client
}

type sandboxDeploymentsDataSourceModel struct {
	ProjectID   types.String              `tfsdk:"project_id"`
	SandboxID   types.String              `tfsdk:"sandbox_id"`
	Deployments []functionDeploymentModel `tfsdk:"deployments"`
}

func (d *sandboxDeploymentsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sandbox_deployments"
}

func (d *sandboxDeploymentsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Deployment history for a sandbox.",
		Attributes: map[string]schema.Attribute{
			"project_id":  schema.StringAttribute{Required: true},
			"sandbox_id":  schema.StringAttribute{Required: true},
			"deployments": schema.ListNestedAttribute{Computed: true, NestedObject: deploymentNestedObject},
		},
	}
}

func (d *sandboxDeploymentsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *sandboxDeploymentsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data sandboxDeploymentsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deployments, err := d.client.ListSandboxDeployments(ctx, data.ProjectID.ValueString(), data.SandboxID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error listing sandbox deployments", err.Error())
		return
	}

	data.Deployments = deploymentModelsFromAPI(deployments)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
