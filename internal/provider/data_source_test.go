package provider

import (
	"context"
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestSlugsDataSource(t *testing.T) {
	ctx := context.Background()
	ds := NewSlugsDataSource()

	// Metadata
	metaResp := &datasource.MetadataResponse{}
	ds.Metadata(ctx, datasource.MetadataRequest{ProviderTypeName: "timeslug"}, metaResp)
	if metaResp.TypeName != "timeslug_slugs" {
		t.Errorf("got type=%q", metaResp.TypeName)
	}

	// Schema
	schemaResp := &datasource.SchemaResponse{}
	ds.Schema(ctx, datasource.SchemaRequest{}, schemaResp)
	if schemaResp.Diagnostics.HasError() {
		t.Fatal(schemaResp.Diagnostics)
	}
	required := []string{"anchor"}
	optional := []string{"length", "window", "interval", "mode"}
	computed := []string{"id", "slugs"}
	for _, attr := range slices.Concat(required, optional, computed) {
		if _, ok := schemaResp.Schema.Attributes[attr]; !ok {
			t.Errorf("missing attribute %q", attr)
		}
	}

	// Configure
	concrete := ds.(*slugsDataSource)
	configResp := &datasource.ConfigureResponse{}
	concrete.Configure(ctx, datasource.ConfigureRequest{ProviderData: "test-seed"}, configResp)
	if configResp.Diagnostics.HasError() {
		t.Fatal(configResp.Diagnostics)
	}
	concrete.Configure(ctx, datasource.ConfigureRequest{ProviderData: 123}, configResp)
	if !configResp.Diagnostics.HasError() {
		t.Error("expected error for wrong type")
	}
}

// Acceptance tests
func TestAccSlugsDataSource_bip39(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: `
provider "timeslug" { seed = "seedphrase" }
data "timeslug_slugs" "test" {
  anchor = "2026-02-03"
  length = 3
  window = 3
  mode   = "bip39"
}`,
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("data.timeslug_slugs.test", "slugs.#", "3"),
				resource.TestCheckResourceAttr("data.timeslug_slugs.test", "slugs.1.slug", "exoticangryanswer"),
				resource.TestCheckResourceAttr("data.timeslug_slugs.test", "slugs.1.hash", "50011c26d0"),
			),
		}},
	})
}

func TestAccSlugsDataSource_obfuscated(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: `
provider "timeslug" { seed = "seedphrase" }
data "timeslug_slugs" "test" {
  anchor = "2026-02-03"
  length = 16
  window = 3
  mode   = "obfuscated"
}`,
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("data.timeslug_slugs.test", "slugs.#", "3"),
				resource.TestCheckResourceAttr("data.timeslug_slugs.test", "slugs.1.slug", "trybeambold8"),
				resource.TestCheckResourceAttr("data.timeslug_slugs.test", "slugs.1.hash", "5d3bf0d55db67ea2"),
			),
		}},
	})
}

func TestAccSlugsDataSource_defaults(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: `
provider "timeslug" { seed = "test" }
data "timeslug_slugs" "test" { anchor = "2026-01-15" }`,
			Check: resource.TestCheckResourceAttr("data.timeslug_slugs.test", "slugs.#", "7"),
		}},
	})
}
