package aestubs

import (
	"appengine/datastore"
	"fmt"
	"testing"
)

type Entity struct {
	A string
	B int
}

func TestPut(t *testing.T) {
	c := NewContext(nil, t)
	defer c.Clean()
	k := datastore.NewKey(c, "MyKind", "", 0, nil)
	k, err := datastore.Put(c, k, &Entity{"Test", 1})
	if err != nil {
		t.Errorf("Unexpected error in datastore.Put call: %v", err)
	}
	t.Logf("New entity put key: %#v", k)
	intID := k.IntID()
	if intID == 0 {
		t.Errorf("New entity put returned zero id")
	}

	// Internal checks
	ds := stubs[DatastoreService].(*datastoreStub)
	if ds.length() != 1 {
		t.Errorf("Internal error: datastore length != 1: %d", ds.length())
	}

	k, err = datastore.Put(c, k, &Entity{"Test", 2})
	if err != nil {
		t.Errorf("Unexpected error in datastore.Put call, with complete key: %v", err)
	}
	if k.IntID() != intID {
		t.Errorf("Put with complete keys changes ID: %d, expected %d", k.IntID(), intID)
	}
	if ds.length() != 1 {
		t.Errorf("Internal error after entity update: datastore length != 1: %d", ds.length())
	}
}

func TestGet(t *testing.T) {
	c := NewContext(nil, t)
	defer c.Clean()
	k := datastore.NewKey(c, "MyKind", "", 0, nil)
	expected := &Entity{"GetTest", 123456}
	k, err := datastore.Put(c, k, &Entity{"GetTest", 123456})
	if err != nil {
		t.Fatalf("Unexpected error in datastore.Put: %v", err)
	}

	got := new(Entity)
	err = datastore.Get(c, k, got)
	if err != nil {
		t.Errorf("Unexpected error in datastore.Get: %v", err)
	}
	if got.A != expected.A {
		t.Errorf("Unexpected property value e.A: %s, expected %s", got.A, expected.A)
	}
	if got.B != expected.B {
		t.Errorf("Unexpected properly value e.B: %d, expected %d", got.B, expected.B)
	}
}

func TestPutMulti(t *testing.T) {
	c := NewContext(nil, t)
	defer c.Clean()
	keys := make([]*datastore.Key, 0)
	vals := make([]*Entity, 0)
	for i := 1; i <= 10; i++ {
		keys = append(keys, datastore.NewKey(c, "TestMultiKind", "", int64(i), nil))
		vals = append(vals, &Entity{fmt.Sprintf("Test Entity %d", i), i})
	}
	keys, err := datastore.PutMulti(c, keys, vals)
	if err != nil {
		t.Errorf("Unexpected error returned in datastore.Put: %v", err)
	}

	// Internal checks
	ds := stubs[DatastoreService].(*datastoreStub)
	if ds.length() != len(vals) {
		t.Logf("Datastore state: %s", ds.dump())
		t.Errorf("Internal error: unexpected datastore length: %d, expected %d", ds.length(), len(vals))
	}
}
