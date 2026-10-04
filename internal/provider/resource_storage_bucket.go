package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/benjamin-lykins/terraform-provider-volcano/internal/client"
)

var (
	_ resource.Resource                = &storageBucketResource{}
	_ resource.ResourceWithConfigure   = &storageBucketResource{}
	_ resource.ResourceWithImportState = &storageBucketResource{}
)

func NewStorageBucketResource() resource.Resource {
	return &storageBucketResource{}
}

type storageBucketResource struct {
	client *client.Client
}

type storageBucketResourceModel struct {
	ID               types.String `tfsdk:"id"`
	ProjectID        types.String `tfsdk:"project_id"`
	Name             types.String `tfsdk:"name"`
	FileSizeLimit    types.Int64  `tfsdk:"file_size_limit"`
	AllowedMimeTypes types.List   `tfsdk:"allowed_mime_types"`
	CreatedAt        types.String `tfsdk:"created_at"`
	UpdatedAt        types.String `tfsdk:"updated_at"`
}

func (r *storageBucketResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_storage_bucket"
}

func (r *storageBucketResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A named container for files within a project. Per-file public access is controlled separately via volcano_storage_object's is_public.",
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
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"file_size_limit": schema.Int64Attribute{
				Optional:    true,
				Description: "Maximum file size in bytes. Omit for no limit.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"allowed_mime_types": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Allowed MIME types. Omit to allow all types.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
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

func (r *storageBucketResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *storageBucketResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan storageBucketResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := client.CreateStorageBucketRequest{Name: plan.Name.ValueString()}
	if !plan.FileSizeLimit.IsNull() {
		v := plan.FileSizeLimit.ValueInt64()
		in.FileSizeLimit = &v
	}
	if !plan.AllowedMimeTypes.IsNull() && !plan.AllowedMimeTypes.IsUnknown() {
		var types_ []string
		resp.Diagnostics.Append(plan.AllowedMimeTypes.ElementsAs(ctx, &types_, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		in.AllowedMimeTypes = types_
	}

	bucket, err := r.client.CreateStorageBucket(ctx, plan.ProjectID.ValueString(), in)
	if err != nil {
		resp.Diagnostics.AddError("Error creating storage bucket", err.Error())
		return
	}

	model := storageBucketModelFromAPI(ctx, plan.ProjectID.ValueString(), bucket, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *storageBucketResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state storageBucketResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	bucket, err := r.client.GetStorageBucket(ctx, state.ProjectID.ValueString(), state.Name.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading storage bucket", err.Error())
		return
	}

	model := storageBucketModelFromAPI(ctx, state.ProjectID.ValueString(), bucket, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *storageBucketResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state storageBucketResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := client.UpdateStorageBucketRequest{}
	if !plan.FileSizeLimit.Equal(state.FileSizeLimit) {
		v := plan.FileSizeLimit.ValueInt64()
		in.FileSizeLimit = &v
	}
	if !plan.AllowedMimeTypes.Equal(state.AllowedMimeTypes) {
		var types_ []string
		resp.Diagnostics.Append(plan.AllowedMimeTypes.ElementsAs(ctx, &types_, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		in.AllowedMimeTypes = types_
	}

	bucket, err := r.client.UpdateStorageBucket(ctx, state.ProjectID.ValueString(), state.Name.ValueString(), in)
	if err != nil {
		resp.Diagnostics.AddError("Error updating storage bucket", err.Error())
		return
	}

	model := storageBucketModelFromAPI(ctx, state.ProjectID.ValueString(), bucket, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *storageBucketResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state storageBucketResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteStorageBucket(ctx, state.ProjectID.ValueString(), state.Name.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting storage bucket", err.Error())
	}
}

func (r *storageBucketResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importStateCompositeID(ctx, req, resp, "project_id", "name")
}
