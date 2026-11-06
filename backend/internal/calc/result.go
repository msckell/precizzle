package calc

import (
	"math/big"
	"strconv"
	"strings"
)

// Kind tells whether a result is a terminating decimal, an exact repeating fraction or an approximation.
type Kind int

// The kinds of result defined in docs/SPEC.md.
const (
	Exact Kind = iota
	Repeating
	Approximate
)

// String returns the name of the kind used in API responses.
func (k Kind) String() string {
	return [...]string{"exact", "repeating", "approximate"}[k]
}

// Result is a calculation result: the value used for chaining, the decimal used for display, and its kind.
type Result struct {
	Value   string
	Decimal string
	Kind    Kind
}

var (
	one            = big.NewInt(1)
	ten            = big.NewInt(10)
	maxMagnitude   = pow10(maxIntegerDigits)
	maxDenominator = pow10(maxFractionDigits)
)

func pow10(n int) *big.Int {
	return new(big.Int).Exp(ten, big.NewInt(int64(n)), nil)
}

func newResult(x Number) (Result, error) {
	return result(x.rat.Num(), x.rat.Denom(), x.approximate)
}

// result checks num/den against the limits and writes it in canonical form.
// den is positive, and for exact values num/den is in lowest terms.
func result(num, den *big.Int, approximate bool) (Result, error) {
	if err := checkLimits(num, den, approximate); err != nil {
		return Result{}, err
	}
	if approximate {
		d := decimal(num, den, approximateDigits, true)
		return Result{Value: "~" + d, Decimal: d, Kind: Approximate}, nil
	}
	r := new(big.Rat).SetFrac(num, den)
	n, terminates := r.FloatPrec()
	if !terminates {
		return Result{Value: r.String(), Decimal: decimal(num, den, repeatingDigits, false), Kind: Repeating}, nil
	}
	scaled := new(big.Int).Quo(new(big.Int).Mul(num, pow10(n)), den)
	value := scaledString(scaled, n, false)
	if num.Sign() != 0 && magnitude(num, den) < scientificBelow {
		return Result{Value: value, Decimal: scaledString(scaled, n, true), Kind: Exact}, nil
	}
	return Result{Value: value, Decimal: value, Kind: Exact}, nil
}

// checkLimits applies the limits of docs/SPEC.md. den is positive, and for
// exact values num/den is in lowest terms.
func checkLimits(num, den *big.Int, approximate bool) error {
	abs := new(big.Int).Abs(num)
	if abs.Cmp(new(big.Int).Mul(den, maxMagnitude)) >= 0 {
		return ErrResultTooLarge
	}
	if approximate {
		// An approximate value is rounded at the 1000th fractional digit at the
		// latest. If it would round to 0 there, |x| ≤ 10^-1000 / 2, it is
		// rejected rather than turned into 0.
		twice := new(big.Int).Lsh(abs.Mul(abs, maxDenominator), 1)
		if num.Sign() != 0 && twice.Cmp(den) <= 0 {
			return ErrResultTooPrecise
		}
		return nil
	}
	// A terminating decimal with at most 1000 fractional digits has a
	// denominator of at most 10^1000, so this check covers both exact kinds.
	if den.Cmp(maxDenominator) > 0 {
		return ErrResultTooPrecise
	}
	if n, terminates := new(big.Rat).SetFrac(num, den).FloatPrec(); terminates && n > maxFractionDigits {
		return ErrResultTooPrecise
	}
	return nil
}

// decimal writes a non-terminating num/den rounded half even to the given
// number of fractional digits, or of mantissa digits in scientific notation.
func decimal(num, den *big.Int, digits int, approximate bool) string {
	if num.Sign() == 0 {
		return "0"
	}
	f, scientific := place(magnitude(num, den), digits, approximate)
	return scaledString(roundHalfEven(num, den, f), f, scientific)
}

// place returns the fractional digit at which a value of magnitude e is
// rounded, and whether it is written in scientific notation. Approximate
// values never go past the 1000th fractional digit.
func place(e, digits int, approximate bool) (f int, scientific bool) {
	if e >= scientificBelow {
		return digits, false
	}
	f = digits - e
	if approximate {
		f = min(f, maxFractionDigits)
	}
	return f, true
}

// magnitude returns ⌊log10 |num/den|⌋ for a non-zero num.
func magnitude(num, den *big.Int) int {
	abs := new(big.Int).Abs(num)
	// log10(2) ≈ 30103/100000 gives an estimate that the loops correct exactly.
	e := int(int64(abs.BitLen()-den.BitLen()) * 30103 / 100000)
	for compareToPow10(abs, den, e) < 0 {
		e--
	}
	for compareToPow10(abs, den, e+1) >= 0 {
		e++
	}
	return e
}

// compareToPow10 compares abs/den with 10^e.
func compareToPow10(abs, den *big.Int, e int) int {
	if e >= 0 {
		return abs.Cmp(new(big.Int).Mul(den, pow10(e)))
	}
	return new(big.Int).Mul(abs, pow10(-e)).Cmp(den)
}

// roundHalfEven returns num/den × 10^f rounded to the nearest integer, ties to even.
func roundHalfEven(num, den *big.Int, f int) *big.Int {
	q, r := new(big.Int).QuoRem(new(big.Int).Mul(new(big.Int).Abs(num), pow10(f)), den, new(big.Int))
	switch r.Lsh(r, 1).Cmp(den) {
	case 1:
		q.Add(q, one)
	case 0:
		if q.Bit(0) == 1 {
			q.Add(q, one)
		}
	}
	if num.Sign() < 0 {
		q.Neg(q)
	}
	return q
}

// scaledString writes n × 10^-f without trailing fractional zeros, in plain or scientific notation.
func scaledString(n *big.Int, f int, scientific bool) string {
	digits := new(big.Int).Abs(n).String()
	var s string
	if scientific {
		s = digits[:1]
		if mantissa := strings.TrimRight(digits[1:], "0"); mantissa != "" {
			s += "." + mantissa
		}
		s += "e" + strconv.Itoa(len(digits)-1-f)
	} else {
		if len(digits) <= f {
			digits = strings.Repeat("0", f-len(digits)+1) + digits
		}
		s = digits[:len(digits)-f]
		if fraction := strings.TrimRight(digits[len(digits)-f:], "0"); fraction != "" {
			s += "." + fraction
		}
	}
	if n.Sign() < 0 {
		s = "-" + s
	}
	return s
}
