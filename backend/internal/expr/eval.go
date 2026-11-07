package expr

import (
	"errors"
	"fmt"

	"github.com/msckell/precizzle/backend/internal/calc"
)

// binaryOps maps each binary operator to the calc operation that computes it.
var binaryOps = map[byte]func(a, b calc.Number) (calc.Result, error){
	'+': calc.Add,
	'-': calc.Subtract,
	'*': calc.Multiply,
	'/': calc.Divide,
	'^': calc.Power,
}

var zero, _ = calc.Parse("0")

// Evaluate computes an expression with the rules of section 5 of
// docs/SPEC.md. An error about the expression or one of its operations is an
// *Error; any other error is a bug.
func Evaluate(s string) (calc.Result, error) {
	tree, err := parse(s)
	if err != nil {
		return calc.Result{}, err
	}
	n, err := eval(tree)
	if err != nil {
		return calc.Result{}, err
	}
	return n.Result(), nil
}

// eval computes a node from left to right: each operation runs as soon as its operands are known.
func eval(n node) (calc.Number, error) {
	switch n := n.(type) {
	case *operand:
		return n.value, nil
	case *unary:
		x, err := eval(n.x)
		if err != nil {
			return x, err
		}
		switch n.op {
		case '-':
			return negate(x, n.pos)
		case '%':
			r, err := calc.Percentage(x)
			return next(r, err, n.pos)
		default:
			r, err := calc.Sqrt(x)
			return next(r, err, n.pos)
		}
	case *binary:
		x, err := eval(n.x)
		if err != nil {
			return x, err
		}
		y, err := eval(n.y)
		if err != nil {
			return y, err
		}
		r, err := binaryOps[n.op](x, y)
		return next(r, err, n.pos)
	case *applied:
		x, err := eval(n.x)
		if err != nil {
			return x, err
		}
		return applyPercent(x, n.p)
	default:
		return evalSum(n.(*sum))
	}
}

// evalSum adds and subtracts the terms of a sum from left to right, each one
// after its percentages are applied (D43).
func evalSum(s *sum) (calc.Number, error) {
	total, err := evalTerm(s.terms[0])
	for _, t := range s.terms[1:] {
		if err != nil {
			break
		}
		var x calc.Number
		if x, err = evalTerm(t); err == nil {
			r, opErr := binaryOps[t.op](total, x)
			total, err = next(r, opErr, t.pos)
		}
	}
	return total, err
}

// evalTerm computes a term of a sum and applies the percentages that follow it.
func evalTerm(t sumTerm) (calc.Number, error) {
	x, err := eval(t.x)
	for _, p := range t.percents {
		if err != nil {
			break
		}
		x, err = applyPercent(x, p)
	}
	return x, err
}

// percentOps maps the operator in front of a percentage to its D7 rule.
var percentOps = map[byte]calc.PercentOperator{
	'+': calc.PercentAdd,
	'-': calc.PercentSubtract,
	'*': calc.PercentMultiply,
	'/': calc.PercentDivide,
}

// applyPercent computes A + B%, A - B%, A × B% or A ÷ B% as one operation
// with the D7 rules, where the percent B carries the minus signs written in
// front of it (D43, D55).
func applyPercent(a calc.Number, p percent) (calc.Number, error) {
	b, err := eval(p.base)
	for k := len(p.negs) - 1; k >= 0 && err == nil; k-- {
		b, err = negate(b, p.negs[k])
	}
	if err != nil {
		return b, err
	}
	r, err := calc.ApplyPercentage(a, percentOps[p.op], b)
	if errors.Is(err, calc.ErrDivisionByZero) {
		return next(r, err, p.opPos)
	}
	return next(r, err, p.pos)
}

// negate computes -x as 0 - x, which keeps the approximate mark and the limits.
func negate(x calc.Number, pos int) (calc.Number, error) {
	r, err := calc.Subtract(zero, x)
	return next(r, err, pos)
}

// next turns the result of the operation at pos into the operand of the next
// one. An intermediate result enters as its value, so an approximate result
// keeps only its rounded digits, as when two steps are chained (D44).
func next(r calc.Result, err error, pos int) (calc.Number, error) {
	if err != nil {
		return calc.Number{}, &Error{err, pos}
	}
	n, err := calc.Parse(r.Value)
	if err != nil {
		// The value of a result is always a valid operand, so this is a bug.
		// %v keeps the calc error out of the chain, so it is not mistaken for
		// a problem with the expression.
		return calc.Number{}, fmt.Errorf("expr: the value %q of an operation does not parse: %v", r.Value, err)
	}
	return n, nil
}
