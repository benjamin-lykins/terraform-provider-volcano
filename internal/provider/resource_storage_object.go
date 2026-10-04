package provider

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/benjamin-lykins/terraform-provider-volcano/internal/client"
)

var (
	_ resource.Resource                = &storageObjectResource{}
	_ resource.ResourceWithConfigure   = &storageObjectResource{}
	_ resource.ResourceWithImportState = &storageObjectResource{}
)

func NewStorageObjectResource() resource.Resource {
	return &storageObjectResource{}
}

type storageObjectResource struct {
	client *client.Client
}

type storageObjectResourceModel struct {
	ID         types.String `tfsdk:"id"`
	ProjectID  types.String `tfsdk:"project_id"`
	BucketName types.String `tfsdk:"bucket_name"`
	Path       types.String `tfsdk:"path"`
	Source     types.String `tfsdk:"source"`
	IsPublic   types.Bool   `tfsdk:"is_public"`
	MimeType   types.String `tfsdk:"mime_type"`
	Size       types.Int64  `tfsdk:"size"`
	ETag       types.String `tfsdk:"etag"`
	PublicURL  types.String `tfsdk:"public_url"`
	CreatedAt  types.String `tfsdk:"created_at"`
	UpdatedAt  types.String `tfsdk:"updated_at"`
}

func (r *storageObjectResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_storage_object"
}

func (r *storageObjectResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A single file in a storage bucket, uploaded from a local file. Content changes always replace the object (the API's in-place update is a chunked resumable-upload protocol this provider does not drive); is_public can be changed in place. project_id scopes the lookup used by Read (there is no per-object metadata GET, only a project-wide search) but is not itself part of the API path.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"bucket_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"path": schema.StringAttribute{
				Required:    true,
				Description: "Full path within the bucket, e.g. users/abc123/avatar.png.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"source": schema.StringAttribute{
				Required:    true,
				Description: "Path to the local file to upload. Forces replacement if changed (content cannot be updated in place).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"is_public": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "If true, the object can be downloaded with just an anon key.",
			},
			"mime_type": schema.StringAttribute{
				Computed: true,
			},
			"size": schema.Int64Attribute{
				Computed: true,
			},
			"etag": schema.StringAttribute{
				Computed: true,
			},
			"public_url": schema.StringAttribute{
				Computed: true,
			},
			"created_at": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (r *storageObjectResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData))
		return
	}
	r.client = c
}

func (r *storageObjectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan storageObjectResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	content, err := os.ReadFile(plan.Source.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading storage object source file", err.Error())
		return
	}

	obj, err := r.client.UploadStorageObject(ctx, plan.BucketName.ValueString(), plan.Path.ValueString(), filepath.Base(plan.Source.ValueString()), content)
	if err != nil {
		resp.Diagnostics.AddError("Error uploading storage object", err.Error())
		return
	}

	if plan.IsPublic.ValueBool() && !obj.IsPublic {
		obj, err = r.client.SetStorageObjectVisibility(ctx, plan.BucketName.ValueString(), plan.Path.ValueString(), true)
		if err != nil {
			resp.Diagnostics.AddError("Error setting storage object visibility", err.Error())
			return
		}
	}

	model := storageObjectModelFromAPI(plan.ProjectID.ValueString(), plan.BucketName.ValueString(), plan.Path.ValueString(), plan.Source, obj)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *storageObjectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state storageObjectResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	obj, err := r.client.GetStorageObject(ctx, state.ProjectID.ValueString(), state.BucketName.ValueString(), state.Path.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading storage object", err.Error())
		return
	}

	model := storageObjectModelFromAPI(state.ProjectID.ValueString(), state.BucketName.ValueString(), state.Path.ValueString(), state.Source, obj)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *storageObjectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state storageObjectResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	obj, err := r.client.SetStorageObjectVisibility(ctx, state.BucketName.ValueString(), state.Path.ValueString(), plan.IsPublic.ValueBool())
	if err != nil {
		resp.Diagnostics.AddError("Error updating storage object visibility", err.Error())
		return
	}

	model := storageObjectModelFromAPI(state.ProjectID.ValueString(), state.BucketName.ValueString(), state.Path.ValueString(), state.Source, obj)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *storageObjectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state storageObjectResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteStorageObject(ctx, state.BucketName.ValueString(), state.Path.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting storage object", err.Error())
	}
}

func (r *storageObjectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importStateCompositeID(ctx, req, resp, "project_id", "bucket_name", "path")
}
