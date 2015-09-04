// Copyright 2014 Ronoaldo JLP <ronoaldo@gmail.com>
// Licensed under the Apache License, Version 2.0

package aestubs

import (
	"appengine/urlfetch"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestUrlfetch(t *testing.T) {
	c := NewContext(nil, t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.Copy(w, r.Body)
	}))
	defer server.Close()

	client := urlfetch.Client(c)
	resp, err := client.Get(server.URL)
	if err != nil {
		t.Errorf("Error running client.Get: %v", err)
	}
	if resp != nil {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("Error reading response body: %v", err)
		}
		if len(b) != 0 {
			t.Errorf("Unexpected response length for GET: %d, expected 0", len(b))
		}
	}

	payload := "Hello, World!"
	resp, err = client.Post(server.URL, "text/plain", strings.NewReader(payload))
	if err != nil {
		t.Errorf("Error running client.Post: %v", err)
	}
	if resp != nil {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("Error reading the response body")
		}
		if len(b) != len([]byte(payload)) {
			t.Errorf("Unexpected response size: %d, expected %d", len(b), len([]byte(payload)))
		}
	}
}
