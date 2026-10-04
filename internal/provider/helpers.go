package provider

import (
	"context"
	"encoding/json"
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

func databaseModelFromAPI(d *client.Database) databaseResourceModel {
	return databaseResourceModel{
		ID:               types.StringValue(d.ID),
		ProjectID:        types.StringValue(d.ProjectID),
		Name:             types.StringValue(d.Name),
		Region:           types.StringValue(d.Region),
		PGVersion:        types.StringValue(d.PGVersion),
		DatabaseType:     types.StringValue(d.DatabaseType),
		Status:           types.StringValue(d.Status),
		ConnectionString: stringOrNull(d.ConnectionString),
		CreatedAt:        stringOrNull(d.CreatedAt),
		UpdatedAt:        stringOrNull(d.UpdatedAt),
	}
}

func databaseBranchModelFromAPI(projectID, databaseName string, b *client.DatabaseBranch) databaseBranchResourceModel {
	return databaseBranchResourceModel{
		ID:               types.StringValue(b.ID),
		ProjectID:        types.StringValue(projectID),
		DatabaseName:     types.StringValue(databaseName),
		Name:             types.StringValue(b.Name),
		TTLSeconds:       types.Int64Value(b.TTLSeconds),
		Status:           stringOrNull(b.Status),
		ConnectionString: stringOrNull(b.ConnectionString),
		ExpiresAt:        stringOrNull(b.ExpiresAt),
		CreatedAt:        stringOrNull(b.CreatedAt),
	}
}

func databaseBackupModelFromAPI(projectID, databaseName string, b *client.DatabaseBackup) databaseBackupResourceModel {
	return databaseBackupResourceModel{
		ProjectID:    types.StringValue(projectID),
		DatabaseName: types.StringValue(databaseName),
		Name:         types.StringValue(b.Name),
		Source:       stringOrNull(b.Source),
		SizeBytes:    types.Int64Value(b.SizeBytes),
		CreatedAt:    stringOrNull(b.CreatedAt),
		ExpiresAt:    stringOrNull(b.ExpiresAt),
	}
}

func databaseRestoreModelFromAPI(projectID, databaseName string, r *client.DatabaseRestore) databaseRestoreResourceModel {
	return databaseRestoreResourceModel{
		ID:           types.StringValue(r.ID),
		ProjectID:    types.StringValue(projectID),
		DatabaseName: types.StringValue(databaseName),
		BackupName:   stringOrNull(r.BackupName),
		RestoreTo:    stringOrNull(r.RestoreTo),
		Kind:         stringOrNull(r.Kind),
		Status:       stringOrNull(r.Status),
		Error:        stringOrNull(r.Error),
		CreatedAt:    stringOrNull(r.CreatedAt),
		CompletedAt:  stringOrNull(r.CompletedAt),
	}
}

// functionModelFromAPI builds state for a function resource. variableScope
// and variables are never echoed back by the API (write-only at create
// time), so the caller threads through the prior plan/state value, the
// same way the project resource handles template_id.
func functionModelFromAPI(ctx context.Context, projectID string, source, variableScope types.String, variables types.List, f *client.Function, diags *diag.Diagnostics) functionResourceModel {
	return functionResourceModel{
		ID:              types.StringValue(f.ID),
		ProjectID:       types.StringValue(projectID),
		Name:            types.StringValue(f.Name),
		Runtime:         types.StringValue(f.Runtime),
		Source:          source,
		Handler:         stringOrNull(f.Handler),
		HTTPAuthMode:    stringOrNull(f.HTTPAuthMode),
		InvocationMode:  stringOrNull(f.InvocationMode),
		IsPublic:        types.BoolValue(f.IsPublic),
		OpenAPISpec:     stringOrNull(f.OpenAPISpec),
		VariableScope:   variableScope,
		Variables:       variables,
		Status:          stringOrNull(f.Status),
		InvokeURL:       stringOrNull(f.InvokeURL),
		DeployedRegions: stringList(ctx, f.DeployedRegions, diags),
		CreatedAt:       stringOrNull(f.CreatedAt),
		UpdatedAt:       stringOrNull(f.UpdatedAt),
	}
}

func durableFunctionModelFromAPI(ctx context.Context, projectID string, source, variableScope types.String, variables types.List, f *client.DurableFunction, diags *diag.Diagnostics) durableFunctionResourceModel {
	return durableFunctionResourceModel{
		ID:              types.StringValue(f.ID),
		ProjectID:       types.StringValue(projectID),
		Name:            types.StringValue(f.Name),
		Runtime:         types.StringValue(f.Runtime),
		Source:          source,
		Handler:         stringOrNull(f.Handler),
		IsPublic:        types.BoolValue(f.IsPublic),
		VariableScope:   variableScope,
		Variables:       variables,
		Status:          stringOrNull(f.Status),
		DeployedRegions: stringList(ctx, f.DeployedRegions, diags),
		CreatedAt:       stringOrNull(f.CreatedAt),
		UpdatedAt:       stringOrNull(f.UpdatedAt),
	}
}

func schedulerCreateRequestFromModel(ctx context.Context, name types.String, enabled types.Bool, cron, payload types.String, regions types.List, diags *diag.Diagnostics) (client.CreateFunctionSchedulerRequest, bool) {
	in := client.CreateFunctionSchedulerRequest{
		Name:           name.ValueString(),
		CronExpression: cron.ValueString(),
	}
	if !enabled.IsNull() && !enabled.IsUnknown() {
		v := enabled.ValueBool()
		in.Enabled = &v
	}
	if !payload.IsNull() && !payload.IsUnknown() && payload.ValueString() != "" {
		var m map[string]any
		if err := json.Unmarshal([]byte(payload.ValueString()), &m); err != nil {
			diags.AddError("Invalid payload", "payload must be a JSON-encoded object: "+err.Error())
			return in, false
		}
		in.Payload = m
	}
	if !regions.IsNull() && !regions.IsUnknown() {
		var r []string
		diags.Append(regions.ElementsAs(ctx, &r, false)...)
		if diags.HasError() {
			return in, false
		}
		in.Regions = r
	}
	return in, true
}

func schedulerUpdateRequestFromModel(ctx context.Context, plan, state functionSchedulerResourceModel, diags *diag.Diagnostics) (client.UpdateFunctionSchedulerRequest, bool) {
	in := client.UpdateFunctionSchedulerRequest{}
	if !plan.Name.Equal(state.Name) {
		v := plan.Name.ValueString()
		in.Name = &v
	}
	if !plan.Enabled.Equal(state.Enabled) {
		v := plan.Enabled.ValueBool()
		in.Enabled = &v
	}
	if !plan.CronExpression.Equal(state.CronExpression) {
		v := plan.CronExpression.ValueString()
		in.CronExpression = &v
	}
	if !plan.Payload.Equal(state.Payload) && !plan.Payload.IsNull() && plan.Payload.ValueString() != "" {
		var m map[string]any
		if err := json.Unmarshal([]byte(plan.Payload.ValueString()), &m); err != nil {
			diags.AddError("Invalid payload", "payload must be a JSON-encoded object: "+err.Error())
			return in, false
		}
		in.Payload = m
	}
	if !plan.Regions.Equal(state.Regions) {
		var r []string
		diags.Append(plan.Regions.ElementsAs(ctx, &r, false)...)
		if diags.HasError() {
			return in, false
		}
		in.Regions = r
	}
	return in, true
}

func functionSchedulerModelFromAPI(ctx context.Context, projectID, functionID string, s *client.FunctionScheduler) (functionSchedulerResourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	m := functionSchedulerResourceModel{
		ID:             types.StringValue(s.ID),
		ProjectID:      types.StringValue(projectID),
		FunctionID:     types.StringValue(functionID),
		Name:           types.StringValue(s.Name),
		Enabled:        types.BoolValue(s.Enabled),
		CronExpression: types.StringValue(s.CronExpression),
		Regions:        stringList(ctx, s.Regions, &diags),
		RunCount:       types.Int64Value(s.RunCount),
		NextRunAt:      stringOrNull(s.NextRunAt),
		CreatedAt:      stringOrNull(s.CreatedAt),
	}
	if len(s.Payload) > 0 {
		b, err := json.Marshal(s.Payload)
		if err != nil {
			diags.AddError("Error encoding scheduler payload", err.Error())
		} else {
			m.Payload = types.StringValue(string(b))
		}
	} else {
		m.Payload = types.StringNull()
	}
	return m, diags
}

func frontendModelFromAPI(ctx context.Context, projectID string, source types.String, f *client.Frontend, diags *diag.Diagnostics) frontendResourceModel {
	variables := stringList(ctx, f.DeclaredVariables, diags)
	return frontendResourceModel{
		ID:              types.StringValue(f.ID),
		ProjectID:       types.StringValue(projectID),
		Name:            types.StringValue(f.Name),
		Source:          source,
		AppRoot:         stringOrNull(f.AppRoot),
		Framework:       stringOrNull(f.Framework),
		VariableScope:   stringOrNull(f.VariableScope),
		Variables:       variables,
		Status:          stringOrNull(f.Status),
		SiteURL:         stringOrNull(f.SiteURL),
		DeployedRegions: stringList(ctx, f.DeployedRegions, diags),
		CreatedAt:       stringOrNull(f.CreatedAt),
		UpdatedAt:       stringOrNull(f.UpdatedAt),
	}
}

func frontendDomainModelFromAPI(ctx context.Context, prior frontendDomainResourceModel, d *client.FrontendDomain, diags *diag.Diagnostics) frontendDomainResourceModel {
	hostname := ""
	if d.RequiredRoutingRecord != nil {
		hostname = d.RequiredRoutingRecord.Value
	}
	if d.RoutingTargetHostname != "" {
		hostname = d.RoutingTargetHostname
	}
	return frontendDomainResourceModel{
		ProjectID:             prior.ProjectID,
		FrontendID:            prior.FrontendID,
		Domain:                types.StringValue(d.Domain),
		CertificatePEM:        prior.CertificatePEM,
		PrivateKeyPEM:         prior.PrivateKeyPEM,
		CertificateChainPEM:   prior.CertificateChainPEM,
		DomainStatus:          types.StringValue(d.DomainStatus),
		VerificationStatus:    types.StringValue(d.VerificationStatus),
		RoutingTargetHostname: stringOrNull(hostname),
		EffectiveURLs:         stringList(ctx, d.EffectiveURLs, diags),
		CreatedAt:             stringOrNull(d.CreatedAt),
		UpdatedAt:             stringOrNull(d.UpdatedAt),
	}
}

func frontendFunctionRouteModelFromAPI(projectID, frontendID string, route *client.FrontendFunctionRoute) frontendFunctionRouteResourceModel {
	return frontendFunctionRouteResourceModel{
		ID:          types.StringValue(route.ID),
		ProjectID:   types.StringValue(projectID),
		FrontendID:  types.StringValue(frontendID),
		FunctionID:  types.StringValue(route.FunctionID),
		PathPrefix:  types.StringValue(route.PathPrefix),
		StripPrefix: types.BoolValue(route.StripPrefix),
		CreatedAt:   stringOrNull(route.CreatedAt),
		UpdatedAt:   stringOrNull(route.UpdatedAt),
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
