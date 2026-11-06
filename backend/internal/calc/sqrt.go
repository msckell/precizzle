package calc

import "math/big"

// Sqrt returns the square root of x. The root of a rational is rational only
// when its reduced numerator and denominator are perfect squares. Any other
// root is irrational, and it is returned correctly rounded as approximate.
func Sqrt(x Number) (Result, error) {
	if x.rat.Sign() < 0 {
		return Result{}, ErrNegativeSquareRoot
	}
	p, q := x.rat.Num(), x.rat.Denom()
	rootP, rootQ := new(big.Int).Sqrt(p), new(big.Int).Sqrt(q)
	if isSquareOf(p, rootP) && isSquareOf(q, rootQ) {
		return newResult(Number{rat: new(big.Rat).SetFrac(rootP, rootQ), approximate: x.approximate})
	}

	// ⌊log10 √x⌋ = ⌊⌊log10 x⌋ / 2⌋, and >> 1 is a floor division also for negative numbers.
	f := place(magnitude(p, q)>>1, approximateDigits, true)
	scaled := new(big.Int).Mul(p, pow10(2*f))
	// s = ⌊√(x·10^2f)⌋, since ⌊√⌊y⌋⌋ = ⌊√y⌋.
	s := new(big.Int).Sqrt(new(big.Int).Quo(scaled, q))
	// Round up when s + 1/2 < √(x·10^2f), that is (2s+1)²·q < 4·p·10^2f.
	// The root is irrational, so it is never exactly halfway.
	t := new(big.Int).Add(new(big.Int).Lsh(s, 1), one)
	if t.Mul(t, t).Mul(t, q).Cmp(scaled.Lsh(scaled, 2)) < 0 {
		s.Add(s, one)
	}
	d := decimalString(s, f)
	return Result{Value: "~" + d, Decimal: d, Kind: Approximate}, nil
}

func isSquareOf(n, root *big.Int) bool {
	return new(big.Int).Mul(root, root).Cmp(n) == 0
}
