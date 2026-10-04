package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/benjamin-lykins/terraform-provider-volcano/internal/client"
)

var pathRootID = path.Root("id")
var pathProjectID = path.Root("project_id")

// importStateCompositeID parses a comma-separated import ID into the given
// attribute names, in order (e.g. "project_id,name" for a resource keyed by
// both). Used by every resource whose natural identity is a child of a
// project rather than a standalone synthetic ID.
func importStateCompositeID(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse, attrNames ...string) {
	parts := strings.Split(req.ID, ",")
	if len(parts) != len(attrNames) {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier of the form %q, got: %q", strings.Join(attrNames, ","), req.ID),
		)
		return
	}
	for i, name := range attrNames {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(name), parts[i])...)
	}
}

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

func anonKeyModelFromAPI(ctx context.Context, projectID string, k *client.AnonKey, diags *diag.Diagnostics) anonKeyResourceModel {
	return anonKeyResourceModel{
		ID:          types.StringValue(k.ID),
		ProjectID:   types.StringValue(projectID),
		Name:        types.StringValue(k.Name),
		Permissions: stringList(ctx, k.Permissions, diags),
		IsDefault:   types.BoolValue(k.IsDefault),
		KeyValue:    types.StringValue(k.KeyValue),
		CreatedAt:   stringOrNull(k.CreatedAt),
	}
}

func serviceKeyModelFromAPI(ctx context.Context, projectID string, k *client.ServiceKey, diags *diag.Diagnostics) serviceKeyResourceModel {
	return serviceKeyResourceModel{
		ID:          types.StringValue(k.ID),
		ProjectID:   types.StringValue(projectID),
		Name:        types.StringValue(k.Name),
		Permissions: stringList(ctx, k.Permissions, diags),
		KeyPrefix:   stringOrNull(k.KeyPrefix),
		KeyValue:    types.StringValue(k.KeyValue),
		CreatedAt:   stringOrNull(k.CreatedAt),
		UpdatedAt:   stringOrNull(k.UpdatedAt),
	}
}

func projectVariableModelFromAPI(projectID string, v *client.Variable) projectVariableResourceModel {
	return projectVariableResourceModel{
		ID:        types.StringValue(v.ID),
		ProjectID: types.StringValue(projectID),
		Name:      types.StringValue(v.Name),
		Value:     types.StringValue(v.Value),
		Shared:    types.BoolValue(v.Shared),
		Status:    stringOrNull(v.Status),
		CreatedAt: stringOrNull(v.CreatedAt),
		UpdatedAt: stringOrNull(v.UpdatedAt),
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
