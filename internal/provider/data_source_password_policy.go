package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/benjamin-lykins/terraform-provider-volcano/internal/client"
)

var _ datasource.DataSource = &passwordPolicyDataSource{}
var _ datasource.DataSourceWithConfigure = &passwordPolicyDataSource{}

func NewPasswordPolicyDataSource() datasource.DataSource {
	return &passwordPolicyDataSource{}
}

type passwordPolicyDataSource struct {
	client *client.Client
}

type passwordPolicyDataSourceModel struct {
	EffectiveMinLength           types.Int64 `tfsdk:"effective_min_length"`
	MinConfigurableLength        types.Int64 `tfsdk:"min_configurable_length"`
	MaxLength                    types.Int64 `tfsdk:"max_length"`
	RequireLowercase             types.Bool  `tfsdk:"require_lowercase"`
	RequireUppercase             types.Bool  `tfsdk:"require_uppercase"`
	RequireNumbers               types.Bool  `tfsdk:"require_numbers"`
	RequireSpecialChars          types.Bool  `tfsdk:"require_special_chars"`
	CompromisedPasswordsRejected types.Bool  `tfsdk:"compromised_passwords_rejected"`
}

func (d *passwordPolicyDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_password_policy"
}

func (d *passwordPolicyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The platform-wide password policy baseline. A project's volcano_project_auth_config may tighten it further.",
		Attributes: map[string]schema.Attribute{
			"effective_min_length":           schema.Int64Attribute{Computed: true},
			"min_configurable_length":        schema.Int64Attribute{Computed: true},
			"max_length":                     schema.Int64Attribute{Computed: true},
			"require_lowercase":              schema.BoolAttribute{Computed: true},
			"require_uppercase":              schema.BoolAttribute{Computed: true},
			"require_numbers":                schema.BoolAttribute{Computed: true},
			"require_special_chars":          schema.BoolAttribute{Computed: true},
			"compromised_passwords_rejected": schema.BoolAttribute{Computed: true},
		},
	}
}

func (d *passwordPolicyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *passwordPolicyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	p, err := d.client.GetPasswordPolicy(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading password policy", err.Error())
		return
	}

	data := passwordPolicyDataSourceModel{
		EffectiveMinLength:           types.Int64Value(p.EffectiveMinLength),
		MinConfigurableLength:        types.Int64Value(p.MinConfigurableLength),
		MaxLength:                    types.Int64Value(p.MaxLength),
		RequireLowercase:             types.BoolValue(p.RequireLowercase),
		RequireUppercase:             types.BoolValue(p.RequireUppercase),
		RequireNumbers:               types.BoolValue(p.RequireNumbers),
		RequireSpecialChars:          types.BoolValue(p.RequireSpecialChars),
		CompromisedPasswordsRejected: types.BoolValue(p.CompromisedPasswordsRejected),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
