package main

import (
	"bytes"
	"net/http"
)

type responseWriterCapture struct {
	http.ResponseWriter
	body   bytes.Buffer
	status int
}

func (rw *responseWriterCapture) WriteHeader(statusCode int) {
	rw.status = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *responseWriterCapture) Write(b []byte) (int, error) {
	rw.body.Write(b)
	return rw.ResponseWriter.Write(b)
}
