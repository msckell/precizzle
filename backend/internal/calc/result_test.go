package calc

import (
	"errors"
	"math/big"
	"strings"
	"testing"
)

// rat builds a test value from a string that math/big parses directly, independently of Parse.
func rat(t *testing.T, s string) *big.Rat {
	t.Helper()
	r, ok := new(big.Rat).SetString(s)
	if !ok {
		t.Fatalf("bad test value %q", s)
	}
	return r
}

func TestKindString(t *testing.T) {
	tests := []struct {
		kind Kind
		want string
	}{
		{Exact, "exact"},
		{Repeating, "repeating"},
		{Approximate, "approximate"},
	}
	for _, tt := range tests {
		if got := tt.kind.String(); got != tt.want {
			t.Errorf("Kind(%d).String() = %q, want %q", tt.kind, got, tt.want)
		}
	}
}

func TestResultFormatsEachKind(t *testing.T) {
	tests := []struct {
		name        string
		value       string
		approximate bool
		want        Result
	}{
		{"zero", "0", false, Result{"0", "0", Exact}},
		{"integer", "-42", false, Result{"-42", "-42", Exact}},
		{"exact decimal", "2.3", false, Result{"2.3", "2.3", Exact}},
		{"exact at 1e-20 stays plain", "1e-20", false, Result{"0.00000000000000000001", "0.00000000000000000001", Exact}},
		{"exact below 1e-20", "1e-25", false, Result{"0." + strings.Repeat("0", 24) + "1", "1e-25", Exact}},
		{"exact tiny with several digits", "-1.25e-30", false, Result{"-0." + strings.Repeat("0", 29) + "125", "-1.25e-30", Exact}},
		{"repeating rounds down", "1/3", false, Result{"1/3", "0." + strings.Repeat("3", 100), Repeating}},
		{"repeating rounds up", "2/3", false, Result{"2/3", "0." + strings.Repeat("6", 99) + "7", Repeating}},
		{"repeating negative", "-4000/3", false, Result{"-4000/3", "-1333." + strings.Repeat("3", 100), Repeating}},
		{"repeating tiny", "1/3000000000000000000000000000000", false, Result{"1/3000000000000000000000000000000", "3." + strings.Repeat("3", 100) + "e-31", Repeating}},
		{"approximate rounds to 20 digits", "1.414213562373095048801688", true, Result{"~1.4142135623730950488", "1.4142135623730950488", Approximate}},
		{"approximate integer", "2", true, Result{"~2", "2", Approximate}},
		{"approximate negative", "-0.123", true, Result{"~-0.123", "-0.123", Approximate}},
		{"approximate tiny", "1/3000000000000000000000000000000", true, Result{"~3.33333333333333333333e-31", "3.33333333333333333333e-31", Approximate}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := rat(t, tt.value)
			got, err := result(r.Num(), r.Denom(), tt.approximate)
			if err != nil {
				t.Fatalf("result error = %v", err)
			}
			if got != tt.want {
				t.Errorf("result = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestResultRoundsHalfEven(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{"tie at the 21st digit goes to even 0", "1.000000000000000000005", "1"},
		{"tie at the 21st digit goes to even 2", "1.000000000000000000015", "1.00000000000000000002"},
		{"negative tie goes to even", "-1.000000000000000000015", "-1.00000000000000000002"},
		{"above the tie rounds up", "1.0000000000000000000050001", "1.00000000000000000001"},
		{"below the tie rounds down", "1.0000000000000000000149999", "1.00000000000000000001"},
		{"rounding carries into the integer part", "9.999999999999999999995", "10"},
		{"rounding a tiny mantissa up to 10 renormalizes", "9.9999999999999999999999e-25", "1e-24"},
		{"rounding up to 10^-20 writes plain notation", "9.9999999999999999999999e-21", "0.00000000000000000001"},
		{"tiny tie goes to even", "1.000000000000000000005e-30", "1e-30"},
		{"mantissa stops at the 1000th fractional digit", "1.23456789012345678901234e-990", "1.2345678901e-990"},
		{"just above half of 10^-1000 rounds to 10^-1000", "6e-1001", "1e-1000"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := rat(t, tt.value)
			got, err := result(r.Num(), r.Denom(), true)
			if err != nil {
				t.Fatalf("result error = %v", err)
			}
			if got.Decimal != tt.want || got.Value != "~"+tt.want {
				t.Errorf("result = %+v, want decimal %q", got, tt.want)
			}
		})
	}
}

func TestResultLimits(t *testing.T) {
	tests := []struct {
		name        string
		value       string
		approximate bool
		want        error
	}{
		{"just below 10^40", strings.Repeat("9", 40), false, nil},
		{"10^40", "1e40", false, ErrResultTooLarge},
		{"-10^40", "-1e40", false, ErrResultTooLarge},
		{"approximate 10^40", "1e40", true, ErrResultTooLarge},
		{"1000 fractional digits", "1e-1000", false, nil},
		{"1001 fractional digits", "1e-1001", false, ErrResultTooPrecise},
		{"1/2^1001 has a small denominator but too many digits", "1/" + new(big.Int).Lsh(big.NewInt(1), 1001).String(), false, ErrResultTooPrecise},
		{"denominator 3·10^999", "1/3" + strings.Repeat("0", 999), false, nil},
		{"denominator 3·10^1000", "1/3" + strings.Repeat("0", 1000), false, ErrResultTooPrecise},
		{"too large wins over too precise", "1" + strings.Repeat("0", 1040) + "1/3" + strings.Repeat("0", 1000), false, ErrResultTooLarge},
		{"approximate 10^-1000", "1e-1000", true, nil},
		{"approximate past 1000 digits is rounded, not rejected", "6e-1001", true, nil},
		{"approximate 10^-1001 rounds to 0", "1e-1001", true, ErrResultTooPrecise},
		{"approximate half of 10^-1000 rounds to 0", "5e-1001", true, ErrResultTooPrecise},
		{"approximate far below 10^-1000", "1e-1200", true, ErrResultTooPrecise},
		{"approximate zero", "0", true, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := rat(t, tt.value)
			_, err := result(r.Num(), r.Denom(), tt.approximate)
			if !errors.Is(err, tt.want) {
				t.Errorf("result error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestResultValueParsesBack(t *testing.T) {
	tests := []string{
		"0", "-7", "0.3", "1e-25", "-1.25e-30", "1/3", "-4000/3", "1/3000000000000000000000000000000",
		"1/" + strings.Repeat("9", 1000), "0." + strings.Repeat("0", 999) + "1",
		"~1.414213562373095048801688", "~-0.123", "~1/3000000000000000000000000000000",
		"~9.9999999999999999999999e-25", "~9.9999999999999999999999e-21", "~1.23456789012345678901234e-990", "~6e-1001",
	}
	for _, in := range tests {
		t.Run(in, func(t *testing.T) {
			approximate := strings.HasPrefix(in, "~")
			r := rat(t, strings.TrimPrefix(in, "~"))
			first, err := result(r.Num(), r.Denom(), approximate)
			if err != nil {
				t.Fatalf("result error = %v", err)
			}
			n, err := Parse(first.Value)
			if err != nil {
				t.Fatalf("Parse(%q) error = %v", first.Value, err)
			}
			second, err := newResult(n)
			if err != nil {
				t.Fatalf("newResult error = %v", err)
			}
			if second != first {
				t.Errorf("result of the parsed value = %+v, want %+v", second, first)
			}
			if !approximate && n.rat.Cmp(r) != 0 {
				t.Errorf("parsed value = %v, want %v", n.rat, r)
			}
		})
	}
}

func FuzzParse(f *testing.F) {
	for _, seed := range []string{
		"0", "-12.5", "1e-3", "2.5E10", "1/3", "-2/7", "4/2", "~1.4142135623730950488", "~-3.2e-25",
		"1e39", "1e40", "1e-1000", "1e-1001", "+1", ".5", "0x10", "1/0", "~1/3", "9.99e39", "~6e-1000",
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, s string) {
		n, err := Parse(s)
		if err != nil {
			if !errors.Is(err, ErrInvalidNumber) && !errors.Is(err, ErrOperandOutOfRange) {
				t.Fatalf("Parse(%q) error = %v", s, err)
			}
			return
		}
		first, err := newResult(n)
		if err != nil {
			t.Fatalf("operand %q is valid but its result fails: %v", s, err)
		}
		m, err := Parse(first.Value)
		if err != nil {
			t.Fatalf("value %q of operand %q does not parse: %v", first.Value, s, err)
		}
		second, err := newResult(m)
		if err != nil || second != first {
			t.Fatalf("value %q of operand %q reads back as %+v, %v", first.Value, s, second, err)
		}
	})
}
