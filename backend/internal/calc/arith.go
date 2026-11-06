package calc

import "math/big"

// Add returns a + b.
func Add(a, b Number) (Result, error) {
	return newResult(derived(new(big.Rat).Add(a.rat, b.rat), a, b))
}

// Subtract returns a - b.
func Subtract(a, b Number) (Result, error) {
	return newResult(derived(new(big.Rat).Sub(a.rat, b.rat), a, b))
}

// Multiply returns a × b.
func Multiply(a, b Number) (Result, error) {
	return newResult(derived(new(big.Rat).Mul(a.rat, b.rat), a, b))
}

// Divide returns a ÷ b, or ErrDivisionByZero when b is zero.
func Divide(a, b Number) (Result, error) {
	if b.rat.Sign() == 0 {
		return Result{}, ErrDivisionByZero
	}
	return newResult(derived(new(big.Rat).Quo(a.rat, b.rat), a, b))
}

// derived wraps a value computed from the operands. It is approximate when any operand is.
func derived(r *big.Rat, operands ...Number) Number {
	approximate := false
	for _, x := range operands {
		approximate = approximate || x.approximate
	}
	return Number{rat: r, approximate: approximate}
}
