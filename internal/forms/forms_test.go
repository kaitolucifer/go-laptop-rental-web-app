package forms

import (
	"net/url"
	"testing"
)

func TestNew(t *testing.T) {
	postData := url.Values{}
	var form interface{} = New(postData)
	form, ok := form.(*Form)
	if !ok {
		t.Error("TestNew failed")
	}
}

func TestForm_Has(t *testing.T) {
	postData := url.Values{}
	form := New(postData)
	if form.Has("a") {
		t.Error("TestForm_Has: shouldn't hava form data")
	}

	field := "test"
	postData.Add(field, "test")
	form = New(postData)
	if !form.Has(field) {
		t.Error("TestForm_Has: should hava form data")
	}

}

func TestForm_Valid(t *testing.T) {
	postData := url.Values{}
	form := New(postData)

	isvalid := form.Valid()
	if !isvalid {
		t.Error("TestForm_Valid: got invalid when should have been valid")
	}
}

func TestForm_Required(t *testing.T) {
	postData := url.Values{}
	form := New(postData)

	form.Required("a", "b", "c")
	if form.Valid() {
		t.Error("TestForm_Required: got valid when should have been invalid")
	}

	postData = url.Values{}
	postData.Add("a", "a")
	postData.Add("b", "b")
	postData.Add("c", "c")

	form = New(postData)
	form.Required("a", "b", "c")
	if !form.Valid() {
		t.Error("TestForm_Required: got invalid when should have been valid")
	}
}

func TestForm_MinLength(t *testing.T) {
	postData := url.Values{}
	field := "test"
	value := "55555"
	postData.Add(field, value)
	form := New(postData)
	if !form.IsAboveMinLength(field, len(value)-1) {
		t.Error("TestForm_MinLength: got false when should be true")
	}
	if form.IsAboveMinLength(field, len(value)+1) {
		t.Error("TestForm_MinLength: got true when should be false")
	}

}

func TestForm_IsEmail(t *testing.T) {
	postData := url.Values{}
	postData.Add("email", "invalid-email")
	form := New(postData)
	form.IsEmail("email")
	if form.Valid() {
		t.Error("TestForm_IsEmail: got valid when it's an invalid email")
	}
	postData.Set("email", "valid@valid.com")
	form = New(postData)
	if !form.Valid() {
		t.Error("TestForm_IsEmail: got invalid when it's an valid email")
	}
}
