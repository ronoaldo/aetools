package aestubs

import (
	"appengine"
	"appengine/datastore"
	"net/http"
	"testing"
)

const (
	testAppID   = "testapp"
	fqTestAppID = "dev~testapp"
)

func TestAppID(t *testing.T) {
	c := NewContext(nil, t)
	id := appengine.AppID(c)
	if id != testAppID {
		t.Errorf("Unexpected AppID: %s, expected %s.", id, testAppID)
	}

	id = c.FullyQualifiedAppID() // Internal use only, but part of the API.
	if id != fqTestAppID {
		t.Errorf("Unexpected qualified AppID: %s, expected %s.", id, testAppID)
	}
}

func TestRequest(t *testing.T) {
	c := NewContext(nil, t)
	req := c.Request() // Internal use only, but part of the API
	if req == nil {
		t.Errorf("Nil request returned from c.Request()")
	} else if _, ok := req.(*http.Request); !ok {
		t.Errorf("Value returned by c.Request() is not *http.Request: %v", req)
	}
}

func TestDatastoreNewKey(t *testing.T) {
	c := NewContext(nil, t)
	k := datastore.NewKey(c, "MyKind", "", 0, nil)
	if k == nil {
		t.Errorf("Nil key returned by NewKey")
	}
}
