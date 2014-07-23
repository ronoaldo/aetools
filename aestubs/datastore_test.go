package aestubs

import (
	"appengine"
	"appengine/datastore"
	"fmt"
	"testing"
)

type Entity struct {
	A string
	B int
}

func TestPut(t *testing.T) {
	ds := NewDatastoreStub()
	c := NewContext(nil, t).Stub(Datastore, ds)

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
	if ds.Length() != 1 {
		t.Errorf("Internal error: datastore length != 1: %d", ds.Length())
	}

	k, err = datastore.Put(c, k, &Entity{"Test", 2})
	if err != nil {
		t.Errorf("Unexpected error in datastore.Put call, with complete key: %v", err)
	}
	if k.IntID() != intID {
		t.Errorf("Put with complete keys changes ID: %d, expected %d", k.IntID(), intID)
	}
	if ds.Length() != 1 {
		t.Errorf("Internal error after entity update: datastore length != 1: %d", ds.Length())
	}
}

func TestGet(t *testing.T) {
	ds := NewDatastoreStub()
	c := NewContext(nil, t).Stub(Datastore, ds)

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
	ds := NewDatastoreStub()
	c := NewContext(nil, t).Stub(Datastore, ds)

	keys, vals := makeSampleEntities(c)
	keys, err := datastore.PutMulti(c, keys, vals)
	if err != nil {
		t.Errorf("Unexpected error returned in datastore.Put: %v", err)
	}

	// Internal checks
	if ds.Length() != len(vals) {
		t.Errorf("Internal error: unexpected datastore length: %d, expected %d", ds.Length(), len(vals))
	}
}

func TestGetMulti(t *testing.T) {
	ds := NewDatastoreStub()
	c := NewContext(nil, t).Stub(Datastore, ds)

	keys, vals := makeSampleEntities(c)
	expected := len(keys)
	keys, err := datastore.PutMulti(c, keys, vals)
	if err != nil {
		t.Errorf("Unexpected error returned in datastore.Put: %v", err)
	}

	err = datastore.GetMulti(c, keys, vals)
	if err != nil {
		t.Errorf("Unexpected error returned in datastore.Get: %v", err)
	}

	if len(keys) != expected || len(vals) != expected {
		t.Errorf("Invalid result from batch get: %d, %d; expected %d", len(keys), len(vals), expected)
	}

	keys = append(keys, datastore.NewKey(c, "KeyNotFound", "", int64(1), nil))
	vals = append(vals, &Entity{})
	err = datastore.GetMulti(c, keys, vals)
	if err == nil {
		t.Errorf("Expecting a datastore error for non existent entity")
	}
}

func makeSampleEntities(c appengine.Context) ([]*datastore.Key, []*Entity) {
	keys := make([]*datastore.Key, 0)
	vals := make([]*Entity, 0)
	for i := 1; i <= 10; i++ {
		keys = append(keys, datastore.NewKey(c, "TestMultiKind", "", int64(i), nil))
		vals = append(vals, &Entity{fmt.Sprintf("Test Entity %d", i), i})
	}

	return keys, vals
}
