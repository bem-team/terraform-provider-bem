package types_test

import (
	"math"
	"math/big"
	"testing"

	bemtypes "github.com/bem-team/terraform-provider-bem/internal/types"
)

func TestNewFloat64Value(t *testing.T) {
	t.Parallel()

	for _, want := range []float64{0, 1, -1, 0.1, 1.25, math.MaxFloat64, math.SmallestNonzeroFloat64} {
		got := bemtypes.NewFloat64Value(want)
		if got.IsNull() || got.IsUnknown() {
			t.Fatalf("NewFloat64Value(%v) produced a null/unknown value", want)
		}
		if got.ValueFloat64() != want {
			t.Errorf("NewFloat64Value(%v).ValueFloat64() = %v", want, got.ValueFloat64())
		}
	}
}

func TestNewFloat64ValueFromBigFloat(t *testing.T) {
	t.Parallel()

	got, err := bemtypes.NewFloat64ValueFromBigFloat(big.NewFloat(2.5))
	if err != nil {
		t.Fatalf("NewFloat64ValueFromBigFloat: %v", err)
	}
	if got.ValueFloat64() != 2.5 {
		t.Errorf("got %v, want 2.5", got.ValueFloat64())
	}
}

func TestNewFloat64ValueFromString(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input   string
		want    float64
		wantErr bool
	}{
		"integer":            {input: "3", want: 3},
		"decimal":            {input: "1.25", want: 1.25},
		"negative":           {input: "-0.5", want: -0.5},
		"exponent":           {input: "1e3", want: 1000},
		"zero":               {input: "0", want: 0},
		"not a number":       {input: "abc", wantErr: true},
		"empty string":       {input: "", wantErr: true},
		"trailing garbage":   {input: "1.5x", wantErr: true},
		"repeating fraction": {input: "0.1", want: 0.1},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got, err := bemtypes.NewFloat64ValueFromString(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("NewFloat64ValueFromString(%q) = %v, want an error", tc.input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("NewFloat64ValueFromString(%q): %v", tc.input, err)
			}
			if got.ValueFloat64() != tc.want {
				t.Errorf("NewFloat64ValueFromString(%q) = %v, want %v", tc.input, got.ValueFloat64(), tc.want)
			}
		})
	}
}

func TestNewNumberValueFromString(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input   string
		want    string
		wantErr bool
	}{
		"integer":      {input: "3", want: "3"},
		"decimal":      {input: "1.25", want: "1.25"},
		"exponent":     {input: "2e2", want: "200"},
		"not a number": {input: "abc", wantErr: true},
		"empty string": {input: "", wantErr: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got, err := bemtypes.NewNumberValueFromString(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("NewNumberValueFromString(%q) = %v, want an error", tc.input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("NewNumberValueFromString(%q): %v", tc.input, err)
			}
			want, _, err := big.ParseFloat(tc.want, 10, 512, big.ToNearestEven)
			if err != nil {
				t.Fatalf("parsing expectation %q: %v", tc.want, err)
			}
			if got.ValueBigFloat().Cmp(want) != 0 {
				t.Errorf("NewNumberValueFromString(%q) = %v, want %v", tc.input, got.ValueBigFloat(), want)
			}
		})
	}
}

// A NumberValue keeps the 512-bit precision the string was parsed at, where
// the float64 conversion necessarily rounds. This is the whole reason both
// constructors exist, so pin the difference down.
func TestNewNumberValueFromStringKeepsPrecisionBeyondFloat64(t *testing.T) {
	t.Parallel()

	// 20 significant digits — more than float64's ~15-17.
	const input = "0.12345678901234567890"

	number, err := bemtypes.NewNumberValueFromString(input)
	if err != nil {
		t.Fatalf("NewNumberValueFromString: %v", err)
	}

	exact, _, err := big.ParseFloat(input, 10, 512, big.ToNearestEven)
	if err != nil {
		t.Fatalf("ParseFloat: %v", err)
	}
	if number.ValueBigFloat().Cmp(exact) != 0 {
		t.Errorf("NumberValue lost precision: got %v, want %v", number.ValueBigFloat(), exact)
	}

	rounded, _ := exact.Float64()
	if new(big.Float).SetFloat64(rounded).Cmp(exact) == 0 {
		t.Skip("float64 happens to represent this value exactly; pick a different literal")
	}
}

func TestUnsafeConstructorsPanicOnBadInput(t *testing.T) {
	t.Parallel()

	t.Run("NewFloat64ValueFromStringUnsafe", func(t *testing.T) {
		t.Parallel()
		defer func() {
			if recover() == nil {
				t.Error("expected a panic on unparseable input")
			}
		}()
		bemtypes.NewFloat64ValueFromStringUnsafe("not-a-number")
	})

	t.Run("NewNumberValueFromStringUnsafe", func(t *testing.T) {
		t.Parallel()
		defer func() {
			if recover() == nil {
				t.Error("expected a panic on unparseable input")
			}
		}()
		bemtypes.NewNumberValueFromStringUnsafe("not-a-number")
	})
}

func TestUnsafeConstructorsSucceedOnGoodInput(t *testing.T) {
	t.Parallel()

	if got := bemtypes.NewFloat64ValueFromStringUnsafe("1.5"); got.ValueFloat64() != 1.5 {
		t.Errorf("NewFloat64ValueFromStringUnsafe = %v, want 1.5", got.ValueFloat64())
	}
	if got := bemtypes.NewNumberValueFromStringUnsafe("1.5"); got.ValueBigFloat().Cmp(big.NewFloat(1.5)) != 0 {
		t.Errorf("NewNumberValueFromStringUnsafe = %v, want 1.5", got.ValueBigFloat())
	}
}
