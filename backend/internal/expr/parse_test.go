package expr

import (
	"errors"
	"strings"
	"testing"

	"github.com/msckell/precizzle/backend/internal/calc"
)

// checkError checks that err is an *Error with the given cause and position.
func checkError(t *testing.T, input string, err, wantErr error, wantPos int) {
	t.Helper()
	e, ok := errors.AsType[*Error](err)
	if !ok {
		t.Fatalf("%q: error = %v, want an *Error", input, err)
	}
	if !errors.Is(e, wantErr) || e.Position != wantPos {
		t.Errorf("%q: error = %v at %d, want %v at %d", input, e.Err, e.Position, wantErr, wantPos)
	}
}

func TestParseAccepts(t *testing.T) {
	for _, s := range []string{
		"0", "12", "0.5", "1e-3", "2.5E10", "2e+3",
		"[1/7]", "[-2.5]", "[~1.4142135623730950488]", "[4/2]",
		"2+3*4", "(2+3)*4", "2^3^2", "-2^2", "2^-2", "--2", "8/4/2",
		"sqrt(2)", "sqrt(9/4)", "sqrt(sqrt(16))", "sqrt(2)*sqrt(2)",
		"15%", "200+15%", "200+-15%", "(50+200)+15%", "sqrt(4)%", "[1/2]%", "(15)%",
		" 2 + 3 ", "( 2 )", "sqrt( 2 )", "2 ^ - 2",
	} {
		if _, err := parse(s); err != nil {
			t.Errorf("parse(%q) error = %v", s, err)
		}
	}
}

func TestParseErrors(t *testing.T) {
	tests := []struct {
		input   string
		wantErr error
		wantPos int
	}{
		// SPEC §5 examples.
		{"2+", ErrInvalidExpression, 2},
		{"(2+3", ErrInvalidExpression, 0},
		{"(2+", ErrInvalidExpression, 3},
		{"2(3)", ErrInvalidExpression, 1},
		{"15%%", ErrInvalidExpression, 3},
		{"1.2.3", calc.ErrInvalidNumber, 0},

		// Missing operands point to the end.
		{"", ErrInvalidExpression, 0},
		{"   ", ErrInvalidExpression, 3},
		{"2 + ", ErrInvalidExpression, 4},
		{"-", ErrInvalidExpression, 1},
		{"2^", ErrInvalidExpression, 2},
		{"sqrt(", ErrInvalidExpression, 5},

		// Unexpected tokens and characters outside the syntax.
		{")", ErrInvalidExpression, 0},
		{"()", ErrInvalidExpression, 1},
		{"2)", ErrInvalidExpression, 1},
		{"2 3", ErrInvalidExpression, 2},
		{"*2", ErrInvalidExpression, 0},
		{"+2", ErrInvalidExpression, 0},
		{"2**3", ErrInvalidExpression, 2},
		{"%", ErrInvalidExpression, 0},
		{"2+%", ErrInvalidExpression, 2},
		{"]", ErrInvalidExpression, 0},
		{"2\t+3", ErrInvalidExpression, 1},
		{"2\x00", ErrInvalidExpression, 1},
		{"\x00", ErrInvalidExpression, 0},
		{"(2\x00", ErrInvalidExpression, 2},
		{"2+é", ErrInvalidExpression, 2},
		{"2×3", ErrInvalidExpression, 1},
		{"2,5", ErrInvalidExpression, 1},
		{"x", ErrInvalidExpression, 0},
		{"2+e", ErrInvalidExpression, 2},
		{"1 2", ErrInvalidExpression, 2},
		{"(2 3)", ErrInvalidExpression, 3},

		// Unclosed groups point to their opener; the innermost is met first.
		{"sqrt(2", ErrInvalidExpression, 0},
		{"2*sqrt(2+3", ErrInvalidExpression, 2},
		{"[1/7", ErrInvalidExpression, 0},
		{"2+[1/7", ErrInvalidExpression, 2},
		{"((2+3)", ErrInvalidExpression, 0},
		{"((2+3", ErrInvalidExpression, 1},

		// D54: a token is checked only when the grammar accepts it there.
		{"2 1.2.3", ErrInvalidExpression, 2},
		{"1.2.3+(", calc.ErrInvalidNumber, 0},
		// D54: sqrt( is matched character by character.
		{"sqrx(4)", ErrInvalidExpression, 3},
		{"sqrt (4)", ErrInvalidExpression, 4},
		{"sqrt", ErrInvalidExpression, 4},
		{"sq", ErrInvalidExpression, 2},
		{"SQRT(4)", ErrInvalidExpression, 0},

		// Numbers and values.
		{".5", calc.ErrInvalidNumber, 0},
		{"5.", calc.ErrInvalidNumber, 0},
		{"2e", calc.ErrInvalidNumber, 0},
		{"2e+", calc.ErrInvalidNumber, 0},
		{"1+2E+-3", calc.ErrInvalidNumber, 2},
		{"1e5e5", calc.ErrInvalidNumber, 0},
		{"[ 1 ]", calc.ErrInvalidNumber, 0},
		{"[]", calc.ErrInvalidNumber, 0},
		{"2*[1/0]", calc.ErrInvalidNumber, 2},
		{"[+1]", calc.ErrInvalidNumber, 0},
		{"[1,5]", calc.ErrInvalidNumber, 0},
		{"[~1/3]", calc.ErrInvalidNumber, 0},
	}
	for _, tt := range tests {
		_, err := parse(tt.input)
		checkError(t, tt.input, err, tt.wantErr, tt.wantPos)
	}
}

// repeat joins n copies of s with sep.
func repeat(s, sep string, n int) string {
	return strings.TrimSuffix(strings.Repeat(s+sep, n), sep)
}

func TestParseLimits(t *testing.T) {
	nested := func(open string, n int) string {
		return strings.Repeat(open, n) + "1" + strings.Repeat(")", n)
	}
	tests := []struct {
		name    string
		input   string
		wantErr error
		wantPos int
	}{
		{"100 operations", repeat("1", "+", 101), nil, 0},
		{"101 operations", repeat("1", "+", 102), ErrExpressionTooComplex, 201},
		{"10 powers", repeat("1", "^", 11), nil, 0},
		{"11 powers", repeat("1", "^", 12), ErrExpressionTooComplex, 21},
		{"unary minus counts", strings.Repeat("-", 100) + "1", nil, 0},
		{"101 unary minus", strings.Repeat("-", 101) + "1", ErrExpressionTooComplex, 100},
		{"sqrt counts", nested("sqrt(", 100), nil, 0},
		{"percent counts", "1" + strings.Repeat("+1%", 50), nil, 0},
		{"a second percent is a syntax error, before the count", "1" + strings.Repeat("+1%", 50) + "%", ErrInvalidExpression, 151},
		{"101st operation is a percent", "-1" + strings.Repeat("+1%", 50), ErrExpressionTooComplex, 151},
		// D54: the earlier of the 101st operation and the 11th power wins.
		{"power limit first", repeat("1", "^", 12) + strings.Repeat("+1", 90), ErrExpressionTooComplex, 21},
		{"operation limit first", strings.Repeat("1+", 100) + repeat("1", "^", 12), ErrExpressionTooComplex, 201},
		{"100 levels of parentheses", nested("(", 100), nil, 0},
		{"101 levels of parentheses", nested("(", 101), ErrExpressionTooComplex, 100},
		{"sqrt at level 101", "(" + strings.Repeat("(", 99) + "sqrt(1)" + strings.Repeat(")", 100), ErrExpressionTooComplex, 100},
		{"levels reset after closing", nested("(", 100) + "+" + nested("(", 100), nil, 0},
		{"nesting is a syntax error, met before a later one", nested("(", 101) + ")", ErrExpressionTooComplex, 100},
		{"a syntax error before the nesting", "2 3" + nested("(", 101), ErrInvalidExpression, 2},
	}
	for _, tt := range tests {
		_, err := parse(tt.input)
		if tt.wantErr == nil {
			if err != nil {
				t.Errorf("%s: error = %v", tt.name, err)
			}
			continue
		}
		checkError(t, tt.name, err, tt.wantErr, tt.wantPos)
	}
}

func TestParseOperandLimits(t *testing.T) {
	tests := []struct {
		input   string
		wantErr error
		wantPos int
	}{
		{"1e40", calc.ErrOperandOutOfRange, 0},
		{"2+[1e40]", calc.ErrOperandOutOfRange, 2},
		{"1e-1001*1e40", calc.ErrOperandOutOfRange, 0},
		{"1+1e40+1e41", calc.ErrOperandOutOfRange, 2},
		// The syntax is checked first, even after an out-of-range operand.
		{"1e40+", ErrInvalidExpression, 5},
		{"1e40+1.2.3", calc.ErrInvalidNumber, 5},
		// Operand limits come before the number of operations.
		{repeat("1", "^", 12) + "+1e40", calc.ErrOperandOutOfRange, 24},
	}
	for _, tt := range tests {
		_, err := parse(tt.input)
		checkError(t, tt.input, err, tt.wantErr, tt.wantPos)
	}
}

func TestErrorMessage(t *testing.T) {
	err := &Error{Err: ErrInvalidExpression, Position: 3}
	if got, want := err.Error(), "invalid expression at position 3"; got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}
