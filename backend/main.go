package main

import (
	"log"
	"net/http"
)

func newHandler() *http.ServeMux {
	return http.NewServeMux()
}

func main() {
	log.Fatal(http.ListenAndServe("127.0.0.1:8080", newHandler()))
}
