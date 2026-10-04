package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/benjamin-lykins/terraform-provider-volcano/internal/client"
)

var _ datasource.DataSource = &projectHealthDataSource{}
var _ datasource.DataSourceWithConfigure = &projectHealthDataSource{}

func NewProjectHealthDataSource() datasource.DataSource {
	return &projectHealthDataSource{}
}

type projectHealthDataSource struct {
	client *client.Client
}

type projectHealthCheckModel struct {
	ID           types.String `tfsdk:"id"`
	Category     types.String `tfsdk:"category"`
	Status       types.String `tfsdk:"status"`
	ReasonCode   types.String `tfsdk:"reason_code"`
	ResourceType types.String `tfsdk:"resource_type"`
	ResourceID   types.String `tfsdk:"resource_id"`
	ResourceName types.String `tfsdk:"resource_name"`
}

var projectHealthCheckNestedObject = schema.NestedAttributeObject{
	Attributes: map[string]schema.Attribute{
		"id":            schema.StringAttribute{Computed: true},
		"category":      schema.StringAttribute{Computed: true},
		"status":        schema.StringAttribute{Computed: true},
		"reason_code":   schema.StringAttribute{Computed: true},
		"resource_type": schema.StringAttribute{Computed: true},
		"resource_id":   schema.StringAttribute{Computed: true},
		"resource_name": schema.StringAttribute{Computed: true},
	},
}

func projectHealthCheckModelsFromAPI(checks []client.ProjectHealthCheck) []projectHealthCheckModel {
	out := make([]projectHealthCheckModel, 0, len(checks))
	for _, c := range checks {
		out = append(out, projectHealthCheckModel{
			ID:           types.StringValue(c.ID),
			Category:     stringOrNull(c.Category),
			Status:       stringOrNull(c.Status),
			ReasonCode:   stringOrNull(c.ReasonCode),
			ResourceType: stringOrNull(c.Scope.Resource.Type),
			ResourceID:   stringOrNull(c.Scope.Resource.ID),
			ResourceName: stringOrNull(c.Scope.Resource.Name),
		})
	}
	return out
}

type projectHealthDataSourceModel struct {
	ProjectID    types.String              `tfsdk:"project_id"`
	Status       types.String              `tfsdk:"status"`
	DataStatus   types.String              `tfsdk:"data_status"`
	ObservedAt   types.String              `tfsdk:"observed_at"`
	FreshThrough types.String              `tfsdk:"fresh_through"`
	Checks       []projectHealthCheckModel `tfsdk:"checks"`
	Findings     []projectHealthCheckModel `tfsdk:"findings"`
}

func (d *projectHealthDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_health"
}

func (d *projectHealthDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A project's current health status, derived from automated lifecycle/storage checks across its resources.",
		Attributes: map[string]schema.Attribute{
			"project_id":    schema.StringAttribute{Required: true},
			"status":        schema.StringAttribute{Computed: true, Description: "healthy, degraded, critical, or unknown."},
			"data_status":   schema.StringAttribute{Computed: true, Description: "complete, partial, no_data, or stale."},
			"observed_at":   schema.StringAttribute{Computed: true},
			"fresh_through": schema.StringAttribute{Computed: true},
			"checks":        schema.ListNestedAttribute{Computed: true, NestedObject: projectHealthCheckNestedObject},
			"findings":      schema.ListNestedAttribute{Computed: true, NestedObject: projectHealthCheckNestedObject, Description: "Top findings ordered by severity."},
		},
	}
}

func (d *projectHealthDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *projectHealthDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data projectHealthDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	health, err := d.client.GetProjectHealth(ctx, data.ProjectID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading project health", err.Error())
		return
	}

	data.Status = stringOrNull(health.Status)
	data.DataStatus = stringOrNull(health.DataStatus)
	data.ObservedAt = stringOrNull(health.ObservedAt)
	data.FreshThrough = stringOrNull(health.FreshThrough)
	data.Checks = projectHealthCheckModelsFromAPI(health.Checks)
	data.Findings = projectHealthCheckModelsFromAPI(health.Findings)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
