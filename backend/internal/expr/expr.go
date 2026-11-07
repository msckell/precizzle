// Package expr evaluates the expressions of section 5 of docs/SPEC.md with
// the operations of package calc.
package expr

import (
	"errors"
	"fmt"
)

// Errors about the text of an expression. The other causes of an *Error are the errors of package calc.
var (
	// ErrInvalidExpression means that an expression does not follow the expression syntax.
	ErrInvalidExpression = errors.New("invalid expression")
	// ErrExpressionTooComplex means that an expression has too many operations or powers, or nests too deep.
	ErrExpressionTooComplex = errors.New("expression too complex")
)

// Error is an error at a position of an expression: the 0-based index of the character at fault.
type Error struct {
	Err      error
	Position int
}

// Error returns the cause and the position.
func (e *Error) Error() string {
	return fmt.Sprintf("%v at position %d", e.Err, e.Position)
}

// Unwrap returns the cause, so errors.Is matches it.
func (e *Error) Unwrap() error {
	return e.Err
}
