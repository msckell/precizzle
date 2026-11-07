package api

import (
	"encoding/json/v2"
	"net/http"
	"net/http/httptest"
	"testing"
)

// problemBody is the Problem Details object as a client reads it.
type problemBody struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Detail   string `json:"detail"`
	Code     string `json:"code"`
	Field    string `json:"field"`
	Position *int   `json:"position"`
}

// readProblem checks the headers of a Problem Details response and decodes its body strictly.
func readProblem(t *testing.T, rec *httptest.ResponseRecorder) problemBody {
	t.Helper()
	if got := rec.Header().Get("Content-Type"); got != "application/problem+json" {
		t.Fatalf("Content-Type = %q, want application/problem+json", got)
	}
	var p problemBody
	if err := json.Unmarshal(rec.Body.Bytes(), &p, json.RejectUnknownMembers(true)); err != nil {
		t.Fatalf("problem body %q does not decode: %v", rec.Body.String(), err)
	}
	if p.Status != rec.Code {
		t.Errorf("status member = %d, HTTP status = %d", p.Status, rec.Code)
	}
	return p
}

func TestWriteProblem(t *testing.T) {
	tests := []struct {
		code   string
		status int
		title  string
	}{
		{codeInvalidJSON, http.StatusBadRequest, "Invalid JSON"},
		{codeMissingField, http.StatusBadRequest, "Missing field"},
		{codeUnknownField, http.StatusBadRequest, "Unknown field"},
		{codeInvalidNumber, http.StatusBadRequest, "Invalid number"},
		{codeOperandOutOfRange, http.StatusBadRequest, "Operand out of range"},
		{codeExponentNotInteger, http.StatusBadRequest, "Exponent not integer"},
		{codeExponentOutOfRange, http.StatusBadRequest, "Exponent out of range"},
		{codeInvalidOperator, http.StatusBadRequest, "Invalid operator"},
		{codeInvalidExpression, http.StatusBadRequest, "Invalid expression"},
		{codeExpressionTooComplex, http.StatusBadRequest, "Expression too complex"},
		{codeDivisionByZero, http.StatusUnprocessableEntity, "Division by zero"},
		{codeNegativeSquareRoot, http.StatusUnprocessableEntity, "Negative square root"},
		{codeResultTooLarge, http.StatusUnprocessableEntity, "Result too large"},
		{codeResultTooPrecise, http.StatusUnprocessableEntity, "Result too precise"},
		{codeNotFound, http.StatusNotFound, "Not found"},
		{codeMethodNotAllowed, http.StatusMethodNotAllowed, "Method not allowed"},
		{codePayloadTooLarge, http.StatusRequestEntityTooLarge, "Payload too large"},
		{codeUnsupportedMediaType, http.StatusUnsupportedMediaType, "Unsupported media type"},
		{codeInternalError, http.StatusInternalServerError, "Internal error"},
	}
	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			rec := httptest.NewRecorder()
			writeProblem(rec, &problem{code: tt.code, detail: "Something specific.", field: "a"})
			if rec.Code != tt.status {
				t.Errorf("status = %d, want %d", rec.Code, tt.status)
			}
			want := problemBody{
				Type:   "https://github.com/msckell/precizzle/blob/main/docs/SPEC.md#" + tt.code,
				Title:  tt.title,
				Status: tt.status,
				Detail: "Something specific.",
				Code:   tt.code,
				Field:  "a",
			}
			if got := readProblem(t, rec); got != want {
				t.Errorf("problem = %+v, want %+v", got, want)
			}
		})
	}
}

func TestWriteProblemOmitsAnEmptyField(t *testing.T) {
	rec := httptest.NewRecorder()
	writeProblem(rec, &problem{code: codeResultTooLarge, detail: "Too large."})
	var members map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &members); err != nil {
		t.Fatal(err)
	}
	if _, ok := members["field"]; ok {
		t.Errorf("body %s has a field member", rec.Body.String())
	}
}
