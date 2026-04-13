// File generated from our OpenAPI spec by Stainless. See CONTRIBUTING.md for details.

package function_test

import (
	"context"
	"testing"

	"github.com/bem-team/terraform-provider-bem/internal/services/function"
	"github.com/bem-team/terraform-provider-bem/internal/test_helpers"
)

func TestFunctionModelSchemaParity(t *testing.T) {
	t.Parallel()
	model := (*function.FunctionModel)(nil)
	schema := function.ResourceSchema(context.TODO())
	errs := test_helpers.ValidateResourceModelSchemaIntegrity(model, schema)
	errs.Report(t)
}
