package api

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
	"net/http"

	"github.com/msckell/precizzle/backend/internal/calc"
)

// resultBody is the success response of every operation.
type resultBody struct {
	Value   string `json:"value"`
	Decimal string `json:"decimal"`
	Kind    string `json:"kind"`
}

// failures maps the errors that are specific to an operation to the problem they produce.
type failures map[error]*problem

var (
	divideFailures = failures{
		calc.ErrDivisionByZero: {code: codeDivisionByZero, detail: "The divisor b is zero.", field: "b"},
	}
	powerFailures = failures{
		calc.ErrExponentNotInteger: {code: codeExponentNotInteger, detail: "The exponent must be an exact integer.", field: "exponent"},
		calc.ErrExponentOutOfRange: {code: codeExponentOutOfRange, detail: "The exponent must be between -1000 and 1000.", field: "exponent"},
		calc.ErrDivisionByZero:     {code: codeDivisionByZero, detail: "The base is 0 and the exponent is negative, which divides by zero.", field: "base"},
	}
	sqrtFailures = failures{
		calc.ErrNegativeSquareRoot: {code: codeNegativeSquareRoot, detail: "The value is negative, so its square root is not a real number.", field: "value"},
	}
	percentageFailures = failures{
		calc.ErrDivisionByZero: {code: codeDivisionByZero, detail: "The percent is zero, so the base cannot be divided by it.", field: "percent"},
	}
	percentOperators = map[string]calc.PercentOperator{
		"add":      calc.PercentAdd,
		"subtract": calc.PercentSubtract,
		"multiply": calc.PercentMultiply,
		"divide":   calc.PercentDivide,
	}
)

// operation builds the handler of an endpoint whose request object may have
// the allowed fields. run checks the fields and computes the result.
func operation(allowed []string, run func(fields) (calc.Result, *problem)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		f, p := decodeFields(w, r, allowed)
		if p == nil {
			var res calc.Result
			if res, p = run(f); p == nil {
				writeJSON(w, "application/json", http.StatusOK, resultBody{res.Value, res.Decimal, res.Kind.String()})
				return
			}
		}
		writeProblem(w, p)
	}
}

// binary serves the operations that take the fields a and b.
func binary(op func(a, b calc.Number) (calc.Result, error), specific failures) http.HandlerFunc {
	return operation([]string{"a", "b"}, func(f fields) (calc.Result, *problem) {
		if p := missing(f, "a", "b"); p != nil {
			return calc.Result{}, p
		}
		a, aSyntax, aLimit := number(f, "a")
		b, bSyntax, bLimit := number(f, "b")
		if p := first(aSyntax, bSyntax, aLimit, bLimit); p != nil {
			return calc.Result{}, p
		}
		res, err := op(a, b)
		return res, calcProblem(err, specific)
	})
}

func power() http.HandlerFunc {
	return operation([]string{"base", "exponent"}, func(f fields) (calc.Result, *problem) {
		if p := missing(f, "base", "exponent"); p != nil {
			return calc.Result{}, p
		}
		base, baseSyntax, baseLimit := number(f, "base")
		exponent, exponentSyntax, exponentLimit := number(f, "exponent")
		if p := first(baseSyntax, exponentSyntax, baseLimit, exponentLimit); p != nil {
			return calc.Result{}, p
		}
		res, err := calc.Power(base, exponent)
		return res, calcProblem(err, powerFailures)
	})
}

func sqrt() http.HandlerFunc {
	return operation([]string{"value"}, func(f fields) (calc.Result, *problem) {
		if p := missing(f, "value"); p != nil {
			return calc.Result{}, p
		}
		value, syntax, limit := number(f, "value")
		if p := first(syntax, limit); p != nil {
			return calc.Result{}, p
		}
		res, err := calc.Sqrt(value)
		return res, calcProblem(err, sqrtFailures)
	})
}

// percentage serves B% alone, or A op B% when base and operator are both given (D7).
func percentage() http.HandlerFunc {
	return operation([]string{"percent", "base", "operator"}, func(f fields) (calc.Result, *problem) {
		if p := missing(f, "percent"); p != nil {
			return calc.Result{}, p
		}
		_, hasBase := f["base"]
		_, hasOperator := f["operator"]
		if hasBase != hasOperator {
			name, other := "operator", "base"
			if hasOperator {
				name, other = "base", "operator"
			}
			return calc.Result{}, &problem{
				code:   codeMissingField,
				detail: fmt.Sprintf("The field %s is required when %s is given.", name, other),
				field:  name,
			}
		}

		percent, percentSyntax, percentLimit := number(f, "percent")
		if !hasBase {
			if p := first(percentSyntax, percentLimit); p != nil {
				return calc.Result{}, p
			}
			res, err := calc.Percentage(percent)
			return res, calcProblem(err, nil)
		}
		base, baseSyntax, baseLimit := number(f, "base")
		op, opProblem := percentOperator(f)
		if p := first(percentSyntax, baseSyntax, opProblem, percentLimit, baseLimit); p != nil {
			return calc.Result{}, p
		}
		res, err := calc.ApplyPercentage(base, op, percent)
		return res, calcProblem(err, percentageFailures)
	})
}

// missing reports the first of names that the request does not give.
func missing(f fields, names ...string) *problem {
	for _, name := range names {
		if _, ok := f[name]; !ok {
			return &problem{code: codeMissingField, detail: fmt.Sprintf("The field %s is required.", name), field: name}
		}
	}
	return nil
}

// number reads the number field name. A syntax problem and a limit problem
// are returned apart, because the syntax of every field is checked before
// the limits of any field.
func number(f fields, name string) (n calc.Number, syntax, limit *problem) {
	s, ok := stringValue(f[name])
	if !ok {
		return n, &problem{code: codeInvalidNumber, detail: fmt.Sprintf("The field %s must be a JSON string.", name), field: name}, nil
	}
	n, err := calc.Parse(s)
	switch {
	case errors.Is(err, calc.ErrInvalidNumber):
		syntax = &problem{code: codeInvalidNumber, detail: fmt.Sprintf("The field %s does not match the operand syntax.", name), field: name}
	case errors.Is(err, calc.ErrOperandOutOfRange):
		limit = &problem{code: codeOperandOutOfRange, detail: fmt.Sprintf("The field %s is outside the operand limits.", name), field: name}
	}
	return n, syntax, limit
}

func percentOperator(f fields) (calc.PercentOperator, *problem) {
	s, _ := stringValue(f["operator"])
	if op, ok := percentOperators[s]; ok {
		return op, nil
	}
	return 0, &problem{code: codeInvalidOperator, detail: "The field operator must be add, subtract, multiply or divide.", field: "operator"}
}

// stringValue returns the string that a JSON value holds, and false when the value is not a string.
func stringValue(v jsontext.Value) (string, bool) {
	if v.Kind() != '"' {
		return "", false
	}
	var s string
	// A valid JSON string always unmarshals into a Go string.
	_ = json.Unmarshal(v, &s)
	return s, true
}

// first returns the first problem that is not nil.
func first(problems ...*problem) *problem {
	for _, p := range problems {
		if p != nil {
			return p
		}
	}
	return nil
}

// calcProblem turns an error from package calc into a problem. specific
// holds the errors whose problem depends on the operation.
func calcProblem(err error, specific failures) *problem {
	if err == nil {
		return nil
	}
	for target, p := range specific {
		if errors.Is(err, target) {
			return p
		}
	}
	switch {
	case errors.Is(err, calc.ErrResultTooLarge):
		return &problem{code: codeResultTooLarge, detail: "The absolute value of the result is 10^40 or more."}
	case errors.Is(err, calc.ErrResultTooPrecise):
		return &problem{code: codeResultTooPrecise, detail: "The result needs more digits than the limits allow."}
	}
	panic(fmt.Sprintf("api: unexpected error from calc: %v", err))
}
