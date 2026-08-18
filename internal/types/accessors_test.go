package types_test

import (
	"math/big"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	fwtypes "github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	bemtypes "github.com/bem-team/terraform-provider-bem/internal/types"
)

// These accessors back customfield's semantic-equality comparison and
// customvalidator's dynamic validators, which is how `output_schema` (a
// dynamic attribute) avoids spurious diffs. A false negative here shows up as
// a perpetual plan diff rather than a test failure, so the "not ok" cases
// matter as much as the "ok" ones.

func TestIntValue(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		value    attr.Value
		wantOK   bool
		expected int64
	}{
		"nil":                         {value: nil, wantOK: false},
		"int32":                       {value: basetypes.NewInt32Value(5), wantOK: true, expected: 5},
		"int64":                       {value: basetypes.NewInt64Value(-7), wantOK: true, expected: -7},
		"number holding an integer":   {value: basetypes.NewNumberValue(big.NewFloat(42)), wantOK: true, expected: 42},
		"number holding a fraction":   {value: basetypes.NewNumberValue(big.NewFloat(1.5)), wantOK: false},
		"null number":                 {value: basetypes.NewNumberNull(), wantOK: false},
		"unknown number":              {value: basetypes.NewNumberUnknown(), wantOK: false},
		"float64 with no fraction":    {value: basetypes.NewFloat64Value(3), wantOK: true, expected: 3},
		"float64 with a fraction":     {value: basetypes.NewFloat64Value(3.5), wantOK: false},
		"float32 with no fraction":    {value: basetypes.NewFloat32Value(2), wantOK: true, expected: 2},
		"string is not a number":      {value: basetypes.NewStringValue("4"), wantOK: false},
		"bool is not a number":        {value: basetypes.NewBoolValue(true), wantOK: false},
		"negative float64, no frac":   {value: basetypes.NewFloat64Value(-11), wantOK: true, expected: -11},
		"number larger than an int64": {value: basetypes.NewNumberValue(new(big.Float).SetInt(new(big.Int).Lsh(big.NewInt(1), 70))), wantOK: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ok, got := bemtypes.IntValue(tc.value)
			if ok != tc.wantOK {
				t.Fatalf("IntValue() ok = %v, want %v (value %v)", ok, tc.wantOK, got)
			}
			if !ok {
				if got != nil {
					t.Errorf("IntValue() returned %v alongside ok=false, want nil", got)
				}
				return
			}
			if tc.expected != 0 && got.Cmp(big.NewInt(tc.expected)) != 0 {
				t.Errorf("IntValue() = %v, want %d", got, tc.expected)
			}
		})
	}
}

// A 2^70 number exceeds int64 but is still an exact integer; it must survive
// as a *big.Int rather than silently overflowing.
func TestIntValueBeyondInt64(t *testing.T) {
	t.Parallel()

	want := new(big.Int).Lsh(big.NewInt(1), 70)
	ok, got := bemtypes.IntValue(basetypes.NewNumberValue(new(big.Float).SetInt(want)))
	if !ok {
		t.Fatal("IntValue() ok = false, want true")
	}
	if got.Cmp(want) != 0 {
		t.Errorf("IntValue() = %v, want %v", got, want)
	}
}

func TestFloatValue(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		value    attr.Value
		wantOK   bool
		expected float64
	}{
		"nil":                        {value: nil, wantOK: false},
		"float64":                    {value: basetypes.NewFloat64Value(1.25), wantOK: true, expected: 1.25},
		"float32":                    {value: basetypes.NewFloat32Value(0.5), wantOK: true, expected: 0.5},
		"number":                     {value: basetypes.NewNumberValue(big.NewFloat(2.75)), wantOK: true, expected: 2.75},
		"null number":                {value: basetypes.NewNumberNull(), wantOK: false},
		"unknown number":             {value: basetypes.NewNumberUnknown(), wantOK: false},
		"int64 widens to a float":    {value: basetypes.NewInt64Value(9), wantOK: true, expected: 9},
		"int32 widens to a float":    {value: basetypes.NewInt32Value(-3), wantOK: true, expected: -3},
		"string is not a number":     {value: basetypes.NewStringValue("1.5"), wantOK: false},
		"list is not a number":       {value: basetypes.NewListNull(fwtypes.StringType), wantOK: false},
		"dynamic is not a raw float": {value: basetypes.NewDynamicNull(), wantOK: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ok, got := bemtypes.FloatValue(tc.value)
			if ok != tc.wantOK {
				t.Fatalf("FloatValue() ok = %v, want %v (value %v)", ok, tc.wantOK, got)
			}
			if !ok {
				if got != nil {
					t.Errorf("FloatValue() returned %v alongside ok=false, want nil", got)
				}
				return
			}
			if got.Cmp(big.NewFloat(tc.expected)) != 0 {
				t.Errorf("FloatValue() = %v, want %v", got, tc.expected)
			}
		})
	}
}

// Documents a real asymmetry rather than endorsing it: NumberValue guards
// null/unknown explicitly, the fixed-width numeric types do not, so a null
// Float64/Int64 reports ok=true with the zero value. Anything comparing two
// dynamic values therefore treats "null" and "0" as equal for those types. If
// that guard is ever added, this test should be updated, not deleted.
func TestNumericAccessorsTreatNullFixedWidthValuesAsZero(t *testing.T) {
	t.Parallel()

	for name, value := range map[string]attr.Value{
		"float64": basetypes.NewFloat64Null(),
		"float32": basetypes.NewFloat32Null(),
		"int64":   basetypes.NewInt64Null(),
		"int32":   basetypes.NewInt32Null(),
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ok, got := bemtypes.FloatValue(value)
			if !ok || got.Sign() != 0 {
				t.Errorf("FloatValue(null %s) = (%v, %v), want (true, 0) per current behaviour", name, ok, got)
			}
		})
	}
}

func TestChildItems(t *testing.T) {
	t.Parallel()

	elems := []attr.Value{basetypes.NewStringValue("a"), basetypes.NewStringValue("b")}

	list, diags := basetypes.NewListValue(fwtypes.StringType, elems)
	if diags.HasError() {
		t.Fatalf("NewListValue: %v", diags)
	}
	set, diags := basetypes.NewSetValue(fwtypes.StringType, elems)
	if diags.HasError() {
		t.Fatalf("NewSetValue: %v", diags)
	}
	tuple, diags := basetypes.NewTupleValue([]attr.Type{fwtypes.StringType, fwtypes.StringType}, elems)
	if diags.HasError() {
		t.Fatalf("NewTupleValue: %v", diags)
	}
	object, diags := basetypes.NewObjectValue(
		map[string]attr.Type{"a": fwtypes.StringType},
		map[string]attr.Value{"a": basetypes.NewStringValue("a")},
	)
	if diags.HasError() {
		t.Fatalf("NewObjectValue: %v", diags)
	}

	tests := map[string]struct {
		value     attr.Value
		wantOK    bool
		wantCount int
	}{
		"nil":                  {value: nil, wantOK: false},
		"list":                 {value: list, wantOK: true, wantCount: 2},
		"set":                  {value: set, wantOK: true, wantCount: 2},
		"tuple":                {value: tuple, wantOK: true, wantCount: 2},
		"empty list":           {value: basetypes.NewListValueMust(fwtypes.StringType, nil), wantOK: true, wantCount: 0},
		"object is not a list": {value: object, wantOK: false},
		"string is not a list": {value: basetypes.NewStringValue("a"), wantOK: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ok, got := bemtypes.ChildItems(tc.value)
			if ok != tc.wantOK {
				t.Fatalf("ChildItems() ok = %v, want %v", ok, tc.wantOK)
			}
			if ok && len(got) != tc.wantCount {
				t.Errorf("ChildItems() returned %d items, want %d", len(got), tc.wantCount)
			}
		})
	}
}

func TestChildAttributes(t *testing.T) {
	t.Parallel()

	object, diags := basetypes.NewObjectValue(
		map[string]attr.Type{"a": fwtypes.StringType, "b": fwtypes.StringType},
		map[string]attr.Value{"a": basetypes.NewStringValue("1"), "b": basetypes.NewStringValue("2")},
	)
	if diags.HasError() {
		t.Fatalf("NewObjectValue: %v", diags)
	}
	m, diags := basetypes.NewMapValue(fwtypes.StringType, map[string]attr.Value{
		"a": basetypes.NewStringValue("1"),
	})
	if diags.HasError() {
		t.Fatalf("NewMapValue: %v", diags)
	}

	tests := map[string]struct {
		value     attr.Value
		wantOK    bool
		wantCount int
	}{
		"nil":                   {value: nil, wantOK: false},
		"object":                {value: object, wantOK: true, wantCount: 2},
		"map":                   {value: m, wantOK: true, wantCount: 1},
		"list is not an object": {value: basetypes.NewListValueMust(fwtypes.StringType, nil), wantOK: false},
		"string is not object":  {value: basetypes.NewStringValue("a"), wantOK: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ok, got := bemtypes.ChildAttributes(tc.value)
			if ok != tc.wantOK {
				t.Fatalf("ChildAttributes() ok = %v, want %v", ok, tc.wantOK)
			}
			if ok && len(got) != tc.wantCount {
				t.Errorf("ChildAttributes() returned %d attributes, want %d", len(got), tc.wantCount)
			}
		})
	}
}
