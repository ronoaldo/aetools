// Copyright 2014 Ronoaldo JLP <ronoaldo@gmail.com>
// Licensed under the Apache License, Version 2.0

package aestubs

import (
	"fmt"
	"testing"

	"appengine"
	"appengine/datastore"
)

type Entity struct {
	A string
	B int
}

func (e *Entity) String() string {
	return fmt.Sprintf("&TestEntity{A:'%s', B:%d}", e.A, e.B)
}

func TestPut(t *testing.T) {
	c := NewContext(nil, t)
	ds := c.Stub(Datastore).(*DatastoreStub)

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
	c := NewContext(nil, t)

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
	ds := c.Stub(Datastore).(*DatastoreStub)

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
	c := NewContext(nil, t)

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

func TestAllocateIDs(t *testing.T) {
	c := NewContext(nil, t)
	anc := datastore.NewKey(c, "Ancestor", "", 1, nil)
	cases := []struct {
		anc  *datastore.Key
		n    int
		low  int64
		high int64
	}{
		{nil, 1, 1, 2},
		{nil, 1, 2, 3},
		{nil, 4, 3, 7},
		{anc, 1, 7, 8},
		{anc, 1, 8, 9},
	}

	for _, test := range cases {
		l, h, err := datastore.AllocateIDs(c, "Test", test.anc, test.n)
		if err != nil {
			t.Errorf("Unexpected error alocating IDs for %v, %d: %v", test.anc, test.n, err)
		}
		if test.low != l {
			t.Errorf("Unexpected low value %d, expecting %d (%v, %d)", l, test.low, test.anc, test.n)
		}
		if test.high != h {
			t.Errorf("Unexpected high value %d, expecting %d (%v, %d)", h, test.high, test.anc, test.n)
		}
	}
}

func TestRunQuery(t *testing.T) {
	c := NewContext(nil, t)
	keys, vals := makeSampleEntities(c)
	keys, err := datastore.PutMulti(c, keys, vals)
	if err != nil {
		t.Errorf("Unable to setup test data: %v", err)
	}

	cases := []struct {
		query *datastore.Query
		size  int
	}{
		{datastore.NewQuery("Test"), 10},
		{datastore.NewQuery("Test").Filter("A =", "Test Entity 1"), 1},
		{datastore.NewQuery("Test").Filter("A =", "Test Entity 1").Filter("B=", 1), 1},
	}
	for i, tc := range cases {
		result := make([]*Entity, 0)
		k, err := tc.query.GetAll(c, &result)
		if err != nil {
			t.Errorf("Error running query %d (%s): %v", i, tc.query, err)
			continue
		}
		if len(k) != len(result) {
			t.Errorf("Keys returned differ from entity count: %d != %d", len(k), len(result))
		}
		t.Logf("Test query #%02d:\n\tquery => %v, \n\tresult => %v", i, tc.query, result)
		if len(result) != tc.size {
			t.Errorf("Invalid query results %d, expected %d", len(result), tc.size)
		}
	}
}

func makeSampleEntities(c appengine.Context) ([]*datastore.Key, []*Entity) {
	keys := make([]*datastore.Key, 0)
	vals := make([]*Entity, 0)
	for i := 1; i <= 10; i++ {
		keys = append(keys, datastore.NewKey(c, "Test", "", int64(i), nil))
		vals = append(vals, &Entity{fmt.Sprintf("Test Entity %d", i), i})
	}

	return keys, vals
}
