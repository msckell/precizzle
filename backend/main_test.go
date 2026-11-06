package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUnknownPathReturnsNotFound(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string
	}{
		{name: "root", method: http.MethodGet, path: "/"},
		{name: "unknown path", method: http.MethodPost, path: "/unknown"},
	}

	handler := newHandler()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, httptest.NewRequest(tt.method, tt.path, nil))

			if rec.Code != http.StatusNotFound {
				t.Errorf("%s %s: status = %d, want %d", tt.method, tt.path, rec.Code, http.StatusNotFound)
			}
		})
	}
}
