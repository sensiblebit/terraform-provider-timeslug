package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &slugsDataSource{}
	_ datasource.DataSourceWithConfigure = &slugsDataSource{}
)

type slugsDataSource struct {
	seed string
}

type slugsModel struct {
	Anchor   types.String `tfsdk:"anchor"`
	Length   types.Int64  `tfsdk:"length"`
	Window   types.Int64  `tfsdk:"window"`
	Interval types.String `tfsdk:"interval"`
	Mode     types.String `tfsdk:"mode"`
	ID       types.String `tfsdk:"id"`
	Slugs    types.List   `tfsdk:"slugs"`
}

func NewSlugsDataSource() datasource.DataSource {
	return &slugsDataSource{}
}

func (d *slugsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_slugs"
}

func (d *slugsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Generates deterministic slugs for a rolling time window.",
		Attributes: map[string]schema.Attribute{
			"anchor": schema.StringAttribute{
				Description: "Center point for the time window (e.g., 2006-01-02 or 2006-01-02T15:04:05).",
				Required:    true,
			},
			"length": schema.Int64Attribute{
				Description: "Slug length: words (1-24) for bip39, characters for obfuscated. Default: 3",
				Optional:    true,
			},
			"window": schema.Int64Attribute{
				Description: "Number of periods in the window. Default: 7",
				Optional:    true,
			},
			"interval": schema.StringAttribute{
				Description: "Rotation interval: second, minute, hour, day, week. Default: day",
				Optional:    true,
			},
			"mode": schema.StringAttribute{
				Description: "Output mode: bip39 (words) or obfuscated (alphanumeric). Default: bip39",
				Optional:    true,
			},
			"id": schema.StringAttribute{
				Computed: true,
			},
			"slugs": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"slug":   schema.StringAttribute{Computed: true},
						"period": schema.StringAttribute{Computed: true},
						"hash":   schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *slugsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	seed, ok := req.ProviderData.(string)
	if !ok {
		resp.Diagnostics.AddError("Config Error", fmt.Sprintf("expected string, got %T", req.ProviderData))
		return
	}
	d.seed = seed
}

func (d *slugsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data slugsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Defaults
	length := int64(3)
	window := int64(7)
	interval := "day"
	mode := "bip39"

	if !data.Length.IsNull() {
		length = data.Length.ValueInt64()
	}
	if !data.Window.IsNull() {
		window = data.Window.ValueInt64()
	}
	if !data.Interval.IsNull() {
		interval = data.Interval.ValueString()
	}
	if !data.Mode.IsNull() {
		mode = data.Mode.ValueString()
	}

	slugs, err := Generate(d.seed, data.Anchor.ValueString(), int(length), int(window), interval, mode)
	if err != nil {
		resp.Diagnostics.AddError("Generation Failed", err.Error())
		return
	}

	// Build list
	attrTypes := map[string]attr.Type{
		"slug":   types.StringType,
		"period": types.StringType,
		"hash":   types.StringType,
	}

	values := make([]attr.Value, len(slugs))
	for i, s := range slugs {
		values[i], _ = types.ObjectValue(attrTypes, map[string]attr.Value{
			"slug":   types.StringValue(s.Value),
			"period": types.StringValue(s.Period),
			"hash":   types.StringValue(s.Hash),
		})
	}

	list, diags := types.ListValue(types.ObjectType{AttrTypes: attrTypes}, values)
	resp.Diagnostics.Append(diags...)

	data.ID = types.StringValue(fmt.Sprintf("%s-%s-%s-%d-%d", data.Anchor.ValueString(), mode, interval, length, window))
	data.Slugs = list

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
