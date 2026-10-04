package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/benjamin-lykins/terraform-provider-volcano/internal/client"
)

type deploymentModel struct {
	ID           types.String `tfsdk:"id"`
	ProjectID    types.String `tfsdk:"project_id"`
	Resource     types.String `tfsdk:"resource"`
	Operation    types.String `tfsdk:"operation"`
	Status       types.String `tfsdk:"status"`
	Progress     types.Int64  `tfsdk:"progress"`
	DeploySource types.String `tfsdk:"deploy_source"`
	ErrorMessage types.String `tfsdk:"error_message"`
	CreatedAt    types.String `tfsdk:"created_at"`
	UpdatedAt    types.String `tfsdk:"updated_at"`
	CompletedAt  types.String `tfsdk:"completed_at"`
}

var deploymentFullNestedObject = schema.NestedAttributeObject{
	Attributes: map[string]schema.Attribute{
		"id":            schema.StringAttribute{Computed: true},
		"project_id":    schema.StringAttribute{Computed: true},
		"resource":      schema.StringAttribute{Computed: true},
		"operation":     schema.StringAttribute{Computed: true},
		"status":        schema.StringAttribute{Computed: true},
		"progress":      schema.Int64Attribute{Computed: true},
		"deploy_source": schema.StringAttribute{Computed: true},
		"error_message": schema.StringAttribute{Computed: true},
		"created_at":    schema.StringAttribute{Computed: true},
		"updated_at":    schema.StringAttribute{Computed: true},
		"completed_at":  schema.StringAttribute{Computed: true},
	},
}

func deploymentFullModelsFromAPI(deployments []client.Deployment) []deploymentModel {
	out := make([]deploymentModel, 0, len(deployments))
	for _, d := range deployments {
		out = append(out, deploymentModel{
			ID:           types.StringValue(d.ID),
			ProjectID:    types.StringValue(d.ProjectID),
			Resource:     stringOrNull(d.Resource),
			Operation:    stringOrNull(d.Operation),
			Status:       stringOrNull(d.Status),
			Progress:     types.Int64Value(d.Progress),
			DeploySource: stringOrNull(d.DeploySource),
			ErrorMessage: stringOrNull(d.ErrorMessage),
			CreatedAt:    stringOrNull(d.CreatedAt),
			UpdatedAt:    stringOrNull(d.UpdatedAt),
			CompletedAt:  stringOrNull(d.CompletedAt),
		})
	}
	return out
}

var _ datasource.DataSource = &projectDeploymentsDataSource{}
var _ datasource.DataSourceWithConfigure = &projectDeploymentsDataSource{}

func NewProjectDeploymentsDataSource() datasource.DataSource {
	return &projectDeploymentsDataSource{}
}

type projectDeploymentsDataSource struct {
	client *client.Client
}

type projectDeploymentsDataSourceModel struct {
	ProjectID   types.String      `tfsdk:"project_id"`
	Deployments []deploymentModel `tfsdk:"deployments"`
}

func (d *projectDeploymentsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_deployments"
}

func (d *projectDeploymentsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Deployment history across every deployable resource (functions, durable functions, frontends) in a project.",
		Attributes: map[string]schema.Attribute{
			"project_id":  schema.StringAttribute{Required: true},
			"deployments": schema.ListNestedAttribute{Computed: true, NestedObject: deploymentFullNestedObject},
		},
	}
}

func (d *projectDeploymentsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *projectDeploymentsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data projectDeploymentsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deployments, err := d.client.ListProjectDeployments(ctx, data.ProjectID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error listing project deployments", err.Error())
		return
	}

	data.Deployments = deploymentFullModelsFromAPI(deployments)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
