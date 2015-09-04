// Copyright 2014 Ronoaldo JLP <ronoaldo@gmail.com>
// Licensed under the Apache License, Version 2.0

package aetools

import (
	"testing"

	"ronoaldo.gopkg.net/aetools/aestubs"

	"appengine/datastore"
)

func TestKeyPath(t *testing.T) {
	c := aestubs.NewContext(nil, t)
	defer c.Clean()

	keys := []*datastore.Key{
		datastore.NewKey(c, "Incomplete", "", 0, nil),
		datastore.NewKey(c, "WithID", "", 1, nil),
		datastore.NewKey(c, "WithName", "Name", 0, nil),
		datastore.NewKey(c, "WithAncestor", "", 1, datastore.NewKey(c, "Ancestor", "Name", 0, nil)),
	}

	for _, k := range keys {
		t.Logf("KeyPath of '%s' => `%s`", k, KeyPath(k))
	}
}

func TestMarshalFloat(t *testing.T) {
	cases := []struct {
		Value    float64
		Expected string
	}{
		{1, "1.0"},
		{1.01, "1.01"},
		{1e7, "1e+07"},
	}
	for i, c := range cases {
		b, err := float(c.Value).MarshalJSON()
		if err != nil {
			t.Fatal(err)
		}
		if string(b) != c.Expected {
			t.Errorf("%d: Unexpected float value %s, expected: %s", i, string(b), c.Expected)
		}
	}
}
