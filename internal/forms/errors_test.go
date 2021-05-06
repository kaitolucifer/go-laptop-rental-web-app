package forms

import (
	"net/url"
	"testing"
)

func TestErrors_Add(t *testing.T) {
	postData := url.Values{}
	form := New(postData)
	form.Errors.Add("test", "error")
}

func TestErrors_Get(t *testing.T) {
	postData := url.Values{}
	form := New(postData)
	if form.Errors.Get("test") != "" {
		t.Error("TestErrors_Get: there should be no error message")
	}
	form.Errors.Add("test", "error")
	if form.Errors.Get("test") != "error" {
		t.Error("TestErrors_Get: failed to get the error message")
	}
}
