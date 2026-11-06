package calc

import (
	"errors"
	"math/big"
	"strings"
	"testing"
)

func TestPower(t *testing.T) {
	fiveTo1000 := new(big.Int).Exp(big.NewInt(5), big.NewInt(1000), nil).String()
	twoToMinus1000 := Result{
		Value:   "0." + strings.Repeat("0", 1000-len(fiveTo1000)) + fiveTo1000,
		Decimal: fiveTo1000[:1] + "." + strings.TrimRight(fiveTo1000[1:], "0") + "e-302",
		Kind:    Exact,
	}
	oneAnd1000Digits := "1." + strings.Repeat("0", 999) + "1"

	runBinary(t, Power, []binaryCase{
		{name: "2 ^ -2", a: "2", b: "-2", want: Result{"0.25", "0.25", Exact}},
		{name: "3 ^ -1", a: "3", b: "-1", want: Result{"1/3", "0." + strings.Repeat("3", 100), Repeating}},
		{name: "0 ^ 0", a: "0", b: "0", want: Result{"1", "1", Exact}},
		{name: "0 ^ 5", a: "0", b: "5", want: Result{"0", "0", Exact}},
		{name: "0 ^ -1", a: "0", b: "-1", wantErr: ErrDivisionByZero},
		{name: "approximate 0 ^ 0", a: "~0", b: "0", want: Result{"~1", "1", Approximate}},
		{name: "negative base, odd exponent", a: "-2", b: "3", want: Result{"-8", "-8", Exact}},
		{name: "negative base, even exponent", a: "-2", b: "2", want: Result{"4", "4", Exact}},
		{name: "negative fraction, negative exponent", a: "-1/2", b: "-3", want: Result{"-8", "-8", Exact}},
		{name: "(1/3) ^ -2", a: "1/3", b: "-2", want: Result{"9", "9", Exact}},
		{name: "exponent 2.0", a: "2", b: "2.0", want: Result{"4", "4", Exact}},
		{name: "exponent 4/2", a: "2", b: "4/2", want: Result{"4", "4", Exact}},
		{name: "exponent 1e2", a: "2", b: "1e2", want: Result{"1267650600228229401496703205376", "1267650600228229401496703205376", Exact}},
		{name: "exponent 2.5", a: "2", b: "2.5", wantErr: ErrExponentNotInteger},
		{name: "exponent 1/2", a: "4", b: "1/2", wantErr: ErrExponentNotInteger},
		{name: "approximate exponent", a: "2", b: "~2", wantErr: ErrExponentNotInteger},
		{name: "exponent 1000", a: "1", b: "1000", want: Result{"1", "1", Exact}},
		{name: "exponent -1000", a: "-1", b: "-1000", want: Result{"1", "1", Exact}},
		{name: "2 ^ 1001", a: "2", b: "1001", wantErr: ErrExponentOutOfRange},
		{name: "2 ^ -1001", a: "2", b: "-1001", wantErr: ErrExponentOutOfRange},
		{name: "1.0001 ^ -2000", a: "1.0001", b: "-2000", wantErr: ErrExponentOutOfRange},
		{name: "exponent 1e39", a: "1", b: "1e39", wantErr: ErrExponentOutOfRange},
		{name: "exponent range wins over division by zero", a: "0", b: "-1001", wantErr: ErrExponentOutOfRange},
		{name: "2 ^ 132", a: "2", b: "132", want: Result{"5444517870735015415413993718908291383296", "5444517870735015415413993718908291383296", Exact}},
		{name: "2 ^ 133", a: "2", b: "133", wantErr: ErrResultTooLarge},
		{name: "10 ^ 40", a: "10", b: "40", wantErr: ErrResultTooLarge},
		{name: "1e10 ^ 4", a: "1e10", b: "4", wantErr: ErrResultTooLarge},
		{name: "9999999999 ^ 4", a: "9999999999", b: "4", want: Result{"9999999996000000000599999999960000000001", "9999999996000000000599999999960000000001", Exact}},
		{name: "1e-20 ^ -2", a: "1e-20", b: "-2", wantErr: ErrResultTooLarge},
		{name: "-10 ^ 41", a: "-10", b: "41", wantErr: ErrResultTooLarge},
		{name: "just above 10^(40/3) cubed", a: "21544346900318.83721760", b: "3", wantErr: ErrResultTooLarge},
		{name: "just below 10^(40/3) cubed", a: "21544346900318.83721759", b: "3", want: Result{
			"9999999999999999999995912154765580514273.886956099150079026448479",
			"9999999999999999999995912154765580514273.886956099150079026448479",
			Exact,
		}},
		{name: "too close to 10^40 for the bounds, above", a: "21544346900318.8372175929356651935050", b: "3", wantErr: ErrResultTooLarge},
		{name: "too close to 10^40 for the bounds, below", a: "21544346900318.8372175929356651935049", b: "3", want: Result{
			"9999999999999999999999999999999999926764.849732609788082946463401779497140587063003493248915650239962922649",
			"9999999999999999999999999999999999926764.849732609788082946463401779497140587063003493248915650239962922649",
			Exact,
		}},
		{name: "0.5 ^ 1000", a: "0.5", b: "1000", want: twoToMinus1000},
		{name: "2 ^ -1000", a: "2", b: "-1000", want: twoToMinus1000},
		{name: "0.25 ^ 501", a: "0.25", b: "501", wantErr: ErrResultTooPrecise},
		{name: "(1/11) ^ 1000", a: "1/11", b: "1000", wantErr: ErrResultTooPrecise},
		{name: "exact base near 1 with 1000 digits", a: oneAnd1000Digits, b: "1000", wantErr: ErrResultTooPrecise},
		{name: "approximate base", a: "~2", b: "3", want: Result{"~8", "8", Approximate}},
		{name: "approximate base, negative exponent", a: "~1.5", b: "-2", want: Result{"~0.44444444444444444444", "0.44444444444444444444", Approximate}},
		{name: "approximate base near 1 with 1000 digits", a: "~" + oneAnd1000Digits, b: "1000", want: Result{"~1", "1", Approximate}},
		{name: "approximate 1e-500 squared", a: "~1e-500", b: "2", want: Result{"~1e-1000", "1e-1000", Approximate}},
		{name: "approximate 1e-500 cubed", a: "~1e-500", b: "3", wantErr: ErrResultTooPrecise},
		{name: "approximate 1e-500 ^ 1000", a: "~1e-500", b: "1000", wantErr: ErrResultTooPrecise},
		{name: "approximate 0.1 ^ 1000", a: "~0.1", b: "1000", want: Result{"~1e-1000", "1e-1000", Approximate}},
		{name: "approximate 0.1 ^ -1000", a: "~0.1", b: "-1000", wantErr: ErrResultTooLarge},
	})
}

func TestPowerOfAThirdKeepsTheExactFraction(t *testing.T) {
	den := new(big.Int).Exp(big.NewInt(3), big.NewInt(1000), nil)
	pow10 := func(n int64) *big.Int { return new(big.Int).Exp(big.NewInt(10), big.NewInt(n), nil) }

	// The expected decimal is built here from its definition, independently of
	// the rounding code. 10^477 < 3^1000 < 10^478 puts 3^-1000 between 10^-478
	// and 10^-477, so its 100-digit mantissa is 3^-1000 × 10^578 rounded half
	// even. That equals the truncated quotient when the remainder is below half
	// the divisor, which the test checks instead of assuming.
	if den.Cmp(pow10(477)) <= 0 || den.Cmp(pow10(478)) >= 0 {
		t.Fatalf("3^1000 is not between 10^477 and 10^478")
	}
	q, r := new(big.Int).QuoRem(pow10(578), den, new(big.Int))
	if r.Lsh(r, 1).Cmp(den) >= 0 {
		t.Fatalf("the remainder is not below half the divisor, so truncating is not rounding")
	}
	digits := q.String()
	if len(digits) != 101 {
		t.Fatalf("the mantissa has %d digits, want 101", len(digits))
	}
	want := Result{
		Value:   "1/" + den.String(),
		Decimal: digits[:1] + "." + strings.TrimRight(digits[1:], "0") + "e-478",
		Kind:    Repeating,
	}

	got, err := Power(num(t, "1/3"), num(t, "1000"))
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if got != want {
		t.Errorf("(1/3) ^ 1000 = %+v, want %+v", got, want)
	}
}

func TestBitsAboveMaxDenominator(t *testing.T) {
	above := new(big.Int).Lsh(one, bitsAboveMaxDenominator)
	below := new(big.Int).Lsh(one, bitsAboveMaxDenominator-1)
	if above.Cmp(maxDenominator) <= 0 || below.Cmp(maxDenominator) >= 0 {
		t.Errorf("2^%d must be the smallest power of 2 above 10^1000", bitsAboveMaxDenominator)
	}
}

// floorRoot returns ⌊y^(1/n)⌋ for y > 0, with Newton's method on integers.
func floorRoot(y *big.Int, n int64) *big.Int {
	bn := big.NewInt(n)
	x := new(big.Int).Lsh(one, uint(y.BitLen())/uint(n)+1)
	for {
		// next = ((n-1)·x + y / x^(n-1)) / n, which decreases toward the root from above.
		next := new(big.Int).Quo(y, new(big.Int).Exp(x, big.NewInt(n-1), nil))
		next.Add(next, new(big.Int).Mul(x, big.NewInt(n-1))).Quo(next, bn)
		if next.Cmp(x) >= 0 {
			return x
		}
		x = next
	}
}

// BenchmarkPowerWorstCase measures the costliest powers. The "rejected early"
// cases are refused by powerBounds; the "full power" cases compute numbers of
// millions of bits. Each case checks its path and its result before timing.
func BenchmarkPowerWorstCase(b *testing.B) {
	oneAnd1000Digits := "1." + strings.Repeat("0", 999) + "1"
	parse := func(s string) Number {
		n, err := Parse(s)
		if err != nil {
			b.Fatal(err)
		}
		return n
	}
	// 10^(1/25) raised to 1000 is 10^40. Its 1000-digit truncation
	// ⌊10^(1/25) × 10^1000⌋ / 10^1000 = ⌊(10^25001)^(1/25)⌋ / 10^1000 raised to
	// 1000 lands just below 10^40: too close for the bounds to decide, so the
	// full power is computed.
	root := floorRoot(new(big.Int).Exp(big.NewInt(10), big.NewInt(1+25*1000), nil), 25)
	nearBoundary := Number{rat: new(big.Rat).SetFrac(root, new(big.Int).Exp(big.NewInt(10), big.NewInt(1000), nil))}

	tests := []struct {
		name      string
		base      Number
		exponent  Number
		fullPower bool
		want      Result
		wantErr   error
	}{
		{"exact, rejected early", parse(oneAnd1000Digits), parse("1000"), false, Result{}, ErrResultTooPrecise},
		{"exact, full power", nearBoundary, parse("1000"), true, Result{}, ErrResultTooPrecise},
		{"approximate, rejected early", parse("~1e-500"), parse("1000"), false, Result{}, ErrResultTooPrecise},
		{"approximate, full power", parse("~" + oneAnd1000Digits), parse("1000"), true, Result{"~1", "1", Approximate}, nil},
	}
	for _, tt := range tests {
		k := tt.exponent.rat.Num().Int64()
		early := powerBounds(tt.base.rat.Num(), tt.base.rat.Denom(), k, tt.base.approximate) != nil
		if early == tt.fullPower {
			b.Fatalf("%s: rejected early = %v, want %v", tt.name, early, !tt.fullPower)
		}
		if got, err := Power(tt.base, tt.exponent); got != tt.want || !errors.Is(err, tt.wantErr) {
			b.Fatalf("%s: Power = %+v, %v, want %+v, %v", tt.name, got, err, tt.want, tt.wantErr)
		}
		b.Run(tt.name, func(b *testing.B) {
			for b.Loop() {
				_, _ = Power(tt.base, tt.exponent)
			}
		})
	}
}
