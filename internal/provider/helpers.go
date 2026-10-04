package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/benjamin-lykins/terraform-provider-volcano/internal/client"
)

var pathRootID = path.Root("id")

// stringList converts a Go string slice into a framework types.List,
// appending any diagnostics to diags. A nil/empty slice becomes an empty
// (non-null) list, matching how the API represents "no regions selected".
func stringList(ctx context.Context, values []string, diags *diag.Diagnostics) types.List {
	l, d := types.ListValueFrom(ctx, types.StringType, values)
	diags.Append(d...)
	return l
}

func projectModelFromAPI(ctx context.Context, p *client.Project, templateID, initialPrompt types.String, diags *diag.Diagnostics) projectResourceModel {
	return projectResourceModel{
		ID:              types.StringValue(p.ID),
		Name:            types.StringValue(p.Name),
		Status:          types.StringValue(p.Status),
		Plan:            stringOrNull(p.Plan),
		AllRegions:      types.BoolValue(p.AllRegions),
		SelectedRegions: stringList(ctx, p.SelectedRegions, diags),
		TemplateID:      templateID,
		InitialPrompt:   initialPrompt,
		LogoURL:         stringOrNull(p.LogoURL),
		CreatedAt:       types.StringValue(p.CreatedAt),
		UpdatedAt:       types.StringValue(p.UpdatedAt),
	}
}

// stringOrNull returns a null StringValue for an empty Go string, matching
// the convention that optional/absent API fields should be represented as
// null in state rather than an empty string.
func stringOrNull(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}
