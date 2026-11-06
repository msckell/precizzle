package api

import (
	"encoding/json/v2"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/msckell/precizzle/backend/internal/calc"
)

// serve sends one request to the API. A non-empty body is sent as application/json.
func serve(t *testing.T, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	NewHandler().ServeHTTP(rec, req)
	return rec
}

// readResult checks a success response and decodes its body strictly.
func readResult(t *testing.T, rec *httptest.ResponseRecorder) resultBody {
	t.Helper()
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body %s", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", got)
	}
	var r resultBody
	if err := json.Unmarshal(rec.Body.Bytes(), &r, json.RejectUnknownMembers(true)); err != nil {
		t.Fatalf("result body %q does not decode: %v", rec.Body.String(), err)
	}
	return r
}

// operand parses a test operand for a direct call to package calc.
func operand(t *testing.T, s string) calc.Number {
	t.Helper()
	n, err := calc.Parse(s)
	if err != nil {
		t.Fatalf("calc.Parse(%q) error = %v", s, err)
	}
	return n
}

// TestEndpointsMatchCalc checks that each endpoint returns exactly what package calc computes.
func TestEndpointsMatchCalc(t *testing.T) {
	tests := []struct {
		path string
		body string
		calc func(t *testing.T) (calc.Result, error)
	}{
		{"/api/v1/add", `{"a": "0.1", "b": "0.2"}`, func(t *testing.T) (calc.Result, error) { return calc.Add(operand(t, "0.1"), operand(t, "0.2")) }},
		{"/api/v1/add", `{"b": "1.20", "a": "1.10"}`, func(t *testing.T) (calc.Result, error) { return calc.Add(operand(t, "1.10"), operand(t, "1.20")) }},
		{"/api/v1/add", `{"a": "1e-25", "b": "0"}`, func(t *testing.T) (calc.Result, error) { return calc.Add(operand(t, "1e-25"), operand(t, "0")) }},
		{"/api/v1/subtract", `{"a": "1", "b": "1/3"}`, func(t *testing.T) (calc.Result, error) { return calc.Subtract(operand(t, "1"), operand(t, "1/3")) }},
		{"/api/v1/multiply", `{"a": "1/3", "b": "3"}`, func(t *testing.T) (calc.Result, error) { return calc.Multiply(operand(t, "1/3"), operand(t, "3")) }},
		{"/api/v1/multiply", `{"a": "~1.4142135623730950488", "b": "~1.4142135623730950488"}`, func(t *testing.T) (calc.Result, error) {
			return calc.Multiply(operand(t, "~1.4142135623730950488"), operand(t, "~1.4142135623730950488"))
		}},
		{"/api/v1/divide", `{"a": "1", "b": "3"}`, func(t *testing.T) (calc.Result, error) { return calc.Divide(operand(t, "1"), operand(t, "3")) }},
		{"/api/v1/divide", `{"a": "1", "b": "7"}`, func(t *testing.T) (calc.Result, error) { return calc.Divide(operand(t, "1"), operand(t, "7")) }},
		{"/api/v1/divide", `{"a": "1", "b": "3e30"}`, func(t *testing.T) (calc.Result, error) { return calc.Divide(operand(t, "1"), operand(t, "3e30")) }},
		{"/api/v1/power", `{"base": "2", "exponent": "-2"}`, func(t *testing.T) (calc.Result, error) { return calc.Power(operand(t, "2"), operand(t, "-2")) }},
		{"/api/v1/power", `{"base": "3", "exponent": "-1"}`, func(t *testing.T) (calc.Result, error) { return calc.Power(operand(t, "3"), operand(t, "-1")) }},
		{"/api/v1/power", `{"base": "0", "exponent": "0"}`, func(t *testing.T) (calc.Result, error) { return calc.Power(operand(t, "0"), operand(t, "0")) }},
		{"/api/v1/sqrt", `{"value": "2"}`, func(t *testing.T) (calc.Result, error) { return calc.Sqrt(operand(t, "2")) }},
		{"/api/v1/sqrt", `{"value": "2.25"}`, func(t *testing.T) (calc.Result, error) { return calc.Sqrt(operand(t, "2.25")) }},
		{"/api/v1/sqrt", `{"value": "1/9"}`, func(t *testing.T) (calc.Result, error) { return calc.Sqrt(operand(t, "1/9")) }},
		{"/api/v1/percentage", `{"percent": "15"}`, func(t *testing.T) (calc.Result, error) { return calc.Percentage(operand(t, "15")) }},
		{"/api/v1/percentage", `{"percent": "15", "base": "200", "operator": "add"}`, func(t *testing.T) (calc.Result, error) {
			return calc.ApplyPercentage(operand(t, "200"), calc.PercentAdd, operand(t, "15"))
		}},
		{"/api/v1/percentage", `{"percent": "15", "base": "200", "operator": "subtract"}`, func(t *testing.T) (calc.Result, error) {
			return calc.ApplyPercentage(operand(t, "200"), calc.PercentSubtract, operand(t, "15"))
		}},
		{"/api/v1/percentage", `{"percent": "15", "base": "200", "operator": "multiply"}`, func(t *testing.T) (calc.Result, error) {
			return calc.ApplyPercentage(operand(t, "200"), calc.PercentMultiply, operand(t, "15"))
		}},
		{"/api/v1/percentage", `{"operator": "divide", "base": "200", "percent": "15"}`, func(t *testing.T) (calc.Result, error) {
			return calc.ApplyPercentage(operand(t, "200"), calc.PercentDivide, operand(t, "15"))
		}},
	}
	for _, tt := range tests {
		t.Run(tt.path+" "+tt.body, func(t *testing.T) {
			want, err := tt.calc(t)
			if err != nil {
				t.Fatalf("calc error = %v", err)
			}
			got := readResult(t, serve(t, http.MethodPost, tt.path, tt.body))
			if got != (resultBody{want.Value, want.Decimal, want.Kind.String()}) {
				t.Errorf("response = %+v, calc = %+v", got, want)
			}
		})
	}
}

// TestContractExamples follows the examples of the API contract in docs/SPEC.md.
func TestContractExamples(t *testing.T) {
	third := "0." + strings.Repeat("3", 100)
	tests := []struct {
		path string
		body string
		want resultBody
	}{
		{"/api/v1/add", `{"a": "0.1", "b": "0.2"}`, resultBody{"0.3", "0.3", "exact"}},
		{"/api/v1/divide", `{"a": "1", "b": "3"}`, resultBody{"1/3", third, "repeating"}},
		{"/api/v1/multiply", `{"a": "1/3", "b": "3"}`, resultBody{"1", "1", "exact"}},
		{"/api/v1/subtract", `{"a": "1", "b": "1/3"}`, resultBody{"2/3", "0." + strings.Repeat("6", 99) + "7", "repeating"}},
		{"/api/v1/power", `{"base": "2", "exponent": "-2"}`, resultBody{"0.25", "0.25", "exact"}},
		{"/api/v1/sqrt", `{"value": "2"}`, resultBody{"~1.4142135623730950488", "1.4142135623730950488", "approximate"}},
		{"/api/v1/percentage", `{"percent": "15", "base": "200", "operator": "add"}`, resultBody{"230", "230", "exact"}},
		{"/api/v1/percentage", `{"percent": "15"}`, resultBody{"0.15", "0.15", "exact"}},
		{"/api/v1/divide", `{"a": "1", "b": "3e30"}`, resultBody{"1/3" + strings.Repeat("0", 30), "3." + strings.Repeat("3", 100) + "e-31", "repeating"}},
	}
	for _, tt := range tests {
		t.Run(tt.path+" "+tt.body, func(t *testing.T) {
			if got := readResult(t, serve(t, http.MethodPost, tt.path, tt.body)); got != tt.want {
				t.Errorf("response = %+v, want %+v", got, tt.want)
			}
		})
	}
}

// TestChainingSendsBackTheValue chains requests the way the frontend does (D21).
func TestChainingSendsBackTheValue(t *testing.T) {
	third := readResult(t, serve(t, http.MethodPost, "/api/v1/divide", `{"a": "1", "b": "3"}`))
	body, err := json.Marshal(map[string]string{"a": third.Value, "b": "3"})
	if err != nil {
		t.Fatal(err)
	}
	if got := readResult(t, serve(t, http.MethodPost, "/api/v1/multiply", string(body))); got != (resultBody{"1", "1", "exact"}) {
		t.Errorf("1 ÷ 3 × 3 = %+v, want exactly 1", got)
	}

	root := readResult(t, serve(t, http.MethodPost, "/api/v1/sqrt", `{"value": "2"}`))
	body, err = json.Marshal(map[string]string{"a": root.Value, "b": root.Value})
	if err != nil {
		t.Fatal(err)
	}
	if got := readResult(t, serve(t, http.MethodPost, "/api/v1/multiply", string(body))); got != (resultBody{"~2", "2", "approximate"}) {
		t.Errorf("√2 × √2 = %+v, want ~2", got)
	}
}

func TestErrors(t *testing.T) {
	tooLarge := "{" + strings.Repeat(" ", maxBodyBytes) + `"a": "1", "b": "2"}`
	tests := []struct {
		name      string
		method    string
		path      string
		body      string
		wantCode  string
		wantField string
		wantAllow string
	}{
		{name: "invalid JSON", path: "/api/v1/add", body: `{"a": "1",`, wantCode: codeInvalidJSON},
		{name: "missing b", path: "/api/v1/add", body: `{"a": "1"}`, wantCode: codeMissingField, wantField: "b"},
		{name: "null a", path: "/api/v1/add", body: `{"a": null, "b": "1"}`, wantCode: codeMissingField, wantField: "a"},
		{name: "missing value", path: "/api/v1/sqrt", body: `{}`, wantCode: codeMissingField, wantField: "value"},
		{name: "missing exponent", path: "/api/v1/power", body: `{"base": "2"}`, wantCode: codeMissingField, wantField: "exponent"},
		{name: "missing percent", path: "/api/v1/percentage", body: `{"base": "200", "operator": "add"}`, wantCode: codeMissingField, wantField: "percent"},
		{name: "base without operator", path: "/api/v1/percentage", body: `{"percent": "15", "base": "200"}`, wantCode: codeMissingField, wantField: "operator"},
		{name: "operator without base", path: "/api/v1/percentage", body: `{"percent": "15", "operator": "add"}`, wantCode: codeMissingField, wantField: "base"},
		{name: "unknown field", path: "/api/v1/add", body: `{"a": "1", "b": "2", "c": "3"}`, wantCode: codeUnknownField, wantField: "c"},
		{name: "field of another endpoint", path: "/api/v1/sqrt", body: `{"a": "4"}`, wantCode: codeUnknownField, wantField: "a"},
		{name: "operand syntax", path: "/api/v1/add", body: `{"a": "1.", "b": "2"}`, wantCode: codeInvalidNumber, wantField: "a"},
		{name: "JSON number", path: "/api/v1/add", body: `{"a": 1, "b": "2"}`, wantCode: codeInvalidNumber, wantField: "a"},
		{name: "JSON boolean", path: "/api/v1/add", body: `{"a": "1", "b": true}`, wantCode: codeInvalidNumber, wantField: "b"},
		{name: "JSON object", path: "/api/v1/sqrt", body: `{"value": {}}`, wantCode: codeInvalidNumber, wantField: "value"},
		{name: "operand out of range", path: "/api/v1/add", body: `{"a": "1e40", "b": "1"}`, wantCode: codeOperandOutOfRange, wantField: "a"},
		{name: "exponent 0.5", path: "/api/v1/power", body: `{"base": "2", "exponent": "0.5"}`, wantCode: codeExponentNotInteger, wantField: "exponent"},
		{name: "approximate exponent", path: "/api/v1/power", body: `{"base": "2", "exponent": "~2"}`, wantCode: codeExponentNotInteger, wantField: "exponent"},
		{name: "exponent 1001", path: "/api/v1/power", body: `{"base": "2", "exponent": "1001"}`, wantCode: codeExponentOutOfRange, wantField: "exponent"},
		{name: "unknown operator", path: "/api/v1/percentage", body: `{"percent": "15", "base": "200", "operator": "plus"}`, wantCode: codeInvalidOperator, wantField: "operator"},
		{name: "operator in another case", path: "/api/v1/percentage", body: `{"percent": "15", "base": "200", "operator": "Add"}`, wantCode: codeInvalidOperator, wantField: "operator"},
		{name: "operator that is not a string", path: "/api/v1/percentage", body: `{"percent": "15", "base": "200", "operator": 1}`, wantCode: codeInvalidOperator, wantField: "operator"},
		{name: "1 ÷ 0", path: "/api/v1/divide", body: `{"a": "1", "b": "0"}`, wantCode: codeDivisionByZero, wantField: "b"},
		{name: "0 ^ -1", path: "/api/v1/power", body: `{"base": "0", "exponent": "-1"}`, wantCode: codeDivisionByZero, wantField: "base"},
		{name: "200 ÷ 0%", path: "/api/v1/percentage", body: `{"percent": "0", "base": "200", "operator": "divide"}`, wantCode: codeDivisionByZero, wantField: "percent"},
		{name: "√-4", path: "/api/v1/sqrt", body: `{"value": "-4"}`, wantCode: codeNegativeSquareRoot, wantField: "value"},
		{name: "10 ^ 40", path: "/api/v1/power", body: `{"base": "10", "exponent": "40"}`, wantCode: codeResultTooLarge},
		{name: "sum too large", path: "/api/v1/add", body: `{"a": "` + strings.Repeat("9", 40) + `", "b": "1"}`, wantCode: codeResultTooLarge},
		{name: "1e-1000 × 1e-1000", path: "/api/v1/multiply", body: `{"a": "1e-1000", "b": "1e-1000"}`, wantCode: codeResultTooPrecise},
		{name: "0.25 ^ 501", path: "/api/v1/power", body: `{"base": "0.25", "exponent": "501"}`, wantCode: codeResultTooPrecise},
		{name: "percentage too precise", path: "/api/v1/percentage", body: `{"percent": "1e-1000"}`, wantCode: codeResultTooPrecise},
		{name: "root path", method: http.MethodGet, path: "/", wantCode: codeNotFound},
		{name: "unknown operation", path: "/api/v1/modulo", body: `{}`, wantCode: codeNotFound},
		{name: "trailing slash", path: "/api/v1/add/", body: `{}`, wantCode: codeNotFound},
		{name: "path outside the API", method: http.MethodGet, path: "/favicon.ico", wantCode: codeNotFound},
		{name: "GET an operation, checked before the media type", method: http.MethodGet, path: "/api/v1/add", wantCode: codeMethodNotAllowed, wantAllow: "POST"},
		{name: "PUT an operation", method: http.MethodPut, path: "/api/v1/sqrt", body: `{"value": "4"}`, wantCode: codeMethodNotAllowed, wantAllow: "POST"},
		{name: "OPTIONS an operation", method: http.MethodOptions, path: "/api/v1/percentage", wantCode: codeMethodNotAllowed, wantAllow: "POST"},
		{name: "body over 64 KiB", path: "/api/v1/add", body: tooLarge, wantCode: codePayloadTooLarge},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			method := tt.method
			if method == "" {
				method = http.MethodPost
			}
			rec := serve(t, method, tt.path, tt.body)
			p := readProblem(t, rec)
			if p.Code != tt.wantCode || p.Field != tt.wantField {
				t.Errorf("problem = %+v, want code %q and field %q", p, tt.wantCode, tt.wantField)
			}
			if p.Detail == "" {
				t.Error("detail is empty")
			}
			if got := rec.Header().Get("Allow"); got != tt.wantAllow {
				t.Errorf("Allow = %q, want %q", got, tt.wantAllow)
			}
		})
	}
}

func TestUnsupportedMediaType(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/add", strings.NewReader(`{"a": "1", "b": "2"}`))
	req.Header.Set("Content-Type", "text/plain")
	rec := httptest.NewRecorder()
	NewHandler().ServeHTTP(rec, req)
	if p := readProblem(t, rec); p.Code != codeUnsupportedMediaType {
		t.Errorf("problem = %+v, want %s", p, codeUnsupportedMediaType)
	}
}

// TestCheckOrder sends requests with two problems and expects the first one in the order of docs/SPEC.md.
func TestCheckOrder(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		body      string
		wantCode  string
		wantField string
	}{
		{"unknown field before a JSON number", "/api/v1/add", `{"a": 1, "c": "2"}`, codeUnknownField, "c"},
		{"missing field before invalid syntax", "/api/v1/add", `{"a": "x"}`, codeMissingField, "b"},
		{"missing fields in table order", "/api/v1/power", `{}`, codeMissingField, "base"},
		{"syntax of b before the limits of a", "/api/v1/add", `{"a": "1e40", "b": "x"}`, codeInvalidNumber, "b"},
		{"limits in field order", "/api/v1/add", `{"b": "1e40", "a": "1e40"}`, codeOperandOutOfRange, "a"},
		{"operand limits before exponent rules", "/api/v1/power", `{"base": "1e40", "exponent": "0.5"}`, codeOperandOutOfRange, "base"},
		{"exponent rules before the calculation", "/api/v1/power", `{"base": "0", "exponent": "-0.5"}`, codeExponentNotInteger, "exponent"},
		{"operator before operand limits", "/api/v1/percentage", `{"percent": "1e40", "base": "1", "operator": "x"}`, codeInvalidOperator, "operator"},
		{"base syntax before the operator", "/api/v1/percentage", `{"percent": "1", "base": "x", "operator": "x"}`, codeInvalidNumber, "base"},
		{"percent syntax before the operator", "/api/v1/percentage", `{"percent": "x", "base": "1", "operator": "x"}`, codeInvalidNumber, "percent"},
		{"percent limits without a base", "/api/v1/percentage", `{"percent": "1e40"}`, codeOperandOutOfRange, "percent"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := readProblem(t, serve(t, http.MethodPost, tt.path, tt.body))
			if p.Code != tt.wantCode || p.Field != tt.wantField {
				t.Errorf("problem = %+v, want code %q and field %q", p, tt.wantCode, tt.wantField)
			}
		})
	}
}
