// Copyright 2014 Ronoaldo JLP <ronoaldo@gmail.com>
// Licensed under the Apache License, Version 2.0

package bigquerysync_test

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"strings"
	"testing"

	"golang.org/x/net/context"
	bigquery "google.golang.org/api/bigquery/v2"
	"google.golang.org/appengine/aetest"
	"ronoaldo.gopkg.net/aetools"
	"ronoaldo.gopkg.net/aetools/bigquerysync"
)

type TestContext interface {
	context.Context

	TestServer() *httptest.Server
	Close() error
}

type testContext struct {
	context.Context
	clean  func()
	server *httptest.Server
}

func (t *testContext) TestServer() *httptest.Server {
	return t.server
}

func (t *testContext) Close() error {
	defer t.clean()
	t.server.Close()
	return nil
}

func SetupEnv(t *testing.T) TestContext {
	c, clean, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}

	err = aetools.Load(c, strings.NewReader(SampleEntities), aetools.LoadSync)
	if err != nil {
		defer clean()
		t.Fatal(err)
	}

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := httputil.DumpRequest(r, true)
		log.Printf("Received request:\n%s\n", string(b))
		enc := json.NewEncoder(w)
		err := enc.Encode(&bigquery.TableDataInsertAllResponse{})
		log.Printf("Error writing response: %v\n", err)
	}))

	// TODO(ronoaldo): enable parallel testing.
	bigquerysync.InsertAllURL = fmt.Sprintf("%s/%%s/%%s/%%s", s.URL)

	tc := &testContext{
		Context: c,
		clean:   clean,
		server:  s,
	}
	return tc
}

var SampleEntities = `[
	{
		"__key__": ["Sample", 1],
		"Name": "Sample #1",
		"Order": 1
	},
	{
		"__key__": ["Sample", 2],
		"Name": "Sample #2",
		"Order": 2
	},
	{
		"__key__": ["Sample", 3],
		"Name": "Sample #3",
		"Order": 3
	},
	{
		"__key__": ["Log", "log-entry-1"],
		"Level": "INFO",
		"Message": "Sample log message #1"
	},
	{
		"__key__": ["Log", "log-entry-2"],
		"Level": "WARN",
		"Message": "Sample log message #2"
	},
	{
		"__key__": ["RangeTest", 1],
		"Data": "range"
	},
	{
		"__key__": ["RangeTest", 2],
		"Data": "range"
	},
	{
		"__key__": ["RangeTest", 30],
		"_scatter__": 3,
		"Data": "range"
	},
	{
		"__key__": ["RangeTest", 40],
		"Data": "range"
	},
	{
		"__key__": ["RangeTest", 50],
		"_scatter__": 2,
		"Data": "range"
	},
	{
		"__key__": ["RangeTest", 6],
		"Data": "range"
	},
	{
		"__key__": ["RangeTest", 1000],
		"_scatter__": 1,
		"Data": "range"
	},
	{
		"__key__": ["RangeTest", 1001],
		"Data": "range"
	}
]`
