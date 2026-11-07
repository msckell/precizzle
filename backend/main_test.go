package main

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestListenAddr(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
		want string
	}{
		{name: "default", env: map[string]string{}, want: "127.0.0.1:8080"},
		{name: "empty variable", env: map[string]string{"PRECIZZLE_ADDR": ""}, want: "127.0.0.1:8080"},
		{name: "custom port", env: map[string]string{"PRECIZZLE_ADDR": "127.0.0.1:8081"}, want: "127.0.0.1:8081"},
		{name: "all interfaces", env: map[string]string{"PRECIZZLE_ADDR": "0.0.0.0:8080"}, want: "0.0.0.0:8080"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getenv := func(key string) string { return tt.env[key] }
			if got := listenAddr(getenv); got != tt.want {
				t.Errorf("listenAddr = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNewHandler(t *testing.T) {
	const indexHTML = `<!doctype html><div id="root"></div>`
	const script = "console.log(1)"
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "index.html"), indexHTML)
	writeFile(t, filepath.Join(dir, "assets", "app.js"), script)

	tests := []struct {
		name        string
		staticDir   string
		method      string
		path        string
		body        string
		wantStatus  int
		wantType    string
		wantBody    string
		wantNoBody  bool
		wantProblem bool
	}{
		{name: "no static dir: root is not found", method: "GET", path: "/", wantStatus: 404, wantProblem: true},
		{name: "no static dir: API works", method: "POST", path: "/api/v1/evaluate", body: `{"expression":"0.1+0.2"}`, wantStatus: 200, wantBody: `"value":"0.3"`},
		{name: "root serves index.html", staticDir: dir, method: "GET", path: "/", wantStatus: 200, wantType: "text/html", wantBody: indexHTML},
		{name: "HEAD root has no body", staticDir: dir, method: "HEAD", path: "/", wantStatus: 200, wantType: "text/html", wantNoBody: true},
		{name: "asset is served", staticDir: dir, method: "GET", path: "/assets/app.js", wantStatus: 200, wantBody: script},
		{name: "missing file is a plain 404", staticDir: dir, method: "GET", path: "/missing.js", wantStatus: 404, wantType: "text/plain"},
		{name: "API evaluates", staticDir: dir, method: "POST", path: "/api/v1/evaluate", body: `{"expression":"0.1+0.2"}`, wantStatus: 200, wantType: "application/json", wantBody: `"value":"0.3"`},
		{name: "health", staticDir: dir, method: "GET", path: "/healthz", wantStatus: 200, wantBody: `{"status":"ok"}`},
		{name: "health rejects POST", staticDir: dir, method: "POST", path: "/healthz", wantStatus: 405, wantProblem: true},
		{name: "API rejects GET", staticDir: dir, method: "GET", path: "/api/v1/add", wantStatus: 405, wantProblem: true},
		{name: "unknown API path", staticDir: dir, method: "GET", path: "/api/v1/nope", wantStatus: 404, wantProblem: true},
		{name: "POST outside the API", staticDir: dir, method: "POST", path: "/", wantStatus: 404, wantProblem: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			if tt.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}
			rec := httptest.NewRecorder()
			newHandler(tt.staticDir).ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body %q", rec.Code, tt.wantStatus, rec.Body.String())
			}
			contentType := rec.Header().Get("Content-Type")
			if tt.wantProblem && contentType != "application/problem+json" {
				t.Errorf("Content-Type = %q, want application/problem+json", contentType)
			}
			if tt.wantType != "" && !strings.HasPrefix(contentType, tt.wantType) {
				t.Errorf("Content-Type = %q, want prefix %q", contentType, tt.wantType)
			}
			if !strings.Contains(rec.Body.String(), tt.wantBody) {
				t.Errorf("body = %q, want it to contain %q", rec.Body.String(), tt.wantBody)
			}
			if tt.wantNoBody && rec.Body.Len() != 0 {
				t.Errorf("body = %q, want empty", rec.Body.String())
			}
		})
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
