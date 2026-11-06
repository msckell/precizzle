package calc

import "math/big"

const (
	maxExponent = 1000
	// 2^3322 is the smallest power of 2 above 10^1000.
	bitsAboveMaxDenominator = 3322
	// Extra bits of precision for the bounds that powerBounds computes.
	boundBits = 64
)

// Power returns base raised to exponent, which must be an exact integer
// between -1000 and 1000 (D8, D18). 0^0 is 1.
func Power(base, exponent Number) (Result, error) {
	if exponent.approximate || !exponent.rat.IsInt() {
		return Result{}, ErrExponentNotInteger
	}
	if exponent.rat.Num().CmpAbs(big.NewInt(maxExponent)) > 0 {
		return Result{}, ErrExponentOutOfRange
	}
	k := exponent.rat.Num().Int64()
	num, den := base.rat.Num(), base.rat.Denom()
	if k < 0 {
		if num.Sign() == 0 {
			return Result{}, ErrDivisionByZero
		}
		// (p/q)^-k = (q/p)^k, with the sign kept on the numerator.
		num, den = new(big.Int).Set(den), new(big.Int).Abs(num)
		if base.rat.Sign() < 0 {
			num.Neg(num)
		}
		k = -k
	}
	if err := powerBounds(num, den, k, base.approximate); err != nil {
		return Result{}, err
	}
	e := big.NewInt(k)
	// num and den are coprime, so their powers are coprime too and the
	// result needs no reduction.
	return result(new(big.Int).Exp(num, e, nil), new(big.Int).Exp(den, e, nil), base.approximate)
}

// powerBounds rejects |num/den|^k early when it is certainly outside the
// limits, so the exact power, which can reach millions of bits, is computed
// only when the result may be valid. It brackets |num/den| between
// lo/2^64 and (lo+1)/2^64, whose powers are cheap to compute.
func powerBounds(num, den *big.Int, k int64, approximate bool) error {
	e := big.NewInt(k)
	lo := new(big.Int).Quo(new(big.Int).Lsh(new(big.Int).Abs(num), boundBits), den)
	hiPower := new(big.Int).Exp(new(big.Int).Add(lo, one), e, nil)
	unit := new(big.Int).Lsh(one, uint(boundBits*k))
	limit := new(big.Int).Mul(maxMagnitude, unit)
	if new(big.Int).Exp(lo, e, nil).Cmp(limit) >= 0 {
		return ErrResultTooLarge
	}
	if hiPower.Cmp(limit) > 0 {
		// The result is too close to 10^40 to decide here, so it must be computed.
		return nil
	}
	if approximate {
		// |x|^k < hiPower/unit, so it certainly rounds to 0 at the 1000th
		// fractional digit when hiPower/unit ≤ 10^-1000 / 2.
		if new(big.Int).Lsh(hiPower.Mul(hiPower, maxDenominator), 1).Cmp(unit) <= 0 {
			return ErrResultTooPrecise
		}
		return nil
	}
	// den^k ≥ 2^((bits-1)·k), which exceeds 10^1000 from 3322 bits on.
	if int64(den.BitLen()-1)*k >= bitsAboveMaxDenominator {
		return ErrResultTooPrecise
	}
	return nil
}
