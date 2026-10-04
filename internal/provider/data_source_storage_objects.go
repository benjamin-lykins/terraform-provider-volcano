package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/benjamin-lykins/terraform-provider-volcano/internal/client"
)

var _ datasource.DataSource = &storageObjectsDataSource{}
var _ datasource.DataSourceWithConfigure = &storageObjectsDataSource{}

func NewStorageObjectsDataSource() datasource.DataSource {
	return &storageObjectsDataSource{}
}

type storageObjectsDataSource struct {
	client *client.Client
}

type storageObjectListItemModel struct {
	ID         types.String `tfsdk:"id"`
	BucketName types.String `tfsdk:"bucket_name"`
	Name       types.String `tfsdk:"name"`
	MimeType   types.String `tfsdk:"mime_type"`
	Size       types.Int64  `tfsdk:"size"`
	IsPublic   types.Bool   `tfsdk:"is_public"`
	PublicURL  types.String `tfsdk:"public_url"`
	OwnerID    types.String `tfsdk:"owner_id"`
	CreatedAt  types.String `tfsdk:"created_at"`
	UpdatedAt  types.String `tfsdk:"updated_at"`
}

type storageObjectsDataSourceModel struct {
	ProjectID types.String                 `tfsdk:"project_id"`
	Search    types.String                 `tfsdk:"search"`
	OwnerID   types.String                 `tfsdk:"owner_id"`
	Objects   []storageObjectListItemModel `tfsdk:"objects"`
}

func (d *storageObjectsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_storage_objects"
}

func (d *storageObjectsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "List storage objects across all of a project's buckets, optionally filtered.",
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{Required: true},
			"search":     schema.StringAttribute{Optional: true, Description: "Case-insensitive substring match on object name."},
			"owner_id":   schema.StringAttribute{Optional: true, Description: "Filter by owning auth user ID."},
			"objects": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":          schema.StringAttribute{Computed: true},
						"bucket_name": schema.StringAttribute{Computed: true},
						"name":        schema.StringAttribute{Computed: true},
						"mime_type":   schema.StringAttribute{Computed: true},
						"size":        schema.Int64Attribute{Computed: true},
						"is_public":   schema.BoolAttribute{Computed: true},
						"public_url":  schema.StringAttribute{Computed: true},
						"owner_id":    schema.StringAttribute{Computed: true},
						"created_at":  schema.StringAttribute{Computed: true},
						"updated_at":  schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *storageObjectsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *storageObjectsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data storageObjectsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	objects, err := d.client.ListStorageObjects(ctx, data.ProjectID.ValueString(), client.ListStorageObjectsOptions{
		Search:  data.Search.ValueString(),
		OwnerID: data.OwnerID.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error listing storage objects", err.Error())
		return
	}

	data.Objects = make([]storageObjectListItemModel, 0, len(objects))
	for _, o := range objects {
		data.Objects = append(data.Objects, storageObjectListItemModel{
			ID:         types.StringValue(o.ID),
			BucketName: stringOrNull(o.BucketName),
			Name:       types.StringValue(o.Name),
			MimeType:   stringOrNull(o.MimeType),
			Size:       types.Int64Value(o.Size),
			IsPublic:   types.BoolValue(o.IsPublic),
			PublicURL:  stringOrNull(o.PublicURL),
			OwnerID:    stringOrNull(o.OwnerID),
			CreatedAt:  stringOrNull(o.CreatedAt),
			UpdatedAt:  stringOrNull(o.UpdatedAt),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
