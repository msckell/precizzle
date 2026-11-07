// Package calc implements the number model of docs/SPEC.md: exact rational
// arithmetic, approximate values, their limits and their canonical text forms.
package calc

import (
	"math/big"
	"strconv"
	"strings"
)

const (
	maxIntegerDigits  = 40
	maxFractionDigits = 1000
	repeatingDigits   = 100
	approximateDigits = 20
	// Values whose magnitude is below 10^scientificBelow are written in scientific notation.
	scientificBelow = -20
	// Exponents with more digits than this are out of range for any non-zero mantissa.
	maxExponentDigits = 9
)

// Number is an operand: an exact rational, marked approximate when its digits come from a rounded result.
type Number struct {
	rat         *big.Rat
	approximate bool
}

// Parse reads an operand written in the operand syntax of docs/SPEC.md.
func Parse(s string) (Number, error) {
	digits, approximate := strings.CutPrefix(s, "~")
	var r *big.Rat
	var err error
	if num, den, ok := strings.Cut(digits, "/"); ok && !approximate {
		r, err = parseFraction(num, den)
	} else {
		r, err = parseDecimal(digits)
	}
	if err != nil {
		return Number{}, err
	}
	if checkLimits(r.Num(), r.Denom(), approximate) != nil {
		return Number{}, ErrOperandOutOfRange
	}
	return Number{rat: r, approximate: approximate}, nil
}

// Result returns the operand in canonical form, as an expression with no operation returns it.
func (n Number) Result() Result {
	// Parse accepts only values within the limits, so the result cannot fail.
	r, _ := newResult(n)
	return r
}

func parseFraction(num, den string) (*big.Rat, error) {
	num, negative := strings.CutPrefix(num, "-")
	if !isDigits(num) || !isDigits(den) {
		return nil, ErrInvalidNumber
	}
	q := parseInt(den)
	if q.Sign() == 0 {
		return nil, ErrInvalidNumber
	}
	p := parseInt(num)
	if negative {
		p.Neg(p)
	}
	return new(big.Rat).SetFrac(p, q), nil
}

// parseDecimal reads [-]digits[.digits][(e|E)[+|-]digits]. It checks the
// magnitude from the digit counts before building the value, so an input such
// as 1e999999999 is rejected without allocating a billion-digit number.
func parseDecimal(s string) (*big.Rat, error) {
	s, negative := strings.CutPrefix(s, "-")
	whole, s := leadingDigits(s)
	if whole == "" {
		return nil, ErrInvalidNumber
	}
	var fraction string
	if rest, ok := strings.CutPrefix(s, "."); ok {
		if fraction, s = leadingDigits(rest); fraction == "" {
			return nil, ErrInvalidNumber
		}
	}
	var exponent string
	negativeExponent := false
	if len(s) > 0 && (s[0] == 'e' || s[0] == 'E') {
		s = s[1:]
		if len(s) > 0 && (s[0] == '+' || s[0] == '-') {
			negativeExponent = s[0] == '-'
			s = s[1:]
		}
		if exponent, s = leadingDigits(s); exponent == "" {
			return nil, ErrInvalidNumber
		}
	}
	if s != "" {
		return nil, ErrInvalidNumber
	}

	significant := strings.TrimLeft(whole+fraction, "0")
	if significant == "" {
		return new(big.Rat), nil
	}
	trimmed := strings.TrimRight(significant, "0")
	exponent = strings.TrimLeft(exponent, "0")
	if len(exponent) > maxExponentDigits {
		return nil, ErrOperandOutOfRange
	}
	e := 0
	if exponent != "" {
		e, _ = strconv.Atoi(exponent)
	}
	if negativeExponent {
		e = -e
	}
	// The value is trimmed × 10^scale, and its leading digit sits at 10^(scale+len(trimmed)-1).
	scale := e - len(fraction) + len(significant) - len(trimmed)
	if scale+len(trimmed) > maxIntegerDigits || scale < -maxFractionDigits {
		return nil, ErrOperandOutOfRange
	}

	m := parseInt(trimmed)
	if negative {
		m.Neg(m)
	}
	if scale >= 0 {
		return new(big.Rat).SetInt(m.Mul(m, pow10(scale))), nil
	}
	return new(big.Rat).SetFrac(m, pow10(-scale)), nil
}

func leadingDigits(s string) (digits, rest string) {
	i := 0
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		i++
	}
	return s[:i], s[i:]
}

func isDigits(s string) bool {
	digits, rest := leadingDigits(s)
	return digits != "" && rest == ""
}

// parseInt reads a string that is known to contain only ASCII digits.
func parseInt(s string) *big.Int {
	n, _ := new(big.Int).SetString(s, 10)
	return n
}
