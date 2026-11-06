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
		Handler:           api.NewHandler(),
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
