package calc

import (
	"errors"
	"strings"
	"testing"
)

func TestSqrt(t *testing.T) {
	// The digits of irrational roots come from an independent BigInt computation.
	tests := []struct {
		in      string
		want    Result
		wantErr error
	}{
		{in: "2", want: Result{"~1.4142135623730950488", "1.4142135623730950488", Approximate}},
		{in: "1/3", want: Result{"~0.57735026918962576451", "0.57735026918962576451", Approximate}},
		{in: "10", want: Result{"~3.162277660168379332", "3.162277660168379332", Approximate}},
		{in: "1e39", want: Result{"~31622776601683793319.9889354443271853372", "31622776601683793319.9889354443271853372", Approximate}},
		{in: "2e-100", want: Result{"~1.4142135623730950488e-50", "1.4142135623730950488e-50", Approximate}},
		{in: "2e-1000", want: Result{"~1.4142135623730950488e-500", "1.4142135623730950488e-500", Approximate}},
		{in: "~2", want: Result{"~1.4142135623730950488", "1.4142135623730950488", Approximate}},
		{in: "0." + strings.Repeat("9", 42), want: Result{"~1", "1", Approximate}},
		{in: "0." + strings.Repeat("9", 42) + "e-40", want: Result{"~0.00000000000000000001", "0.00000000000000000001", Approximate}},
		{in: "2.25", want: Result{"1.5", "1.5", Exact}},
		{in: "1/9", want: Result{"1/3", "0." + strings.Repeat("3", 100), Repeating}},
		{in: "1e38", want: Result{"10000000000000000000", "10000000000000000000", Exact}},
		{in: "1e-40", want: Result{"0.00000000000000000001", "0.00000000000000000001", Exact}},
		{in: "1e-1000", want: Result{"0." + strings.Repeat("0", 499) + "1", "1e-500", Exact}},
		{in: "0", want: Result{"0", "0", Exact}},
		{in: "~0", want: Result{"~0", "0", Approximate}},
		{in: "~4", want: Result{"~2", "2", Approximate}},
		{in: "-4", wantErr: ErrNegativeSquareRoot},
		{in: "-1/4", wantErr: ErrNegativeSquareRoot},
		{in: "~-4", wantErr: ErrNegativeSquareRoot},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := Sqrt(num(t, tt.in))
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Sqrt error = %v, want %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("Sqrt = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestSqrtValueParsesBack(t *testing.T) {
	for _, in := range []string{"2", "1e39", "2e-1000", "0." + strings.Repeat("9", 42) + "e-40"} {
		t.Run(in, func(t *testing.T) {
			got, err := Sqrt(num(t, in))
			if err != nil {
				t.Fatalf("Sqrt error = %v", err)
			}
			again, err := newResult(num(t, got.Value))
			if err != nil || again != got {
				t.Errorf("value %q reads back as %+v, %v", got.Value, again, err)
			}
		})
	}
}
