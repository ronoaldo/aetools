package aetools

import (
	"appengine/aetest"
	"appengine/datastore"
	"bytes"
	"testing"
	"time"
)

var fixture = []byte(`[
{
	"__key__": ["Profile", 123456],
	"name": "Ronoaldo JLP",
	"height": 175,
	"active": true,
	"birthday": {
		"type": "date",
		"value": "1986-07-19 00:00:00.000 -0000"
	},
	"description": "This is a long value\nblob string",
	"htmlDesc": {
		"unindexed": true,
		"value": "<h1>This is an awesome, unindexed description"
	},
	"tags": [ "a", "b", "c" ]
}, {
	"__key__": ["IncompleteProfile", "test@example.com"],
	"name": "My Name"
}
]`)

type Profile struct {
	Name        string    `datastore:"name"`
	Description string    `datastore:"description"`
	Height      int64     `datastore:"height"`
	Birthday    time.Time `datastore:"birthday"`
	Tags        []string  `datastore:"tags"`
	Active      bool      `datastore:"active"`
	HtmlDesc    string    `datastore:"htmlDesc"`
}

func TestDecodeEntities(t *testing.T) {
	c, err := aetest.NewContext(nil)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	r, err := decodeEntities(c, bytes.NewReader(fixture))
	if err != nil {
		t.Fatal(err)
	}

	if len(r) != 2 {
		t.Errorf("Unexpected entity slice size: %d, expected 2", len(r))
	}

	t.Logf("Decoded entities:")
	for i, e := range r {
		t.Logf("> %d: %#v", i, e)
	}
}

func TestLoadFixtures(t *testing.T) {
	c, err := aetest.NewContext(nil)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	err = LoadFixtures(c, bytes.NewReader(fixture), &Options{GetAfterPut: true})
	if err != nil {
		t.Fatal(err)
	}

	// Make a query to see if the decoding populated a valid Entity
	var ancestor *datastore.Key
	k := datastore.NewKey(c, "Profile", "", 123456, ancestor)
	var p Profile
	err = datastore.Get(c, k, &p)
	if err != nil {
		t.Errorf("Unable to load entity by key. LoadFixture failed: %s", err.Error())
		t.FailNow()
	}

	if p.Name != "Ronoaldo JLP" {
		t.Errorf("Unexpected p.Name %s, expected Ronoaldo JLP", p.Name)
	}
	d := "This is a long value\nblob string"
	if p.Description != d {
		t.Errorf("Unexpected p.Description '%s', expected %s", p.Description, d)
	}
	if p.Height != 175 {
		t.Errorf("Unexpected p.Height: %d, expected 175", p.Height)
	}
	b, _ := time.Parse("2006-01-02 15:04:05.999 -0700", "1986-07-19 00:00:00.000 +0000")
	if p.Birthday != b {
		t.Errorf("Unexpected p.Birthday: %v, expected %v", p.Birthday, b)
	}
	if len(p.Tags) != 3 {
		t.Errorf("Unexpected p.Tags length: %v, expected %v", len(p.Tags), 3)
	}
	tags := []string{"a", "b", "c"}
	for i, tag := range p.Tags {
		if tag != tags[i] {
			t.Errorf("Unexpected value on p.Tags: %v, expected %v", tags[i], tag)
		}
	}
}
