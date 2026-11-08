// Command backend serves the Precizzle API.
package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/msckell/precizzle/backend/internal/api"
)

// defaultAddr keeps the development server off the network (D27).
const defaultAddr = "127.0.0.1:8080"

func main() {
	srv := &http.Server{
		Addr:              listenAddr(os.Getenv),
		Handler:           newHandler(os.Getenv("PRECIZZLE_STATIC_DIR")),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}

// listenAddr returns the address in PRECIZZLE_ADDR, or the default when it is unset or empty (D33).
func listenAddr(getenv func(string) string) string {
	if addr := getenv("PRECIZZLE_ADDR"); addr != "" {
		return addr
	}
	return defaultAddr
}

// newHandler returns the API handler and, when staticDir is not empty, also
// serves the frontend build from it, so the Docker image needs one server (D88).
func newHandler(staticDir string) http.Handler {
	apiHandler := api.NewHandler()
	if staticDir == "" {
		return apiHandler
	}
	files := http.FileServerFS(os.DirFS(staticDir))
	mux := http.NewServeMux()
	// Without the exact "/api" pattern, the subtree pattern would redirect "/api" to "/api/".
	mux.Handle("/api", apiHandler)
	mux.Handle("/api/", apiHandler)
	mux.Handle("/healthz", apiHandler)
	// A "GET /" pattern would conflict with "/api/", so the method is checked here.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet || r.Method == http.MethodHead {
			files.ServeHTTP(w, r)
			return
		}
		apiHandler.ServeHTTP(w, r)
	})
	return mux
}
