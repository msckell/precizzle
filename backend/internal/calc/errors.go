package calc

import "errors"

// Errors returned by Parse and by the operations. Each one maps to an error code in docs/SPEC.md.
var (
	// ErrInvalidNumber means that a string does not match the operand syntax.
	ErrInvalidNumber = errors.New("invalid number")
	// ErrOperandOutOfRange means that an operand is outside the limits.
	ErrOperandOutOfRange = errors.New("operand out of range")
	// ErrDivisionByZero means that a divisor is zero.
	ErrDivisionByZero = errors.New("division by zero")
	// ErrNegativeSquareRoot means that a square root received a negative value.
	ErrNegativeSquareRoot = errors.New("square root of a negative number")
	// ErrExponentNotInteger means that an exponent is not an exact integer.
	ErrExponentNotInteger = errors.New("exponent is not an integer")
	// ErrExponentOutOfRange means that an exponent is below -1000 or above 1000.
	ErrExponentOutOfRange = errors.New("exponent out of range")
	// ErrResultTooLarge means that the absolute value of a result is at least 10^40.
	ErrResultTooLarge = errors.New("result too large")
	// ErrResultTooPrecise means that a result needs more digits than the limits allow.
	ErrResultTooPrecise = errors.New("result too precise")
)
