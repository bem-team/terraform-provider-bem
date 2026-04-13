// File generated from our OpenAPI spec by Stainless. See CONTRIBUTING.md for details.

package function_test

import (
	"context"
	"testing"

	"github.com/bem-team/terraform-provider-bem/internal/services/function"
	"github.com/bem-team/terraform-provider-bem/internal/test_helpers"
)

func TestFunctionsDataSourceModelSchemaParity(t *testing.T) {
	t.Parallel()
	model := (*function.FunctionsDataSourceModel)(nil)
	schema := function.ListDataSourceSchema(context.TODO())
	errs := test_helpers.ValidateDataSourceModelSchemaIntegrity(model, schema)
	errs.Report(t)
}
