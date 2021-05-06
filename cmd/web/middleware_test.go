package main

import (
	"net/http"
	"testing"
)

func TestNoSurve(t *testing.T) {
	var handler http.Handler
	h := NoSurf(handler)

	switch v := h.(type) {
	case http.Handler:
		// do nothing
	default:
		t.Errorf("return type is not http.Handler: %s", v)
	}
}

func TestSessionLoad(t *testing.T) {
	var handler http.Handler
	h := SessionLoad(handler)

	switch v := h.(type) {
	case http.Handler:
		// do nothing
	default:
		t.Errorf("return type is not http.Handler: %s", v)
	}
}
