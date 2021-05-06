package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandlers(t *testing.T) {
	routes := getRoutes()
	ts := httptest.NewTLSServer(routes)
	defer ts.Close()

	for _, test := range tests {
		var resp *http.Response
		var err error
		if test.method == "GET" {
			resp, err = ts.Client().Get(ts.URL + test.path)
		} else {
			resp, err = ts.Client().PostForm(ts.URL+test.path, test.data)
		}
		if err != nil {
			t.Log(err)
			t.Fatal(err)
		}

		if resp.StatusCode != test.expectedStatusCode {
			t.Errorf("for %s, expected status code %d but got %d", test.name, test.expectedStatusCode, resp.StatusCode)
		}
	}
}
