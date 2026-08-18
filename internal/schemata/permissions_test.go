package schemata_test

import (
	"testing"

	"github.com/bem-team/terraform-provider-bem/internal/schemata"
)

func TestDescriptionString(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		description schemata.Description
		expected    string
	}{
		"empty": {
			description: schemata.Description{},
			expected:    "",
		},
		"markdown only": {
			description: schemata.Description{MarkdownDescription: "The function name."},
			expected:    "The function name.",
		},
		"scopes only": {
			description: schemata.Description{Scopes: []string{"functions:read"}},
			expected:    "Accepted Permissions\n\n- `functions:read`\n",
		},
		"multiple scopes": {
			description: schemata.Description{Scopes: []string{"functions:read", "functions:write"}},
			expected:    "Accepted Permissions\n\n- `functions:read`\n- `functions:write`\n",
		},
		"scopes and markdown are separated by a blank line": {
			description: schemata.Description{
				Scopes:              []string{"workflows:read"},
				MarkdownDescription: "The workflow name.",
			},
			expected: "Accepted Permissions\n\n- `workflows:read`\n\nThe workflow name.",
		},
		"backticks in a scope are escaped so they cannot break out of the code span": {
			description: schemata.Description{Scopes: []string{"weird`scope"}},
			expected:    "Accepted Permissions\n\n- `weird\\`scope`\n",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if got := tc.description.String(); got != tc.expected {
				t.Errorf("String() = %q, want %q", got, tc.expected)
			}
		})
	}
}
