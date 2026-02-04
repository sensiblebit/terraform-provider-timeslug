package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"timeslug": providerserver.NewProtocol6WithError(New("test")()),
}

func TestProvider(t *testing.T) {
	ctx := context.Background()
	p := New("1.0.0")()

	// Metadata
	metaResp := &provider.MetadataResponse{}
	p.Metadata(ctx, provider.MetadataRequest{}, metaResp)
	if metaResp.TypeName != "timeslug" || metaResp.Version != "1.0.0" {
		t.Errorf("got type=%q version=%q", metaResp.TypeName, metaResp.Version)
	}

	// Schema
	schemaResp := &provider.SchemaResponse{}
	p.Schema(ctx, provider.SchemaRequest{}, schemaResp)
	if schemaResp.Diagnostics.HasError() {
		t.Fatal(schemaResp.Diagnostics)
	}
	if _, ok := schemaResp.Schema.Attributes["seed"]; !ok {
		t.Error("missing seed attribute")
	}

	// DataSources
	if ds := p.DataSources(ctx); len(ds) != 1 {
		t.Errorf("expected 1 data source, got %d", len(ds))
	}

	// Resources
	if rs := p.Resources(ctx); len(rs) != 0 {
		t.Errorf("expected 0 resources, got %d", len(rs))
	}
}
