package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/benjamin-lykins/terraform-provider-volcano/internal/client"
)

type functionDeploymentModel struct {
	ID        types.String `tfsdk:"id"`
	Status    types.String `tfsdk:"status"`
	CreatedAt types.String `tfsdk:"created_at"`
}

var deploymentNestedObject = schema.NestedAttributeObject{
	Attributes: map[string]schema.Attribute{
		"id":         schema.StringAttribute{Computed: true},
		"status":     schema.StringAttribute{Computed: true},
		"created_at": schema.StringAttribute{Computed: true},
	},
}

func deploymentModelsFromAPI(deployments []client.FunctionDeployment) []functionDeploymentModel {
	out := make([]functionDeploymentModel, 0, len(deployments))
	for _, d := range deployments {
		out = append(out, functionDeploymentModel{
			ID:        types.StringValue(d.ID),
			Status:    stringOrNull(d.Status),
			CreatedAt: stringOrNull(d.CreatedAt),
		})
	}
	return out
}

var _ datasource.DataSource = &functionDeploymentsDataSource{}
var _ datasource.DataSourceWithConfigure = &functionDeploymentsDataSource{}

func NewFunctionDeploymentsDataSource() datasource.DataSource {
	return &functionDeploymentsDataSource{}
}

type functionDeploymentsDataSource struct {
	client *client.Client
}

type functionDeploymentsDataSourceModel struct {
	ProjectID   types.String              `tfsdk:"project_id"`
	FunctionID  types.String              `tfsdk:"function_id"`
	Deployments []functionDeploymentModel `tfsdk:"deployments"`
}

func (d *functionDeploymentsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_function_deployments"
}

func (d *functionDeploymentsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Deployment history for a function.",
		Attributes: map[string]schema.Attribute{
			"project_id":  schema.StringAttribute{Required: true},
			"function_id": schema.StringAttribute{Required: true},
			"deployments": schema.ListNestedAttribute{Computed: true, NestedObject: deploymentNestedObject},
		},
	}
}

func (d *functionDeploymentsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *functionDeploymentsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data functionDeploymentsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deployments, err := d.client.ListFunctionDeployments(ctx, data.ProjectID.ValueString(), data.FunctionID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error listing function deployments", err.Error())
		return
	}

	data.Deployments = deploymentModelsFromAPI(deployments)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
