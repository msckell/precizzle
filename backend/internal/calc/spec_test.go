package calc

import (
	"errors"
	"strings"
	"testing"
)

// TestSpecExamples checks every row of the examples and errors tables in
// section 3 of docs/SPEC.md. Each case is named after its row.
func TestSpecExamples(t *testing.T) {
	binary := func(op func(a, b Number) (Result, error), a, b string) func(t *testing.T) (Result, error) {
		return func(t *testing.T) (Result, error) { return op(num(t, a), num(t, b)) }
	}
	unary := func(op func(x Number) (Result, error), x string) func(t *testing.T) (Result, error) {
		return func(t *testing.T) (Result, error) { return op(num(t, x)) }
	}
	percent := func(base string, op PercentOperator, p string) func(t *testing.T) (Result, error) {
		return func(t *testing.T) (Result, error) { return ApplyPercentage(num(t, base), op, num(t, p)) }
	}
	third := "0." + strings.Repeat("3", 100)

	tests := []struct {
		row     string
		calc    func(t *testing.T) (Result, error)
		want    Result
		wantErr error
	}{
		{row: "0.1 + 0.2", calc: binary(Add, "0.1", "0.2"), want: Result{"0.3", "0.3", Exact}},
		{row: "1.10 + 1.20", calc: binary(Add, "1.10", "1.20"), want: Result{"2.3", "2.3", Exact}},
		{row: "1 ÷ 3", calc: binary(Divide, "1", "3"), want: Result{"1/3", third, Repeating}},
		{row: "1/3 × 3", calc: binary(Multiply, "1/3", "3"), want: Result{"1", "1", Exact}},
		{row: "2 ÷ 3", calc: binary(Divide, "2", "3"), want: Result{"2/3", "0." + strings.Repeat("6", 99) + "7", Repeating}},
		{row: "1 ÷ 7", calc: binary(Divide, "1", "7"), want: Result{"1/7", "0." + strings.Repeat("142857", 16) + "1429", Repeating}},
		{row: "2 ^ −2", calc: binary(Power, "2", "-2"), want: Result{"0.25", "0.25", Exact}},
		{row: "3 ^ −1", calc: binary(Power, "3", "-1"), want: Result{"1/3", third, Repeating}},
		{row: "0 ^ 0", calc: binary(Power, "0", "0"), want: Result{"1", "1", Exact}},
		{row: "√2", calc: unary(Sqrt, "2"), want: Result{"~1.4142135623730950488", "1.4142135623730950488", Approximate}},
		{row: "√2.25", calc: unary(Sqrt, "2.25"), want: Result{"1.5", "1.5", Exact}},
		{row: "√(1/9)", calc: unary(Sqrt, "1/9"), want: Result{"1/3", third, Repeating}},
		{row: "~√2 × ~√2", calc: binary(Multiply, "~1.4142135623730950488", "~1.4142135623730950488"), want: Result{"~2", "2", Approximate}},
		{row: "15%", calc: unary(Percentage, "15"), want: Result{"0.15", "0.15", Exact}},
		{row: "200 + 15%", calc: percent("200", PercentAdd, "15"), want: Result{"230", "230", Exact}},
		{row: "200 − 15%", calc: percent("200", PercentSubtract, "15"), want: Result{"170", "170", Exact}},
		{row: "200 × 15%", calc: percent("200", PercentMultiply, "15"), want: Result{"30", "30", Exact}},
		{row: "200 ÷ 15%", calc: percent("200", PercentDivide, "15"), want: Result{"4000/3", "1333." + strings.Repeat("3", 100), Repeating}},
		{row: "1 ÷ 3e30", calc: binary(Divide, "1", "3e30"), want: Result{"1/3" + strings.Repeat("0", 30), "3." + strings.Repeat("3", 100) + "e-31", Repeating}},
		{row: "1e-25 + 0", calc: binary(Add, "1e-25", "0"), want: Result{"0." + strings.Repeat("0", 24) + "1", "1e-25", Exact}},

		{row: "1 ÷ 0", calc: binary(Divide, "1", "0"), wantErr: ErrDivisionByZero},
		{row: "0 ^ −1", calc: binary(Power, "0", "-1"), wantErr: ErrDivisionByZero},
		{row: "200 ÷ 0%", calc: percent("200", PercentDivide, "0"), wantErr: ErrDivisionByZero},
		{row: "√−4", calc: unary(Sqrt, "-4"), wantErr: ErrNegativeSquareRoot},
		{row: "10 ^ 40", calc: binary(Power, "10", "40"), wantErr: ErrResultTooLarge},
		{row: "1e-1000 × 1e-1000", calc: binary(Multiply, "1e-1000", "1e-1000"), wantErr: ErrResultTooPrecise},
		{row: "0.25 ^ 501", calc: binary(Power, "0.25", "501"), wantErr: ErrResultTooPrecise},
		{row: "2 ^ 1001", calc: binary(Power, "2", "1001"), wantErr: ErrExponentOutOfRange},
		{row: "2 ^ 0.5", calc: binary(Power, "2", "0.5"), wantErr: ErrExponentNotInteger},
		{row: "2 ^ ~2", calc: binary(Power, "2", "~2"), wantErr: ErrExponentNotInteger},
	}
	for _, tt := range tests {
		t.Run(tt.row, func(t *testing.T) {
			got, err := tt.calc(t)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("result = %+v, want %+v", got, tt.want)
			}
		})
	}
}

// TestSpecChaining follows the chaining example of the API contract: each
// step sends back the value returned by the previous one.
func TestSpecChaining(t *testing.T) {
	third, err := Divide(num(t, "1"), num(t, "3"))
	if err != nil || third.Value != "1/3" {
		t.Fatalf("1 ÷ 3 = %+v, %v", third, err)
	}
	whole, err := Multiply(num(t, third.Value), num(t, "3"))
	if err != nil || whole != (Result{"1", "1", Exact}) {
		t.Errorf("1/3 × 3 = %+v, %v, want exactly 1", whole, err)
	}
	twoThirds, err := Subtract(num(t, "1"), num(t, third.Value))
	if err != nil || twoThirds != (Result{"2/3", "0." + strings.Repeat("6", 99) + "7", Repeating}) {
		t.Errorf("1 − 1/3 = %+v, %v", twoThirds, err)
	}
	root, err := Sqrt(num(t, "2"))
	if err != nil {
		t.Fatalf("√2 error = %v", err)
	}
	square, err := Multiply(num(t, root.Value), num(t, root.Value))
	if err != nil || square != (Result{"~2", "2", Approximate}) {
		t.Errorf("√2 × √2 = %+v, %v, want ~2", square, err)
	}
}
