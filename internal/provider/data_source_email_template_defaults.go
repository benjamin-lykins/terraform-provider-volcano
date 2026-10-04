package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/benjamin-lykins/terraform-provider-volcano/internal/client"
)

var _ datasource.DataSource = &emailTemplateDefaultsDataSource{}
var _ datasource.DataSourceWithConfigure = &emailTemplateDefaultsDataSource{}

func NewEmailTemplateDefaultsDataSource() datasource.DataSource {
	return &emailTemplateDefaultsDataSource{}
}

type emailTemplateDefaultsDataSource struct {
	client *client.Client
}

type emailTemplateDefaultModel struct {
	TemplateType types.String `tfsdk:"template_type"`
	Subject      types.String `tfsdk:"subject"`
	HTMLBody     types.String `tfsdk:"html_body"`
	TextBody     types.String `tfsdk:"text_body"`
}

type emailTemplateDefaultsDataSourceModel struct {
	Templates []emailTemplateDefaultModel `tfsdk:"templates"`
}

func (d *emailTemplateDefaultsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_email_template_defaults"
}

func (d *emailTemplateDefaultsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The platform's built-in default content for each auth email template type.",
		Attributes: map[string]schema.Attribute{
			"templates": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"template_type": schema.StringAttribute{Computed: true},
						"subject":       schema.StringAttribute{Computed: true},
						"html_body":     schema.StringAttribute{Computed: true},
						"text_body":     schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *emailTemplateDefaultsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *emailTemplateDefaultsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	defaults, err := d.client.ListEmailTemplateDefaults(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error listing email template defaults", err.Error())
		return
	}

	var data emailTemplateDefaultsDataSourceModel
	for _, t := range defaults {
		data.Templates = append(data.Templates, emailTemplateDefaultModel{
			TemplateType: types.StringValue(t.TemplateType),
			Subject:      types.StringValue(t.Subject),
			HTMLBody:     types.StringValue(t.HTMLBody),
			TextBody:     types.StringValue(t.TextBody),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
