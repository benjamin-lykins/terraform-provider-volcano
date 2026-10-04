package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/benjamin-lykins/terraform-provider-volcano/internal/client"
)

var _ datasource.DataSource = &sandboxPresetsDataSource{}
var _ datasource.DataSourceWithConfigure = &sandboxPresetsDataSource{}

func NewSandboxPresetsDataSource() datasource.DataSource {
	return &sandboxPresetsDataSource{}
}

type sandboxPresetsDataSource struct {
	client *client.Client
}

type sandboxPresetModel struct {
	ID       types.String `tfsdk:"id"`
	Runtime  types.String `tfsdk:"runtime"`
	Version  types.String `tfsdk:"version"`
	MemoryMB types.Int64  `tfsdk:"memory_mb"`
	Regions  types.List   `tfsdk:"regions"`
}

type sandboxPresetsDataSourceModel struct {
	Presets []sandboxPresetModel `tfsdk:"presets"`
}

func (d *sandboxPresetsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sandbox_presets"
}

func (d *sandboxPresetsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Size/runtime presets available for volcano_sandbox.",
		Attributes: map[string]schema.Attribute{
			"presets": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":        schema.StringAttribute{Computed: true},
						"runtime":   schema.StringAttribute{Computed: true},
						"version":   schema.StringAttribute{Computed: true},
						"memory_mb": schema.Int64Attribute{Computed: true},
						"regions":   schema.ListAttribute{Computed: true, ElementType: types.StringType},
					},
				},
			},
		},
	}
}

func (d *sandboxPresetsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *sandboxPresetsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	presets, err := d.client.ListSandboxPresets(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error listing sandbox presets", err.Error())
		return
	}

	var data sandboxPresetsDataSourceModel
	for _, p := range presets {
		data.Presets = append(data.Presets, sandboxPresetModel{
			ID:       types.StringValue(p.ID),
			Runtime:  stringOrNull(p.Runtime),
			Version:  stringOrNull(p.Version),
			MemoryMB: types.Int64Value(p.MemoryMB),
			Regions:  stringList(ctx, p.Regions, &resp.Diagnostics),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
