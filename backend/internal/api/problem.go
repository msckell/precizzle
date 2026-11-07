// Package api serves the HTTP API of docs/SPEC.md on top of package calc.
package api

import (
	"encoding/json/v2"
	"net/http"
)

// The machine-readable error codes of docs/SPEC.md.
const (
	codeInvalidJSON          = "invalid_json"
	codeMissingField         = "missing_field"
	codeUnknownField         = "unknown_field"
	codeInvalidNumber        = "invalid_number"
	codeOperandOutOfRange    = "operand_out_of_range"
	codeExponentNotInteger   = "exponent_not_integer"
	codeExponentOutOfRange   = "exponent_out_of_range"
	codeInvalidOperator      = "invalid_operator"
	codeInvalidExpression    = "invalid_expression"
	codeExpressionTooComplex = "expression_too_complex"
	codeDivisionByZero       = "division_by_zero"
	codeNegativeSquareRoot   = "negative_square_root"
	codeResultTooLarge       = "result_too_large"
	codeResultTooPrecise     = "result_too_precise"
	codeNotFound             = "not_found"
	codeMethodNotAllowed     = "method_not_allowed"
	codePayloadTooLarge      = "payload_too_large"
	codeUnsupportedMediaType = "unsupported_media_type"
	codeInternalError        = "internal_error"
)

// problemTypeBase is the start of every problem type URI. The code is the
// anchor of its row in the errors table of docs/SPEC.md.
const problemTypeBase = "https://github.com/msckell/precizzle/blob/main/docs/SPEC.md#"

// problemTypes holds the status and title of each code, as listed in docs/SPEC.md.
var problemTypes = map[string]struct {
	status int
	title  string
}{
	codeInvalidJSON:          {http.StatusBadRequest, "Invalid JSON"},
	codeMissingField:         {http.StatusBadRequest, "Missing field"},
	codeUnknownField:         {http.StatusBadRequest, "Unknown field"},
	codeInvalidNumber:        {http.StatusBadRequest, "Invalid number"},
	codeOperandOutOfRange:    {http.StatusBadRequest, "Operand out of range"},
	codeExponentNotInteger:   {http.StatusBadRequest, "Exponent not integer"},
	codeExponentOutOfRange:   {http.StatusBadRequest, "Exponent out of range"},
	codeInvalidOperator:      {http.StatusBadRequest, "Invalid operator"},
	codeInvalidExpression:    {http.StatusBadRequest, "Invalid expression"},
	codeExpressionTooComplex: {http.StatusBadRequest, "Expression too complex"},
	codeDivisionByZero:       {http.StatusUnprocessableEntity, "Division by zero"},
	codeNegativeSquareRoot:   {http.StatusUnprocessableEntity, "Negative square root"},
	codeResultTooLarge:       {http.StatusUnprocessableEntity, "Result too large"},
	codeResultTooPrecise:     {http.StatusUnprocessableEntity, "Result too precise"},
	codeNotFound:             {http.StatusNotFound, "Not found"},
	codeMethodNotAllowed:     {http.StatusMethodNotAllowed, "Method not allowed"},
	codePayloadTooLarge:      {http.StatusRequestEntityTooLarge, "Payload too large"},
	codeUnsupportedMediaType: {http.StatusUnsupportedMediaType, "Unsupported media type"},
	codeInternalError:        {http.StatusInternalServerError, "Internal error"},
}

// problem is one occurrence of an error: its code, a human-readable detail,
// and the request field or the position in the expression at fault, if any.
// position is a pointer because 0 is a valid position.
type problem struct {
	code     string
	detail   string
	field    string
	position *int
}

// writeProblem writes p as an RFC 9457 Problem Details response.
func writeProblem(w http.ResponseWriter, p *problem) {
	t := problemTypes[p.code]
	writeJSON(w, "application/problem+json", t.status, struct {
		Type     string `json:"type"`
		Title    string `json:"title"`
		Status   int    `json:"status"`
		Detail   string `json:"detail"`
		Code     string `json:"code"`
		Field    string `json:"field,omitempty"`
		Position *int   `json:"position,omitempty"`
	}{problemTypeBase + p.code, t.title, t.status, p.detail, p.code, p.field, p.position})
}

// writeJSON writes v as the JSON body of a response with the given media type and status.
func writeJSON(w http.ResponseWriter, mediaType string, status int, v any) {
	w.Header().Set("Content-Type", mediaType)
	w.WriteHeader(status)
	// The values written here always marshal, and a failed write means the
	// client has gone away, so the error is ignored.
	_ = json.MarshalWrite(w, v)
}
