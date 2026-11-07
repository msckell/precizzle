package expr

import (
	"strings"
	"testing"

	"github.com/msckell/precizzle/backend/internal/calc"
)

func TestEvaluate(t *testing.T) {
	tests := []struct {
		name       string
		expression string
		value      string
		kind       calc.Kind
	}{
		// Precedence and grouping (D42).
		{"product before sum", "2*3^2", "18", calc.Exact},
		{"unary minus below power", "-3^2*2", "-18", calc.Exact},
		{"negated base in a group", "(-2)^2", "4", calc.Exact},
		{"quotients from the left", "8/4/2", "1", calc.Exact},
		{"differences from the left", "1-2-3", "-4", calc.Exact},
		{"product after quotient", "2/4*2", "1", calc.Exact},
		{"negated exponent group", "2^-(1)", "0.5", calc.Exact},
		{"double negated exponent", "2^--2", "4", calc.Exact},
		{"computed exponent", "2^(3-1)", "4", calc.Exact},
		{"nested roots", "sqrt(sqrt(16))", "2", calc.Exact},
		{"spaces between tokens", " 50 + 200 + 15 % ", "280", calc.Exact},

		// An expression with no operation returns its operand in canonical form.
		{"lone number", "5.0", "5", calc.Exact},
		{"lone fraction", "[1/7]", "1/7", calc.Repeating},
		{"lone approximate value", "[~1.50]", "~1.5", calc.Approximate},
		{"groups only", "((2))", "2", calc.Exact},
		{"negative zero", "-0", "0", calc.Exact},
		{"negated fraction", "-[1/3]", "-1/3", calc.Repeating},
		{"double negation", "--2", "2", calc.Exact},

		// Percentages (D43).
		{"meaning: discount on a subtracted term", "50-200-15%", "-120", calc.Exact},
		{"meaning: percentage in a product", "200-15%*2", "199.7", calc.Exact},
		{"three percentages compound", "200+15%+10%+5%", "265.65", calc.Exact},
		{"markup then discount", "200+15%-10%", "207", calc.Exact},
		{"percentage after a first-term percentage", "15%+10%", "0.165", calc.Exact},
		{"double minus on the percentage", "200--15%", "230", calc.Exact},
		{"negated first-term percentage", "-15%+200", "199.85", calc.Exact},
		{"negated first term keeps its sign", "-(200)+15%", "-230", calc.Exact},
		{"product as the term", "50*2+15%", "115", calc.Exact},
		{"power as the term", "10-2^2+50%", "4", calc.Exact},
		{"percentage times a number is not a percentage term", "50+200+15%*2", "250.3", calc.Exact},
		{"powered percentage is not a percentage term", "200+15%^2", "200.0225", calc.Exact},
		{"percentage of a square root", "200+sqrt(4)%", "204", calc.Exact},
		{"percentage of a value", "200+[1/2]%", "201", calc.Exact},
		{"repeating percentage", "1+[1/3]%", "301/300", calc.Repeating},
		{"approximate percentage", "100+[~2]%", "~102", calc.Approximate},
		{"percentage in a group is plain", "200+(15%)", "200.15", calc.Exact},
		{"lone percentage in a group", "(15%)", "0.15", calc.Exact},
		{"subtotal compounds", "(50+200)+15%+10%", "316.25", calc.Exact},
		{"compound on the last term", "50+200+15%+10%", "303", calc.Exact},
		{"compound on a subtracted term", "50-200+15%-10%", "-157", calc.Exact},

		// Intermediate results (D44).
		{"approximation spreads", "[~2]+1", "~3", calc.Approximate},
		{"approximate zero stays approximate", "sqrt(2)*0", "~0", calc.Approximate},
		{"exact fractions add up", "[1/3]+[2/3]", "1", calc.Exact},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Evaluate(tt.expression)
			if err != nil || got.Value != tt.value || got.Kind != tt.kind {
				t.Errorf("Evaluate(%q) = %+v, %v, want value %q, kind %v", tt.expression, got, err, tt.value, tt.kind)
			}
		})
	}
}

func TestEvaluateErrors(t *testing.T) {
	tests := []struct {
		name       string
		expression string
		wantErr    error
		position   int
	}{
		{"divide by a zero percentage", "200/0%", calc.ErrDivisionByZero, 3},
		{"zero to a computed negative exponent", "0^(0-1)", calc.ErrDivisionByZero, 1},
		{"divide by a computed zero", "1/(1-1)", calc.ErrDivisionByZero, 1},
		{"percentage exponent", "2^50%", calc.ErrExponentNotInteger, 1},
		{"computed fractional exponent", "2^(1/2)", calc.ErrExponentNotInteger, 1},
		{"approximate exponent", "2^[~2]", calc.ErrExponentNotInteger, 1},
		{"inner power gives a fractional exponent", "2^3^-1", calc.ErrExponentNotInteger, 1},
		{"computed negative square root", "2*sqrt(1-5)", calc.ErrNegativeSquareRoot, 2},
		{"product too large", "10^39*10", calc.ErrResultTooLarge, 5},
		{"quotient within the limits", "[1e-999]/10+1/0", calc.ErrDivisionByZero, 13},
		{"quotient too precise", "[1e-1000]/10", calc.ErrResultTooPrecise, 9},
		{"plain percentage too precise", "[1e-999]%", calc.ErrResultTooPrecise, 8},
		{"percentage term too large points to the %", "[9e39]+20%", calc.ErrResultTooLarge, 9},
		{"later percentage term too large", "5+[9e39]+20%", calc.ErrResultTooLarge, 11},

		// Evaluation runs from left to right and stops at the first failure.
		{"first failure wins", "1/0+sqrt(-4)", calc.ErrDivisionByZero, 1},
		{"first failure wins, reversed", "sqrt(-4)+1/0", calc.ErrNegativeSquareRoot, 0},
		{"operand of a negation fails", "-sqrt(-4)", calc.ErrNegativeSquareRoot, 1},
		{"term fails before its percentage", "1/0+15%", calc.ErrDivisionByZero, 1},
		{"percentage base fails", "200+sqrt(-4)%", calc.ErrNegativeSquareRoot, 4},
		{"failing percentage stops the run", "[9e39]+20%+1/0%", calc.ErrResultTooLarge, 9},
		{"a sum runs as soon as its terms are known", "[9e39]+[9e39]+1/0", calc.ErrResultTooLarge, 6},

		// Checks 8 to 10 come before the evaluation.
		{"syntax before evaluation", "1/0+", ErrInvalidExpression, 4},
		{"operand limits before evaluation", "1/0+1e40", calc.ErrOperandOutOfRange, 4},
		{"operation count before evaluation", "1/0" + strings.Repeat("+1", 100), ErrExpressionTooComplex, 201},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Evaluate(tt.expression)
			if got != (calc.Result{}) {
				t.Errorf("result = %+v, want none", got)
			}
			checkError(t, tt.expression, err, tt.wantErr, tt.position)
		})
	}
}

// TestPercentageMatchesD7 checks that every case of D7 written as an
// expression returns what the percentage operation returns (D43).
func TestPercentageMatchesD7(t *testing.T) {
	parse := func(s string) calc.Number {
		n, err := calc.Parse(s)
		if err != nil {
			t.Fatalf("calc.Parse(%q) error = %v", s, err)
		}
		return n
	}
	operators := []struct {
		symbol string
		op     calc.PercentOperator
	}{
		{"+", calc.PercentAdd},
		{"-", calc.PercentSubtract},
		{"*", calc.PercentMultiply},
		{"/", calc.PercentDivide},
	}
	for _, operands := range [][2]string{{"200", "15"}, {"1/3", "~2.5"}, {"-7.5", "1/8"}, {"2/3", "-12.5"}} {
		base, percent := operands[0], operands[1]
		for _, o := range operators {
			expression := "[" + base + "]" + o.symbol + "[" + percent + "]%"
			want, wantErr := calc.ApplyPercentage(parse(base), o.op, parse(percent))
			if got, err := Evaluate(expression); got != want || err != nil || wantErr != nil {
				t.Errorf("Evaluate(%q) = %+v, %v, want %+v, %v", expression, got, err, want, wantErr)
			}
		}
		want, _ := calc.Percentage(parse(percent))
		if got, err := Evaluate("[" + percent + "]%"); got != want || err != nil {
			t.Errorf("Evaluate(%q) = %+v, %v, want %+v", "["+percent+"]%", got, err, want)
		}
	}
}
