package internal_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/bem-team/bem-go-sdk"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/bem-team/terraform-provider-bem/internal"
)

// These tests mutate process environment variables (BEM_API_KEY /
// BEM_BASE_URL), so nothing in this file calls t.Parallel().

func TestProviderMetadata(t *testing.T) {
	p := internal.NewProvider("1.2.3")()

	resp := &provider.MetadataResponse{}
	p.Metadata(context.TODO(), provider.MetadataRequest{}, resp)

	if resp.TypeName != "bem" {
		t.Errorf("TypeName = %q, want %q", resp.TypeName, "bem")
	}
	if resp.Version != "1.2.3" {
		t.Errorf("Version = %q, want %q", resp.Version, "1.2.3")
	}
}

func TestProviderSchema(t *testing.T) {
	p := internal.NewProvider("test")()

	resp := &provider.SchemaResponse{}
	p.Schema(context.TODO(), provider.SchemaRequest{}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema returned diagnostics: %v", resp.Diagnostics)
	}

	for _, name := range []string{"base_url", "api_key"} {
		attribute, ok := resp.Schema.Attributes[name]
		if !ok {
			t.Errorf("provider schema is missing the %q attribute", name)
			continue
		}
		if !attribute.IsOptional() {
			t.Errorf("attribute %q should be Optional", name)
		}
		if attribute.IsRequired() {
			t.Errorf("attribute %q should not be Required — it falls back to an environment variable", name)
		}
	}
}

// If codegen adds a service, this fails and prompts adding coverage for it
// rather than letting a new resource ship with no tests at all.
func TestProviderRegistersExpectedResourcesAndDataSources(t *testing.T) {
	ctx := context.TODO()
	p := internal.NewProvider("test")()

	gotResources := map[string]bool{}
	for _, newResource := range p.Resources(ctx) {
		r := newResource()
		if r == nil {
			t.Fatal("a resource constructor returned nil")
		}
		resp := &resource.MetadataResponse{}
		r.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "bem"}, resp)
		gotResources[resp.TypeName] = true
	}

	for _, want := range []string{"bem_function", "bem_workflow"} {
		if !gotResources[want] {
			t.Errorf("resource %q is not registered", want)
		}
	}
	if len(gotResources) != 2 {
		t.Errorf("registered resources = %v; a new resource needs its own test coverage", gotResources)
	}

	gotDataSources := map[string]bool{}
	for _, newDataSource := range p.DataSources(ctx) {
		d := newDataSource()
		if d == nil {
			t.Fatal("a data source constructor returned nil")
		}
		resp := &datasource.MetadataResponse{}
		d.Metadata(ctx, datasource.MetadataRequest{ProviderTypeName: "bem"}, resp)
		gotDataSources[resp.TypeName] = true
	}

	for _, want := range []string{"bem_function", "bem_functions", "bem_workflow", "bem_workflows"} {
		if !gotDataSources[want] {
			t.Errorf("data source %q is not registered", want)
		}
	}
	if len(gotDataSources) != 4 {
		t.Errorf("registered data sources = %v; a new data source needs its own test coverage", gotDataSources)
	}
}

func TestProviderConfigValidators(t *testing.T) {
	p, ok := internal.NewProvider("test")().(provider.ProviderWithConfigValidators)
	if !ok {
		t.Fatal("provider does not implement ProviderWithConfigValidators")
	}
	if got := p.ConfigValidators(context.TODO()); len(got) != 0 {
		t.Errorf("ConfigValidators() = %v, want none", got)
	}
}

func TestProviderConfigureMissingAPIKeyIsAnError(t *testing.T) {
	unsetEnv(t, "BEM_API_KEY")
	unsetEnv(t, "BEM_BASE_URL")

	resp := configureProvider(t, nil, nil)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected an error when no api_key is configured and BEM_API_KEY is unset")
	}
	if resp.ResourceData != nil || resp.DataSourceData != nil {
		t.Error("no client should be handed to resources when configuration failed")
	}

	var found bool
	for _, d := range resp.Diagnostics.Errors() {
		if d.Summary() == "Missing api_key value" {
			found = true
		}
	}
	if !found {
		t.Errorf("diagnostics = %v, want a \"Missing api_key value\" error", resp.Diagnostics)
	}
}

func TestProviderConfigureAPIKeyFromEnvironment(t *testing.T) {
	t.Setenv("BEM_API_KEY", "env-key")
	unsetEnv(t, "BEM_BASE_URL")

	resp := configureProvider(t, nil, nil)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Configure returned diagnostics: %v", resp.Diagnostics)
	}
	if _, ok := resp.ResourceData.(*bem.Client); !ok {
		t.Errorf("ResourceData = %T, want *bem.Client", resp.ResourceData)
	}
	if _, ok := resp.DataSourceData.(*bem.Client); !ok {
		t.Errorf("DataSourceData = %T, want *bem.Client", resp.DataSourceData)
	}
}

// Asserts the precedence rule behaviourally — by seeing which host the
// configured client actually calls and which key it presents — rather than
// just checking that Configure returned without error. The client's options
// are unexported, so a real round trip is the only way to observe this.
func TestProviderConfigurePrefersConfigOverEnvironment(t *testing.T) {
	envServer := recordingServer(t)
	configServer := recordingServer(t)

	t.Setenv("BEM_BASE_URL", envServer.url)
	t.Setenv("BEM_API_KEY", "env-key")

	configuredURL := configServer.url
	configuredKey := "config-key"
	resp := configureProvider(t, &configuredURL, &configuredKey)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Configure returned diagnostics: %v", resp.Diagnostics)
	}

	client, ok := resp.ResourceData.(*bem.Client)
	if !ok {
		t.Fatalf("ResourceData = %T, want *bem.Client", resp.ResourceData)
	}
	if _, err := client.Functions.Get(context.TODO(), "some-function", bem.FunctionGetParams{}); err != nil {
		t.Fatalf("calling the configured client: %v", err)
	}

	if envServer.hits() != 0 {
		t.Errorf("the BEM_BASE_URL host was called %d times; config base_url should win", envServer.hits())
	}
	if configServer.hits() != 1 {
		t.Fatalf("the config base_url host was called %d times, want 1", configServer.hits())
	}
	if got := configServer.lastAPIKey(); got != "config-key" {
		t.Errorf("x-api-key = %q, want %q — config api_key should win over BEM_API_KEY", got, "config-key")
	}
}

func TestProviderConfigureFallsBackToEnvironmentBaseURL(t *testing.T) {
	envServer := recordingServer(t)

	t.Setenv("BEM_BASE_URL", envServer.url)
	t.Setenv("BEM_API_KEY", "env-key")

	resp := configureProvider(t, nil, nil)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Configure returned diagnostics: %v", resp.Diagnostics)
	}

	client, ok := resp.ResourceData.(*bem.Client)
	if !ok {
		t.Fatalf("ResourceData = %T, want *bem.Client", resp.ResourceData)
	}
	if _, err := client.Functions.Get(context.TODO(), "some-function", bem.FunctionGetParams{}); err != nil {
		t.Fatalf("calling the configured client: %v", err)
	}

	if envServer.hits() != 1 {
		t.Errorf("the BEM_BASE_URL host was called %d times, want 1", envServer.hits())
	}
	if got := envServer.lastAPIKey(); got != "env-key" {
		t.Errorf("x-api-key = %q, want %q", got, "env-key")
	}
}

// configureProvider runs Configure with the given base_url / api_key, where a
// nil pointer means the attribute is null in config.
func configureProvider(t *testing.T, baseURL, apiKey *string) *provider.ConfigureResponse {
	t.Helper()

	ctx := context.TODO()
	schema := internal.ProviderSchema(ctx)

	objectType, ok := schema.Type().TerraformType(ctx).(tftypes.Object)
	if !ok {
		t.Fatalf("provider schema type = %T, want tftypes.Object", schema.Type().TerraformType(ctx))
	}

	raw := tftypes.NewValue(objectType, map[string]tftypes.Value{
		"base_url": tftypes.NewValue(tftypes.String, baseURL),
		"api_key":  tftypes.NewValue(tftypes.String, apiKey),
	})

	resp := &provider.ConfigureResponse{}
	internal.NewProvider("test")().Configure(ctx, provider.ConfigureRequest{
		Config: tfsdk.Config{Raw: raw, Schema: schema},
	}, resp)

	return resp
}

type stubServer struct {
	url      string
	requests chan *http.Request
	seen     []*http.Request
}

func (s *stubServer) drain() {
	for {
		select {
		case req := <-s.requests:
			s.seen = append(s.seen, req)
		default:
			return
		}
	}
}

func (s *stubServer) hits() int {
	s.drain()
	return len(s.seen)
}

func (s *stubServer) lastAPIKey() string {
	s.drain()
	if len(s.seen) == 0 {
		return ""
	}
	return s.seen[len(s.seen)-1].Header.Get("x-api-key")
}

// recordingServer stands in for the bem API. It returns a valid, minimal
// function payload so the SDK neither retries nor errors.
func recordingServer(t *testing.T) *stubServer {
	t.Helper()

	s := &stubServer{requests: make(chan *http.Request, 16)}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.requests <- r.Clone(context.Background())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"name":"some-function","displayName":"Some Function"}`))
	}))
	t.Cleanup(srv.Close)

	s.url = srv.URL
	return s
}

// unsetEnv removes a variable for the duration of the test and restores
// whatever was there before. t.Setenv cannot express "absent" — the provider
// uses os.LookupEnv, for which an empty string is still present.
func unsetEnv(t *testing.T, key string) {
	t.Helper()

	prev, had := os.LookupEnv(key)
	t.Cleanup(func() {
		if had {
			_ = os.Setenv(key, prev)
			return
		}
		_ = os.Unsetenv(key)
	})
	_ = os.Unsetenv(key)
}
