package main

import (
	"log"
	"net/http"
)

func newHandler() *http.ServeMux {
	return http.NewServeMux()
}

func main() {
	log.Fatal(http.ListenAndServe(":8080", newHandler()))
}
