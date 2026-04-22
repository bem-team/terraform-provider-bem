// File generated from our OpenAPI spec by Stainless. See CONTRIBUTING.md for details.

package function

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ resource.ResourceWithUpgradeState = (*FunctionResource)(nil)

func (r *FunctionResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{}
}
