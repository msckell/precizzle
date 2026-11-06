package api

import (
	"fmt"
	"net/http"

	"github.com/msckell/precizzle/backend/internal/calc"
)

// NewHandler returns the handler that serves every route of the API.
func NewHandler() http.Handler {
	mux := http.NewServeMux()
	operations := map[string]http.Handler{
		"/api/v1/add":        binary(calc.Add, nil),
		"/api/v1/subtract":   binary(calc.Subtract, nil),
		"/api/v1/multiply":   binary(calc.Multiply, nil),
		"/api/v1/divide":     binary(calc.Divide, divideFailures),
		"/api/v1/power":      power(),
		"/api/v1/sqrt":       sqrt(),
		"/api/v1/percentage": percentage(),
	}
	for path, handler := range operations {
		mux.Handle("POST "+path, handler)
		// A pattern with a method is more specific than one without, so this
		// one only receives the other methods.
		mux.Handle(path, methodNotAllowed("POST"))
	}
	mux.HandleFunc("/", notFound)
	return mux
}

// methodNotAllowed answers a request whose path exists but whose method does not.
func methodNotAllowed(allow string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Allow", allow)
		writeProblem(w, &problem{
			code:   codeMethodNotAllowed,
			detail: fmt.Sprintf("%s does not accept %s. Allowed: %s.", r.URL.Path, r.Method, allow),
		})
	}
}

func notFound(w http.ResponseWriter, r *http.Request) {
	writeProblem(w, &problem{code: codeNotFound, detail: fmt.Sprintf("No endpoint exists at %s.", r.URL.Path)})
}
