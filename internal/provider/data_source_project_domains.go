package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/benjamin-lykins/terraform-provider-volcano/internal/client"
)

var _ datasource.DataSource = &projectDomainsDataSource{}
var _ datasource.DataSourceWithConfigure = &projectDomainsDataSource{}

func NewProjectDomainsDataSource() datasource.DataSource {
	return &projectDomainsDataSource{}
}

type projectDomainsDataSource struct {
	client *client.Client
}

type projectDomainModel struct {
	Domain             types.String `tfsdk:"domain"`
	TLSMode            types.String `tfsdk:"tls_mode"`
	DomainStatus       types.String `tfsdk:"domain_status"`
	VerificationStatus types.String `tfsdk:"verification_status"`
	EffectiveURLs      types.List   `tfsdk:"effective_urls"`
	CreatedAt          types.String `tfsdk:"created_at"`
	UpdatedAt          types.String `tfsdk:"updated_at"`
}

type projectDomainsDataSourceModel struct {
	ProjectID types.String         `tfsdk:"project_id"`
	Domains   []projectDomainModel `tfsdk:"domains"`
}

func (d *projectDomainsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_domains"
}

func (d *projectDomainsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Every custom domain attached to any frontend in a project.",
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{Required: true},
			"domains": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"domain":              schema.StringAttribute{Computed: true},
						"tls_mode":            schema.StringAttribute{Computed: true},
						"domain_status":       schema.StringAttribute{Computed: true},
						"verification_status": schema.StringAttribute{Computed: true},
						"effective_urls":      schema.ListAttribute{Computed: true, ElementType: types.StringType},
						"created_at":          schema.StringAttribute{Computed: true},
						"updated_at":          schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *projectDomainsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *projectDomainsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data projectDomainsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domains, err := d.client.ListProjectDomains(ctx, data.ProjectID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error listing project domains", err.Error())
		return
	}

	data.Domains = make([]projectDomainModel, 0, len(domains))
	for _, dom := range domains {
		data.Domains = append(data.Domains, projectDomainModel{
			Domain:             types.StringValue(dom.Domain),
			TLSMode:            stringOrNull(dom.TLSMode),
			DomainStatus:       stringOrNull(dom.DomainStatus),
			VerificationStatus: stringOrNull(dom.VerificationStatus),
			EffectiveURLs:      stringList(ctx, dom.EffectiveURLs, &resp.Diagnostics),
			CreatedAt:          stringOrNull(dom.CreatedAt),
			UpdatedAt:          stringOrNull(dom.UpdatedAt),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
