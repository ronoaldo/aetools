package bigquerysync_test

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"strings"
	"testing"

	"ronoaldo.gopkg.net/aetools"
	"ronoaldo.gopkg.net/aetools/bigquerysync"

	"appengine/aetest"
)

type TestContext interface {
	aetest.Context
	TestServer() *httptest.Server
}

type testContext struct {
	aetest.Context
	server *httptest.Server
}

func (t *testContext) TestServer() *httptest.Server {
	return t.server
}

func (t *testContext) Close() error {
	defer t.Context.Close()
	t.server.Close()
	return nil
}

func SetupEnv(t *testing.T) TestContext {
	c, err := aetest.NewContext(nil)
	if err != nil {
		t.Fatal(err)
	}

	err = aetools.Load(c, strings.NewReader(SampleEntities), aetools.LoadSync)
	if err != nil {
		c.Close()
		t.Fatal(err)
	}

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := httputil.DumpRequest(r, true)
		log.Printf("Received request:\n%s\n", string(b))
	}))

	// TODO(ronoaldo): enable parallel testing.
	bigquerysync.InsertAllURL = fmt.Sprintf("%s/%%s/%%s/%%s", s.URL)

	tc := &testContext{c, s}
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
