package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/benjamin-lykins/terraform-provider-volcano/internal/client"
)

var pathRootID = path.Root("id")
var pathProjectID = path.Root("project_id")
var hexColorRegexp = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

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

func storageBucketModelFromAPI(ctx context.Context, projectID string, b *client.StorageBucket, diags *diag.Diagnostics) storageBucketResourceModel {
	m := storageBucketResourceModel{
		ID:        types.StringValue(b.ID),
		ProjectID: types.StringValue(projectID),
		Name:      types.StringValue(b.Name),
		CreatedAt: stringOrNull(b.CreatedAt),
		UpdatedAt: stringOrNull(b.UpdatedAt),
	}
	if b.FileSizeLimit != nil {
		m.FileSizeLimit = types.Int64Value(*b.FileSizeLimit)
	} else {
		m.FileSizeLimit = types.Int64Null()
	}
	if b.AllowedMimeTypes != nil {
		m.AllowedMimeTypes = stringList(ctx, b.AllowedMimeTypes, diags)
	} else {
		m.AllowedMimeTypes = types.ListNull(types.StringType)
	}
	return m
}

func storageBucketPolicyModelFromAPI(projectID, bucketName string, p *client.StorageBucketPolicy) storageBucketPolicyResourceModel {
	return storageBucketPolicyResourceModel{
		ID:         types.StringValue(p.ID),
		ProjectID:  types.StringValue(projectID),
		BucketName: types.StringValue(bucketName),
		Name:       types.StringValue(p.Name),
		Operation:  types.StringValue(p.Operation),
		Definition: types.StringValue(p.Definition),
		CreatedAt:  stringOrNull(p.CreatedAt),
	}
}

func storageObjectModelFromAPI(projectID, bucketName, path string, source types.String, o *client.StorageObject) storageObjectResourceModel {
	return storageObjectResourceModel{
		ID:         types.StringValue(o.ID),
		ProjectID:  types.StringValue(projectID),
		BucketName: types.StringValue(bucketName),
		Path:       types.StringValue(path),
		Source:     source,
		IsPublic:   types.BoolValue(o.IsPublic),
		MimeType:   stringOrNull(o.MimeType),
		Size:       types.Int64Value(o.Size),
		ETag:       stringOrNull(o.ETag),
		PublicURL:  stringOrNull(o.PublicURL),
		CreatedAt:  stringOrNull(o.CreatedAt),
		UpdatedAt:  stringOrNull(o.UpdatedAt),
	}
}

func projectAuthMethodsModelFromAPI(ctx context.Context, projectID string, m *client.ProjectAuthMethods, diags *diag.Diagnostics) projectAuthMethodsResourceModel {
	providers := make([]oauthProviderToggleModel, 0, len(m.OAuthProviders))
	for _, p := range m.OAuthProviders {
		providers = append(providers, oauthProviderToggleModel{
			Provider: types.StringValue(p.Provider),
			Enabled:  types.BoolValue(p.Enabled),
		})
	}
	return projectAuthMethodsResourceModel{
		ProjectID:           types.StringValue(projectID),
		EnableAnonymous:     types.BoolValue(m.Anonymous.Enabled),
		EnableEmailPassword: types.BoolValue(m.EmailPassword.Enabled),
		OAuthProviders:      providers,
		AvailableMethods:    stringList(ctx, m.AvailableMethods, diags),
	}
}

func projectAuthThemeModelFromAPI(projectID string, t *client.Theme) projectAuthThemeResourceModel {
	return projectAuthThemeResourceModel{
		ProjectID: types.StringValue(projectID),
		Colors: themeColorsModel{
			Background: types.StringValue(t.Colors.Background),
			Surface:    types.StringValue(t.Colors.Surface),
			Text:       types.StringValue(t.Colors.Text),
			Accent:     types.StringValue(t.Colors.Accent),
			AccentText: types.StringValue(t.Colors.AccentText),
		},
		Font:    stringOrNull(t.Font),
		Scale:   stringOrNull(t.Scale),
		Density: stringOrNull(t.Density),
		Radius:  stringOrNull(t.Radius),
	}
}

func projectAuthHostedPageModelFromAPI(projectID, pageType string, p *client.AuthHostedPage) projectAuthHostedPageResourceModel {
	return projectAuthHostedPageResourceModel{
		ProjectID: types.StringValue(projectID),
		PageType:  types.StringValue(pageType),
		HTML:      types.StringValue(p.HTML),
		CSS:       stringOrNull(p.CSS),
	}
}

func emailTemplateModelFromAPI(projectID string, t *client.EmailTemplate) emailTemplateResourceModel {
	return emailTemplateResourceModel{
		ID:           types.StringValue(t.ID),
		ProjectID:    types.StringValue(projectID),
		TemplateType: types.StringValue(t.TemplateType),
		Subject:      types.StringValue(t.Subject),
		HTMLBody:     types.StringValue(t.HTMLBody),
		TextBody:     types.StringValue(t.TextBody),
		CreatedAt:    stringOrNull(t.CreatedAt),
		UpdatedAt:    stringOrNull(t.UpdatedAt),
	}
}

func oauthConfigModelFromAPI(ctx context.Context, projectID string, c *client.OAuthConfig, diags *diag.Diagnostics) oauthConfigResourceModel {
	return oauthConfigResourceModel{
		ID:           types.StringValue(c.ID),
		ProjectID:    types.StringValue(projectID),
		Provider:     types.StringValue(c.Provider),
		ClientID:     stringOrNull(c.ClientID),
		ClientSecret: stringOrNull(c.ClientSecret),
		RedirectURL:  stringOrNull(c.RedirectURL),
		Scopes:       stringList(ctx, c.Scopes, diags),
		Enabled:      types.BoolValue(c.Enabled),
		CreatedAt:    stringOrNull(c.CreatedAt),
		UpdatedAt:    stringOrNull(c.UpdatedAt),
	}
}

func realtimeConfigModelFromAPI(projectID string, c *client.RealtimeConfig) realtimeConfigResourceModel {
	return realtimeConfigResourceModel{
		ProjectID:              types.StringValue(projectID),
		Enabled:                types.BoolValue(c.Enabled),
		BroadcastEnabled:       types.BoolValue(c.BroadcastEnabled),
		PresenceEnabled:        types.BoolValue(c.PresenceEnabled),
		PostgresChangesEnabled: types.BoolValue(c.PostgresChangesEnabled),
	}
}

func sandboxModelFromAPI(projectID string, s *client.Sandbox) sandboxResourceModel {
	return sandboxResourceModel{
		ID:        types.StringValue(s.ID),
		ProjectID: types.StringValue(projectID),
		Name:      types.StringValue(s.Name),
		Preset:    types.StringValue(s.Preset),
		MemoryMB:  types.Int64Value(s.MemoryMB),
		Status:    stringOrNull(s.Status),
		CreatedAt: stringOrNull(s.CreatedAt),
	}
}

func projectGitConnectionModelFromAPI(projectID string, connectionID types.String, g *client.ProjectGitConnection) projectGitConnectionResourceModel {
	return projectGitConnectionResourceModel{
		ProjectID:        types.StringValue(projectID),
		ConnectionID:     connectionID,
		InstallationID:   types.Int64Value(g.RepoInstallationID),
		RepositoryID:     types.Int64Value(g.RepoID),
		RepoFullName:     stringOrNull(g.RepoFullName),
		RootDirectory:    stringOrNull(g.RootDirectory),
		ProductionBranch: stringOrNull(g.ProductionBranch),
		UpdatedAt:        stringOrNull(g.UpdatedAt),
	}
}

func projectGitDeploySettingsModelFromAPI(projectID string, s *client.ProjectGitDeploySettings) projectGitDeploySettingsResourceModel {
	return projectGitDeploySettingsResourceModel{
		ProjectID:         types.StringValue(projectID),
		AutoDeployEnabled: types.BoolValue(s.AutoDeployEnabled),
		DeployFunctions:   types.BoolValue(s.DeployFunctions),
		FrontendName:      stringOrNull(s.FrontendName),
		FrontendAppRoot:   stringOrNull(s.FrontendAppRoot),
		UpdatedAt:         stringOrNull(s.UpdatedAt),
	}
}

func projectSourceExportModelFromAPI(ctx context.Context, projectID string, productionBranch types.String, result *client.ProjectSourceExportResult, status *client.ProjectSourceExportStatus, diags *diag.Diagnostics) projectSourceExportResourceModel {
	m := projectSourceExportResourceModel{
		ProjectID:           types.StringValue(projectID),
		ProductionBranch:    productionBranch,
		Mode:                stringOrNull(status.Mode),
		ExportedAt:          stringOrNull(status.ExportedAt),
		TransitionStartedAt: stringOrNull(status.TransitionStartedAt),
		HandedOverAt:        stringOrNull(status.HandedOverAt),
	}
	if result != nil {
		m.RepoFullName = stringOrNull(result.RepoFullName)
		m.Branch = stringOrNull(result.Branch)
		m.CommitSHA = stringOrNull(result.CommitSHA)
		m.FileCount = types.Int64Value(result.FileCount)
		m.Skipped = stringList(ctx, result.Skipped, diags)
		m.Omitted = stringList(ctx, result.Omitted, diags)
	}
	return m
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
