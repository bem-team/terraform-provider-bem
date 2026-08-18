package function_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// The acceptance test that BEM-1396 finding 2 needed and did not have.
//
// terraform-plugin-testing already asserts, after every apply step, that a
// fresh plan is empty - "After applying this test step, the plan was not
// empty" - because TestStep.ExpectNonEmptyPlan defaults to false. That is
// precisely the assertion the enrich churn violated. It never fired because
// every pre-existing acceptance test used `type = "split"`, so nothing in the
// suite ever created a function with a Float64 attribute.
//
// The bug: the framework quantises a Float64 plan value to float64 while prior
// state carries the API's exact decimal, so `config.steps[].score_threshold`
// read as 0.59999999999999997779... against 0.6 and every plan proposed an
// in-place update, forever. `score_threshold` is the provider's ONLY
// Float64Attribute and there are no NumberAttributes at all, which is why
// enrich alone was affected - every other numeric attribute is Int64, and
// integers round-trip through float64 exactly.
//
// The config below deliberately omits the six Optional+Computed leaves the
// server defaults (steps[].top_k / search_mode / score_threshold,
// endpoints[].match_top_k / max_candidates / max_pages). Setting them was the
// documented workaround, so pinning them here would reintroduce the blind spot
// this test exists to close.
//
// No reachable endpoint is required: creating an enrich function only stores the
// URL, it does not call it. So this needs nothing beyond a staging key.
func TestAccFunctionResource_EnrichReplansClean(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' set")
	}
	testAccPreCheck(t)

	name := acctest.RandomWithPrefix("tf-acc-bem-enrich")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy,
		Steps: []resource.TestStep{
			{
				// Step 1's built-in post-apply plan check is the regression
				// guard. Before the fix this failed with "the plan was not
				// empty" - no extra assertions needed.
				Config: testAccEnrichFunctionConfig(name, "TF Acc Enrich"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bem_function.enrich", "function_name", name),
					resource.TestCheckResourceAttr("bem_function.enrich", "type", "enrich"),
					// The server fills these in; state must hold them, or the
					// next plan sees null and proposes an update.
					resource.TestCheckResourceAttrSet("bem_function.enrich", "config.steps.0.top_k"),
					resource.TestCheckResourceAttrSet("bem_function.enrich", "config.steps.0.search_mode"),
					resource.TestCheckResourceAttrSet("bem_function.enrich", "config.steps.0.score_threshold"),
					resource.TestCheckResourceAttrSet("bem_function.enrich", "config.endpoints.0.match_top_k"),
					resource.TestCheckResourceAttrSet("bem_function.enrich", "config.endpoints.0.max_candidates"),
					resource.TestCheckResourceAttrSet("bem_function.enrich", "config.endpoints.0.max_pages"),
				),
			},
			{
				// Re-applying byte-identical config must be a no-op. Explicit
				// rather than relying on step 1 alone, because the churn only
				// became visible on the *second* plan when it was first found.
				Config: testAccEnrichFunctionConfig(name, "TF Acc Enrich"),
			},
			{
				// A real edit must still work: it has to produce a diff, apply,
				// and then settle. This is the counterpart guard - the plan
				// passes collapse no-op plans, and must not swallow real ones.
				Config: testAccEnrichFunctionConfig(name, "TF Acc Enrich Renamed"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bem_function.enrich", "display_name", "TF Acc Enrich Renamed"),
				),
			},
		},
	})
}

// TestAccFunctionResource_EnrichExplicitScoreThreshold pins the numeric half
// directly: a practitioner-set score_threshold must round-trip and must not be
// reverted by normalizeNumberRepresentation, which adopts prior state's
// representation only when the two agree at float64 precision.
func TestAccFunctionResource_EnrichExplicitScoreThreshold(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' set")
	}
	testAccPreCheck(t)

	name := acctest.RandomWithPrefix("tf-acc-bem-enrich-st")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEnrichFunctionConfigScoreThreshold(name, "0.35"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bem_function.enrich", "config.steps.0.score_threshold", "0.35"),
				),
			},
			{
				// Changing it must produce a real diff and stick - the value is
				// not float64-equal to the previous one, so nothing may
				// normalise it away.
				Config: testAccEnrichFunctionConfigScoreThreshold(name, "0.75"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bem_function.enrich", "config.steps.0.score_threshold", "0.75"),
				),
			},
		},
	})
}

func testAccEnrichFunctionConfig(name, displayName string) string {
	return fmt.Sprintf(`
provider "bem" {}

resource "bem_function" "enrich" {
  function_name = %[1]q
  display_name  = %[2]q
  type          = "enrich"
  tags          = ["bem-acc-test"]

  config = {
    endpoints = [
      {
        name          = "catalog"
        method        = "POST"
        url           = "https://example.invalid/enrich"
        body_template = jsonencode({ query = "{value}", limit = 5 })
        headers       = jsonencode({ "X-Bem-Acc-Test" = "true" })
      }
    ]

    steps = [
      {
        source        = "endpoint"
        endpoint_name = "catalog"
        source_field  = "lineItems[*].description"
        target_field  = "lineItems[*].enrichedProducts"
      }
    ]
  }
}
`, name, displayName)
}

func testAccEnrichFunctionConfigScoreThreshold(name, scoreThreshold string) string {
	return fmt.Sprintf(`
provider "bem" {}

resource "bem_function" "enrich" {
  function_name = %[1]q
  display_name  = "TF Acc Enrich Score Threshold"
  type          = "enrich"
  tags          = ["bem-acc-test"]

  config = {
    endpoints = [
      {
        name          = "catalog"
        method        = "POST"
        url           = "https://example.invalid/enrich"
        body_template = jsonencode({ query = "{value}", limit = 5 })
      }
    ]

    steps = [
      {
        source          = "endpoint"
        endpoint_name   = "catalog"
        source_field    = "lineItems[*].description"
        target_field    = "lineItems[*].enrichedProducts"
        score_threshold = %[2]s
      }
    ]
  }
}
`, name, scoreThreshold)
}
