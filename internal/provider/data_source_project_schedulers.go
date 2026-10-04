package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/benjamin-lykins/terraform-provider-volcano/internal/client"
)

var _ datasource.DataSource = &projectSchedulersDataSource{}
var _ datasource.DataSourceWithConfigure = &projectSchedulersDataSource{}

func NewProjectSchedulersDataSource() datasource.DataSource {
	return &projectSchedulersDataSource{}
}

type projectSchedulersDataSource struct {
	client *client.Client
}

type projectSchedulerModel struct {
	ID             types.String `tfsdk:"id"`
	FunctionID     types.String `tfsdk:"function_id"`
	FunctionKind   types.String `tfsdk:"function_kind"`
	Name           types.String `tfsdk:"name"`
	Enabled        types.Bool   `tfsdk:"enabled"`
	CronExpression types.String `tfsdk:"cron_expression"`
	NextRunAt      types.String `tfsdk:"next_run_at"`
	RunCount       types.Int64  `tfsdk:"run_count"`
}

type projectSchedulersDataSourceModel struct {
	ProjectID  types.String            `tfsdk:"project_id"`
	Schedulers []projectSchedulerModel `tfsdk:"schedulers"`
}

func (d *projectSchedulersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_schedulers"
}

func (d *projectSchedulersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Every scheduler across every function and durable function in a project.",
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{Required: true},
			"schedulers": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":              schema.StringAttribute{Computed: true},
						"function_id":     schema.StringAttribute{Computed: true},
						"function_kind":   schema.StringAttribute{Computed: true},
						"name":            schema.StringAttribute{Computed: true},
						"enabled":         schema.BoolAttribute{Computed: true},
						"cron_expression": schema.StringAttribute{Computed: true},
						"next_run_at":     schema.StringAttribute{Computed: true},
						"run_count":       schema.Int64Attribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *projectSchedulersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *projectSchedulersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data projectSchedulersDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	schedulers, err := d.client.ListProjectSchedulers(ctx, data.ProjectID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error listing project schedulers", err.Error())
		return
	}

	data.Schedulers = make([]projectSchedulerModel, 0, len(schedulers))
	for _, s := range schedulers {
		data.Schedulers = append(data.Schedulers, projectSchedulerModel{
			ID:             types.StringValue(s.ID),
			FunctionID:     types.StringValue(s.FunctionID),
			FunctionKind:   stringOrNull(s.FunctionKind),
			Name:           types.StringValue(s.Name),
			Enabled:        types.BoolValue(s.Enabled),
			CronExpression: stringOrNull(s.CronExpression),
			NextRunAt:      stringOrNull(s.NextRunAt),
			RunCount:       types.Int64Value(s.RunCount),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
