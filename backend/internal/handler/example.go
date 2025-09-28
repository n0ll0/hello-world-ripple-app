package handler

import (
	"net/http"
)

// ExampleHandler is a placeholder for your HTTP handlers.
func ExampleHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("This is an example handler."))
}
