package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &timeslugProvider{}

type timeslugProvider struct {
	version string
}

type providerModel struct {
	Seed types.String `tfsdk:"seed"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &timeslugProvider{version: version}
	}
}

func (p *timeslugProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "timeslug"
	resp.Version = p.version
}

func (p *timeslugProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Generates deterministic time-rotating slugs from a seed.",
		Attributes: map[string]schema.Attribute{
			"seed": schema.StringAttribute{
				Description: "Secret seed for slug generation.",
				Required:    true,
				Sensitive:   true,
			},
		},
	}
}

func (p *timeslugProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config providerModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.DataSourceData = config.Seed.ValueString()
	resp.ResourceData = config.Seed.ValueString()
}

func (p *timeslugProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{NewSlugsDataSource}
}

func (p *timeslugProvider) Resources(_ context.Context) []func() resource.Resource {
	return nil
}
