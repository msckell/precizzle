package expr

import (
	"testing"

	"github.com/msckell/precizzle/backend/internal/calc"
)

// TestSpecExamples checks every row of the expression examples in section 5
// of docs/SPEC.md. Each case is named after its row.
func TestSpecExamples(t *testing.T) {
	tests := []struct {
		expression string
		value      string
		kind       calc.Kind
	}{
		{"0.1+0.2", "0.3", calc.Exact},
		{"2+3*4", "14", calc.Exact},
		{"(2+3)*4", "20", calc.Exact},
		{"2^3^2", "512", calc.Exact},
		{"-2^2", "-4", calc.Exact},
		{"2^-2", "0.25", calc.Exact},
		{"1/3*3", "1", calc.Exact},
		{"[1/7]*7", "1", calc.Exact},
		{"sqrt(9/4)", "1.5", calc.Exact},
		{"sqrt(2)*sqrt(2)", "~2", calc.Approximate},
		{"200+15%", "230", calc.Exact},
		{"200-15%", "170", calc.Exact},
		{"200*15%", "30", calc.Exact},
		{"200/15%", "4000/3", calc.Repeating},
		{"15%", "0.15", calc.Exact},
		{"50+200+15%", "280", calc.Exact},
		{"50-200+15%", "-180", calc.Exact},
		{"200+15%+10%", "253", calc.Exact},
		{"(50+200)+15%", "287.5", calc.Exact},
		{"15%+200", "200.15", calc.Exact},
		{"200+-15%", "170", calc.Exact},
	}
	for _, tt := range tests {
		t.Run(tt.expression, func(t *testing.T) {
			got, err := Evaluate(tt.expression)
			if err != nil || got.Value != tt.value || got.Kind != tt.kind {
				t.Errorf("Evaluate(%q) = %+v, %v, want value %q, kind %v", tt.expression, got, err, tt.value, tt.kind)
			}
		})
	}
}

// TestSpecErrors checks every row of the expression errors in section 5 of docs/SPEC.md.
func TestSpecErrors(t *testing.T) {
	tests := []struct {
		expression string
		wantErr    error
		position   int
	}{
		{"1/0", calc.ErrDivisionByZero, 1},
		{"0^-1", calc.ErrDivisionByZero, 1},
		{"sqrt(-4)", calc.ErrNegativeSquareRoot, 0},
		{"2^0.5", calc.ErrExponentNotInteger, 1},
		{"2^1001", calc.ErrExponentOutOfRange, 1},
		{"10^50/10^20", calc.ErrResultTooLarge, 2},
		{"2+", ErrInvalidExpression, 2},
		{"(2+3", ErrInvalidExpression, 0},
		{"(2+", ErrInvalidExpression, 3},
		{"2(3)", ErrInvalidExpression, 1},
		{"15%%", ErrInvalidExpression, 3},
		{"1.2.3", calc.ErrInvalidNumber, 0},
	}
	for _, tt := range tests {
		t.Run(tt.expression, func(t *testing.T) {
			got, err := Evaluate(tt.expression)
			if got != (calc.Result{}) {
				t.Errorf("result = %+v, want none", got)
			}
			checkError(t, tt.expression, err, tt.wantErr, tt.position)
		})
	}
}
