package main

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// 0.1.16 regression — SHODAN_API_KEY env var was documented but not
// implemented. This test pins the contract: when api_key is unset in
// config, the env var is used; when both are empty, an error is raised.
func TestProviderConfigure_EnvVarFallback(t *testing.T) {
	t.Setenv("SHODAN_API_KEY", "env-key-value")

	resp := runConfigure(t, types.StringNull(), types.Int64Null())
	if resp.Diagnostics.HasError() {
		t.Fatalf("env var should provide api_key, got errors: %v", resp.Diagnostics.Errors())
	}
	if resp.ResourceData == nil {
		t.Fatal("ResourceData not populated — client not constructed")
	}
}

func TestProviderConfigure_ConfigBeatsEnvVar(t *testing.T) {
	t.Setenv("SHODAN_API_KEY", "env-key-value")

	resp := runConfigure(t, types.StringValue("config-key-value"), types.Int64Null())
	if resp.Diagnostics.HasError() {
		t.Fatalf("explicit config should win, got errors: %v", resp.Diagnostics.Errors())
	}
}

func TestProviderConfigure_MissingApiKeyErrors(t *testing.T) {
	t.Setenv("SHODAN_API_KEY", "")

	resp := runConfigure(t, types.StringNull(), types.Int64Null())
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected error when both config and env are empty, got no diagnostics")
	}
}

// runConfigure builds a provider.ConfigureRequest with the supplied
// api_key / request_interval values and runs the provider's Configure
// against it. Returns the response for assertion.
func runConfigure(t *testing.T, apiKey types.String, interval types.Int64) provider.ConfigureResponse {
	t.Helper()
	p := &ShodanProvider{version: "test"}

	schemaResp := provider.SchemaResponse{}
	p.Schema(context.Background(), provider.SchemaRequest{}, &schemaResp)

	tfValue := tftypes.NewValue(
		tftypes.Object{AttributeTypes: map[string]tftypes.Type{
			"api_key":          tftypes.String,
			"request_interval": tftypes.Number,
		}},
		map[string]tftypes.Value{
			"api_key":          stringToTfValue(apiKey),
			"request_interval": int64ToTfValue(interval),
		},
	)

	req := provider.ConfigureRequest{
		Config: tfsdk.Config{
			Schema: schemaResp.Schema,
			Raw:    tfValue,
		},
	}
	resp := provider.ConfigureResponse{}
	p.Configure(context.Background(), req, &resp)
	return resp
}

func stringToTfValue(v types.String) tftypes.Value {
	if v.IsNull() {
		return tftypes.NewValue(tftypes.String, nil)
	}
	return tftypes.NewValue(tftypes.String, v.ValueString())
}

func int64ToTfValue(v types.Int64) tftypes.Value {
	if v.IsNull() {
		return tftypes.NewValue(tftypes.Number, nil)
	}
	return tftypes.NewValue(tftypes.Number, v.ValueInt64())
}
