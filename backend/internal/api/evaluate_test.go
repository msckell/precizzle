package api

import (
	"bytes"
	"encoding/json/v2"
	"errors"
	"log"
	"net/http"
	"strings"
	"testing"

	"github.com/msckell/precizzle/backend/internal/expr"
)

// evaluateBody builds the request body of evaluate for an expression.
func evaluateBody(t *testing.T, expression string) string {
	t.Helper()
	body, err := json.Marshal(map[string]string{"expression": expression})
	if err != nil {
		t.Fatal(err)
	}
	return string(body)
}

// TestEvaluateMatchesExpr checks each success example of SPEC §5 through HTTP.
func TestEvaluateMatchesExpr(t *testing.T) {
	tests := []struct {
		expression string
		value      string
		kind       string
	}{
		{"0.1+0.2", "0.3", "exact"},
		{"2+3*4", "14", "exact"},
		{"(2+3)*4", "20", "exact"},
		{"2^3^2", "512", "exact"},
		{"-2^2", "-4", "exact"},
		{"2^-2", "0.25", "exact"},
		{"1/3*3", "1", "exact"},
		{"[1/7]*7", "1", "exact"},
		{"sqrt(9/4)", "1.5", "exact"},
		{"sqrt(2)*sqrt(2)", "~2", "approximate"},
		{"200+15%", "230", "exact"},
		{"200-15%", "170", "exact"},
		{"200*15%", "30", "exact"},
		{"200/15%", "4000/3", "repeating"},
		{"15%", "0.15", "exact"},
		{"50+200+15%", "280", "exact"},
		{"50-200+15%", "-180", "exact"},
		{"200+15%+10%", "253", "exact"},
		{"(50+200)+15%", "287.5", "exact"},
		{"15%+200", "200.15", "exact"},
		{"200+-15%", "170", "exact"},
	}
	for _, tt := range tests {
		t.Run(tt.expression, func(t *testing.T) {
			want, err := expr.Evaluate(tt.expression)
			if err != nil {
				t.Fatalf("expr.Evaluate error = %v", err)
			}
			got := readResult(t, serve(t, http.MethodPost, "/api/v1/evaluate", evaluateBody(t, tt.expression)))
			if got != (resultBody{want.Value, want.Decimal, want.Kind.String()}) || got.Value != tt.value || got.Kind != tt.kind {
				t.Errorf("result = %+v, want value %q, kind %q and decimal %q", got, tt.value, tt.kind, want.Decimal)
			}
		})
	}
}

// TestEvaluateErrors checks each error example of SPEC §5 and the request errors of evaluate.
func TestEvaluateErrors(t *testing.T) {
	position := func(n int) *int { return &n }
	tests := []struct {
		name     string
		body     string
		status   int
		code     string
		field    string
		position *int
	}{
		{name: "1/0", status: 422, code: codeDivisionByZero, position: position(1)},
		{name: "0^-1", status: 422, code: codeDivisionByZero, position: position(1)},
		{name: "sqrt(-4)", status: 422, code: codeNegativeSquareRoot, position: position(0)},
		{name: "2^0.5", status: 400, code: codeExponentNotInteger, position: position(1)},
		{name: "2^1001", status: 400, code: codeExponentOutOfRange, position: position(1)},
		{name: "10^50/10^20", status: 422, code: codeResultTooLarge, position: position(2)},
		{name: "[1e-1000]/10", status: 422, code: codeResultTooPrecise, position: position(9)},
		{name: "2+", status: 400, code: codeInvalidExpression, position: position(2)},
		{name: "(2+3", status: 400, code: codeInvalidExpression, position: position(0)},
		{name: "(2+", status: 400, code: codeInvalidExpression, position: position(3)},
		{name: "2(3)", status: 400, code: codeInvalidExpression, position: position(1)},
		{name: "15%%", status: 400, code: codeInvalidExpression, position: position(3)},
		{name: "1.2.3", status: 400, code: codeInvalidNumber, position: position(0)},
		{name: "2+1e40", status: 400, code: codeOperandOutOfRange, position: position(2)},
		{name: "101 operations", body: evaluateBody(t, strings.Repeat("1+", 101)+"1"), status: 400, code: codeExpressionTooComplex, position: position(201)},
		{name: "missing expression", body: `{}`, status: 400, code: codeMissingField, field: "expression"},
		{name: "null expression", body: `{"expression": null}`, status: 400, code: codeMissingField, field: "expression"},
		{name: "unknown field", body: `{"expression": "1", "value": "2"}`, status: 400, code: codeUnknownField, field: "value"},
		{name: "member in another case", body: `{"Expression": "1"}`, status: 400, code: codeUnknownField, field: "Expression"},
		{name: "JSON number", body: `{"expression": 1}`, status: 400, code: codeInvalidExpression, field: "expression"},
		{name: "JSON boolean", body: `{"expression": true}`, status: 400, code: codeInvalidExpression, field: "expression"},
		{name: "JSON object", body: `{"expression": {}}`, status: 400, code: codeInvalidExpression, field: "expression"},
		{name: "JSON array", body: `{"expression": ["1"]}`, status: 400, code: codeInvalidExpression, field: "expression"},
		{name: "invalid JSON", body: `{"expression": "1"`, status: 400, code: codeInvalidJSON},
		{name: "escaped non-ASCII", body: `{"expression": "2+é"}`, status: 400, code: codeInvalidExpression, position: position(2)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := tt.body
			if body == "" {
				body = evaluateBody(t, tt.name)
			}
			rec := serve(t, http.MethodPost, "/api/v1/evaluate", body)
			if rec.Code != tt.status {
				t.Errorf("status = %d, want %d", rec.Code, tt.status)
			}
			p := readProblem(t, rec)
			if p.Code != tt.code || p.Field != tt.field || p.Detail == "" {
				t.Errorf("problem = %+v, want code %q, field %q and a detail", p, tt.code, tt.field)
			}
			if (p.Position == nil) != (tt.position == nil) || (p.Position != nil && *p.Position != *tt.position) {
				t.Errorf("position = %v, want %v", deref(p.Position), deref(tt.position))
			}
		})
	}
}

func deref(n *int) any {
	if n == nil {
		return nil
	}
	return *n
}

// TestEvaluateErrorExample checks the error example of SPEC §5 member by
// member, and that a position of 0 is written.
func TestEvaluateErrorExample(t *testing.T) {
	rec := serve(t, http.MethodPost, "/api/v1/evaluate", `{"expression": "1/0"}`)
	want := `{"type":"https://github.com/msckell/precizzle/blob/main/docs/SPEC.md#division_by_zero","title":"Division by zero","status":422,"detail":"The expression divides by zero at position 1.","code":"division_by_zero","position":1}`
	if rec.Code != http.StatusUnprocessableEntity || rec.Body.String() != want {
		t.Errorf("response = %d %s, want 422 %s", rec.Code, rec.Body.String(), want)
	}
	rec = serve(t, http.MethodPost, "/api/v1/evaluate", `{"expression": "sqrt(-4)"}`)
	if !strings.Contains(rec.Body.String(), `"position":0`) {
		t.Errorf("body %s does not write position 0", rec.Body.String())
	}
}

func TestEvaluateRejectsOtherMethods(t *testing.T) {
	rec := serve(t, http.MethodGet, "/api/v1/evaluate", "")
	if p := readProblem(t, rec); p.Code != codeMethodNotAllowed || rec.Header().Get("Allow") != "POST" {
		t.Errorf("problem = %+v, Allow = %q, want %s with Allow: POST", p, rec.Header().Get("Allow"), codeMethodNotAllowed)
	}
}

// TestEvaluateChainsTheValue sends a result back as a value in brackets (D21, D50).
func TestEvaluateChainsTheValue(t *testing.T) {
	third := readResult(t, serve(t, http.MethodPost, "/api/v1/evaluate", `{"expression": "1/3"}`))
	if third.Value != "1/3" || third.Kind != "repeating" {
		t.Fatalf("1/3 = %+v, want the repeating value 1/3", third)
	}
	got := readResult(t, serve(t, http.MethodPost, "/api/v1/evaluate", evaluateBody(t, "["+third.Value+"]*3")))
	if got != (resultBody{"1", "1", "exact"}) {
		t.Errorf("[1/3]*3 = %+v, want exactly 1", got)
	}
}

// TestUnexpectedExpressionError checks that an error expr does not describe
// becomes internal_error, with no position, and is logged (D37, D53).
func TestUnexpectedExpressionError(t *testing.T) {
	var logs bytes.Buffer
	previous := log.Writer()
	log.SetOutput(&logs)
	t.Cleanup(func() { log.SetOutput(previous) })

	for _, err := range []error{
		errors.New("expr: the value of an operation does not parse"),
		&expr.Error{Err: errors.New("an unknown cause"), Position: 4},
	} {
		p := expressionProblem(err)
		if p.code != codeInternalError || p.position != nil || p.field != "" {
			t.Errorf("expressionProblem(%v) = %+v, want %s with no position", err, p, codeInternalError)
		}
		if !strings.Contains(logs.String(), err.Error()) {
			t.Errorf("log %q does not record %v", logs.String(), err)
		}
	}
	if p := expressionProblem(nil); p != nil {
		t.Errorf("expressionProblem(nil) = %+v, want nil", p)
	}
}
