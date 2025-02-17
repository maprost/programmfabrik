package main

import (
	"net/http"

	"github.com/maprost/programmfabrik/internal"
)

func main() {
	http.HandleFunc("/", internal.MainHandler)
	http.ListenAndServe("localhost:8080", nil)

}
