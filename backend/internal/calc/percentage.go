package calc

import "math/big"

// PercentOperator is the operator that precedes a percentage, as the + in 200 + 15%.
type PercentOperator int

// The operators that can precede a percentage.
const (
	PercentAdd PercentOperator = iota
	PercentSubtract
	PercentMultiply
	PercentDivide
)

var hundred = big.NewRat(100, 1)

// Percentage returns percent ÷ 100, so 15% is 0.15.
func Percentage(percent Number) (Result, error) {
	return newResult(derived(new(big.Rat).Quo(percent.rat, hundred), percent))
}

// ApplyPercentage applies percent to base with the pocket calculator rules of
// D7: 200 + 15% = 230, 200 - 15% = 170, 200 × 15% = 30 and 200 ÷ 15% = 4000/3.
func ApplyPercentage(base Number, op PercentOperator, percent Number) (Result, error) {
	fraction := new(big.Rat).Quo(percent.rat, hundred)
	share := new(big.Rat).Mul(base.rat, fraction)
	var r *big.Rat
	switch op {
	case PercentAdd:
		r = share.Add(base.rat, share)
	case PercentSubtract:
		r = share.Sub(base.rat, share)
	case PercentMultiply:
		r = share
	case PercentDivide:
		if fraction.Sign() == 0 {
			return Result{}, ErrDivisionByZero
		}
		r = new(big.Rat).Quo(base.rat, fraction)
	}
	return newResult(derived(r, base, percent))
}
