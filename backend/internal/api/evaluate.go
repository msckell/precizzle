package api

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/msckell/precizzle/backend/internal/calc"
	"github.com/msckell/precizzle/backend/internal/expr"
)

// expressionFailures maps each cause of an expr.Error to its code and to the
// format of its detail, which receives the position.
var expressionFailures = []struct {
	err    error
	code   string
	detail string
}{
	{expr.ErrInvalidExpression, codeInvalidExpression, "The expression cannot continue at position %d."},
	{expr.ErrExpressionTooComplex, codeExpressionTooComplex, "The expression is too complex at position %d: it allows at most 100 operations, 10 powers and 100 nested levels."},
	{calc.ErrInvalidNumber, codeInvalidNumber, "The number at position %d does not match the operand syntax."},
	{calc.ErrOperandOutOfRange, codeOperandOutOfRange, "The number at position %d is outside the operand limits."},
	{calc.ErrDivisionByZero, codeDivisionByZero, "The expression divides by zero at position %d."},
	{calc.ErrNegativeSquareRoot, codeNegativeSquareRoot, "The square root at position %d receives a negative value."},
	{calc.ErrExponentNotInteger, codeExponentNotInteger, "The exponent of the power at position %d is not an exact integer."},
	{calc.ErrExponentOutOfRange, codeExponentOutOfRange, "The exponent of the power at position %d is not between -1000 and 1000."},
	{calc.ErrResultTooLarge, codeResultTooLarge, "The absolute value of the result at position %d is 10^40 or more."},
	{calc.ErrResultTooPrecise, codeResultTooPrecise, "The result at position %d needs more digits than the limits allow."},
}

// evaluate serves POST /api/v1/evaluate, which computes an expression (D41).
func evaluate() http.HandlerFunc {
	return operation([]string{"expression"}, func(f fields) (calc.Result, *problem) {
		if p := missing(f, "expression"); p != nil {
			return calc.Result{}, p
		}
		s, ok := stringValue(f["expression"])
		if !ok {
			return calc.Result{}, &problem{code: codeInvalidExpression, detail: "The field expression must be a JSON string.", field: "expression"}
		}
		res, err := expr.Evaluate(s)
		return res, expressionProblem(err)
	})
}

// expressionProblem turns an error from expr.Evaluate into a problem with a
// position (D53). Any error that expr does not describe is a bug, which
// calcProblem logs and reports as internal_error.
func expressionProblem(err error) *problem {
	if err == nil {
		return nil
	}
	if e, ok := errors.AsType[*expr.Error](err); ok {
		for _, f := range expressionFailures {
			if errors.Is(e.Err, f.err) {
				pos := e.Position
				return &problem{code: f.code, detail: fmt.Sprintf(f.detail, pos), position: &pos}
			}
		}
	}
	return calcProblem(err, nil)
}
