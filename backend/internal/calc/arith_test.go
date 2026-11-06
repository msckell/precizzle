package calc

import (
	"errors"
	"strings"
	"testing"
)

// num parses a test operand and fails the test if it is invalid.
func num(t *testing.T, s string) Number {
	t.Helper()
	n, err := Parse(s)
	if err != nil {
		t.Fatalf("Parse(%q) error = %v", s, err)
	}
	return n
}

type binaryCase struct {
	name    string
	a, b    string
	want    Result
	wantErr error
}

func runBinary(t *testing.T, op func(a, b Number) (Result, error), tests []binaryCase) {
	t.Helper()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := op(num(t, tt.a), num(t, tt.b))
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("result = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestAdd(t *testing.T) {
	runBinary(t, Add, []binaryCase{
		{name: "0.1 + 0.2", a: "0.1", b: "0.2", want: Result{"0.3", "0.3", Exact}},
		{name: "1.10 + 1.20", a: "1.10", b: "1.20", want: Result{"2.3", "2.3", Exact}},
		{name: "negative sum", a: "-0.1", b: "-0.2", want: Result{"-0.3", "-0.3", Exact}},
		{name: "sum to zero", a: "0.5", b: "-1/2", want: Result{"0", "0", Exact}},
		{name: "fractions", a: "1/3", b: "1/3", want: Result{"2/3", "0." + strings.Repeat("6", 99) + "7", Repeating}},
		{name: "fractions that terminate", a: "1/3", b: "2/3", want: Result{"1", "1", Exact}},
		{name: "tiny plus zero", a: "1e-25", b: "0", want: Result{"0." + strings.Repeat("0", 24) + "1", "1e-25", Exact}},
		{name: "approximate operand", a: "~1.5", b: "1", want: Result{"~2.5", "2.5", Approximate}},
		{name: "largest value", a: strings.Repeat("9", 39) + "8", b: "1", want: Result{strings.Repeat("9", 40), strings.Repeat("9", 40), Exact}},
		{name: "too large", a: strings.Repeat("9", 40), b: "1", wantErr: ErrResultTooLarge},
	})
}

func TestSubtract(t *testing.T) {
	runBinary(t, Subtract, []binaryCase{
		{name: "0.3 - 0.1", a: "0.3", b: "0.1", want: Result{"0.2", "0.2", Exact}},
		{name: "1 - 1/3", a: "1", b: "1/3", want: Result{"2/3", "0." + strings.Repeat("6", 99) + "7", Repeating}},
		{name: "equal operands give 0, not -0", a: "-2.5", b: "-2.5", want: Result{"0", "0", Exact}},
		{name: "negative result", a: "1", b: "3", want: Result{"-2", "-2", Exact}},
		{name: "approximate operand", a: "1", b: "~0.25", want: Result{"~0.75", "0.75", Approximate}},
		{name: "too large", a: "-" + strings.Repeat("9", 40), b: "1", wantErr: ErrResultTooLarge},
	})
}

func TestMultiply(t *testing.T) {
	runBinary(t, Multiply, []binaryCase{
		{name: "1/3 × 3", a: "1/3", b: "3", want: Result{"1", "1", Exact}},
		{name: "decimals", a: "1.5", b: "-0.2", want: Result{"-0.3", "-0.3", Exact}},
		{name: "times zero gives 0, not -0", a: "-1", b: "0", want: Result{"0", "0", Exact}},
		{name: "approximate √2 squared", a: "~1.4142135623730950488", b: "~1.4142135623730950488", want: Result{"~2", "2", Approximate}},
		{name: "approximate zero", a: "~0", b: "5", want: Result{"~0", "0", Approximate}},
		{name: "1e-500 × 1e-500", a: "1e-500", b: "1e-500", want: Result{"0." + strings.Repeat("0", 999) + "1", "1e-1000", Exact}},
		{name: "1e-1000 × 1e-1000", a: "1e-1000", b: "1e-1000", wantErr: ErrResultTooPrecise},
		{name: "1e20 × 1e20", a: "1e20", b: "1e20", wantErr: ErrResultTooLarge},
		{name: "approximate mantissa stops at the 1000th digit", a: "~1e-990", b: "1.23456789012345678901234", want: Result{"~1.2345678901e-990", "1.2345678901e-990", Approximate}},
		{name: "approximate rounding up to 10^-1000", a: "~6e-501", b: "1e-500", want: Result{"~1e-1000", "1e-1000", Approximate}},
		{name: "approximate tie rounds to 0", a: "~5e-501", b: "1e-500", wantErr: ErrResultTooPrecise},
	})
}

func TestDivide(t *testing.T) {
	runBinary(t, Divide, []binaryCase{
		{name: "1 ÷ 3", a: "1", b: "3", want: Result{"1/3", "0." + strings.Repeat("3", 100), Repeating}},
		{name: "2 ÷ 3", a: "2", b: "3", want: Result{"2/3", "0." + strings.Repeat("6", 99) + "7", Repeating}},
		{name: "1 ÷ 7", a: "1", b: "7", want: Result{"1/7", "0." + strings.Repeat("142857", 16) + "1429", Repeating}},
		{name: "1 ÷ 4", a: "1", b: "4", want: Result{"0.25", "0.25", Exact}},
		{name: "negative divisor", a: "1", b: "-3", want: Result{"-1/3", "-0." + strings.Repeat("3", 100), Repeating}},
		{name: "1 ÷ 3e30", a: "1", b: "3e30", want: Result{"1/3" + strings.Repeat("0", 30), "3." + strings.Repeat("3", 100) + "e-31", Repeating}},
		{name: "fraction ÷ fraction", a: "1/3", b: "2/3", want: Result{"0.5", "0.5", Exact}},
		{name: "1 ÷ 0", a: "1", b: "0", wantErr: ErrDivisionByZero},
		{name: "0 ÷ 0", a: "0", b: "0", wantErr: ErrDivisionByZero},
		{name: "divide by approximate zero", a: "1", b: "~0", wantErr: ErrDivisionByZero},
		{name: "approximate quotient", a: "~1", b: "3", want: Result{"~0.33333333333333333333", "0.33333333333333333333", Approximate}},
		{name: "denominator 3·10^999", a: "1e-999", b: "3", want: Result{"1/3" + strings.Repeat("0", 999), "3." + strings.Repeat("3", 100) + "e-1000", Repeating}},
		{name: "denominator 3·10^1000", a: "1e-1000", b: "3", wantErr: ErrResultTooPrecise},
		{name: "too large", a: "1e39", b: "0.1", wantErr: ErrResultTooLarge},
	})
}
