package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealth(t *testing.T) {
	for _, method := range []string{http.MethodGet, http.MethodHead} {
		t.Run(method, func(t *testing.T) {
			rec := httptest.NewRecorder()
			NewHandler().ServeHTTP(rec, httptest.NewRequest(method, "/healthz", nil))
			if rec.Code != http.StatusOK {
				t.Errorf("status = %d, want 200", rec.Code)
			}
			if got := rec.Header().Get("Content-Type"); got != "application/json" {
				t.Errorf("Content-Type = %q, want application/json", got)
			}
			if method == http.MethodGet && rec.Body.String() != `{"status":"ok"}` {
				t.Errorf("body = %s, want {\"status\":\"ok\"}", rec.Body.String())
			}
		})
	}
}

func TestHealthRejectsOtherMethods(t *testing.T) {
	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions} {
		t.Run(method, func(t *testing.T) {
			rec := httptest.NewRecorder()
			NewHandler().ServeHTTP(rec, httptest.NewRequest(method, "/healthz", nil))
			if p := readProblem(t, rec); p.Code != codeMethodNotAllowed {
				t.Errorf("problem = %+v, want %s", p, codeMethodNotAllowed)
			}
			if got := rec.Header().Get("Allow"); got != "GET, HEAD" {
				t.Errorf("Allow = %q, want GET, HEAD", got)
			}
		})
	}
}
