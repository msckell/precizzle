package api

import (
	"bytes"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/msckell/precizzle/backend/internal/calc"
)

// TestUnexpectedCalcError checks that an error the API does not know becomes
// a 500 that hides the cause from the client and records it in the log.
func TestUnexpectedCalcError(t *testing.T) {
	var logs bytes.Buffer
	previous := log.Writer()
	log.SetOutput(&logs)
	t.Cleanup(func() { log.SetOutput(previous) })

	cause := errors.New("big.Rat state 0xdeadbeef")
	handler := operation(nil, func(fields) (calc.Result, *problem) {
		return calc.Result{}, calcProblem(cause, nil)
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/add", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", rec.Code)
	}
	p := readProblem(t, rec)
	if p.Code != codeInternalError || p.Field != "" || p.Detail == "" {
		t.Errorf("problem = %+v, want code %s with a detail and no field", p, codeInternalError)
	}
	if strings.Contains(rec.Body.String(), "0xdeadbeef") {
		t.Errorf("response %s exposes the internal error", rec.Body.String())
	}
	if !strings.Contains(logs.String(), cause.Error()) {
		t.Errorf("log %q does not record the internal error", logs.String())
	}
}
