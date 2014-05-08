package aetools

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	"appengine"

	"github.com/drhodes/golorem"

	"appengine/aetest"
	"appengine/datastore"
)

type Profile struct {
	Name        string    `datastore:"name"`
	Description string    `datastore:"description"`
	Height      int64     `datastore:"height"`
	Birthday    time.Time `datastore:"birthday"`
	Tags        []string  `datastore:"tags"`
	Active      bool      `datastore:"active"`
	HtmlDesc    string    `datastore:"htmlDesc"`
}

func TestEndToEndTest(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	c, err := aetest.NewContext(nil)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	err = createSampleEntities(c, 3)
	if err != nil {
		t.Fatal(err)
	}

	w := new(bytes.Buffer)
	err = DumpFixtures(c, w, &DumpOptions{"User", true})
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Dump output: ", w)

	err = LoadFixtures(c, w, &Options{true})
	if err != nil {
		t.Fatal(err)
	}
}

func TestEncodeEntities(t *testing.T) {
	c, err := aetest.NewContext(nil)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	parent := datastore.NewKey(c, "Parent", "parent-1", 0, nil)

	entities := make([]Entity, 0, 10)
	for i := 0; i < 10; i++ {
		id := i + 1

		k := datastore.NewKey(c, "Test", "", int64(id), nil)
		if i%2 == 0 {
			k = datastore.NewKey(c, "Test", "", int64(id), parent)
		}

		e := Entity{Key: k}
		e.AddProperty(datastore.Property{
			Name:  "name",
			Value: fmt.Sprintf("Test Entity #%d", id),
		})
		for j := 0; j < 3; j++ {
			e.AddProperty(datastore.Property{
				Name:     "tags",
				Value:    fmt.Sprintf("tag%d", j),
				Multiple: true,
			})
		}
		e.AddProperty(datastore.Property{
			Name:  "active",
			Value: i%2 == 0,
		})
		e.AddProperty(datastore.Property{
			Name:  "height",
			Value: i * 10,
		})
		entities = append(entities, e)
	}

	p := encodeKey(entities[0].Key)
	t.Logf("encodeKey: from %s to %#v", entities[0].Key, p)

	w := new(bytes.Buffer)
	err = encodeEntities(entities, w)
	if err != nil {
		t.Fatal(err)
	}

	json := w.String()
	t.Logf("JSON encoded entities: %s", json)
	attrs := []string{"name", "tags", "active", "height"}
	for _, a := range attrs {
		if !strings.Contains(json, a) {
			t.Errorf("Invalid JSON string: missing attribute %s.", a)
		}
	}
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

func createSampleEntities(c appengine.Context, size int) error {
	for i := 1; i <= 10; i++ {
		k := datastore.NewKey(c, "User", "", int64(i), nil)
		e := Entity{Key: k}
		e.AddProperty(datastore.Property{Name: "Title", Value: lorem.Sentence(5, 10)})
		e.AddProperty(datastore.Property{
			Name:    "SubTitle",
			Value:   lorem.Sentence(3, 5),
			NoIndex: true,
		})
		e.AddProperty(datastore.Property{
			Name:    "Description",
			Value:   lorem.Paragraph(3, 5),
			NoIndex: true,
		})
		e.AddProperty(datastore.Property{Name: "Size", Value: int64(32)})
		for j := 0; j < 5; j++ {
			e.AddProperty(datastore.Property{
				Name:     "Tags",
				Value:    lorem.Word(5, 10),
				Multiple: true,
			})
		}
		e.AddProperty(datastore.Property{Name: "Price", Value: float64(123.45)})
		for j := 0; j < 10; j++ {
			e.AddProperty(datastore.Property{
				Name:     "PriceHistory",
				Value:    float64(123.45) - float64(j),
				Multiple: true,
			})
		}
		e.AddProperty(datastore.Property{Name: "Favicon", Value: icon, NoIndex: true})
		e.AddProperty(datastore.Property{Name: "FaviconSource", Value: blobKey})
		for j := 0; j < 3; j++ {
			e.AddProperty(datastore.Property{
				Name:     "Friends",
				Value:    datastore.NewKey(c, "Friend", "", int64(j), k),
				Multiple: true,
			})
		}

		k, err := datastore.Put(c, k, &e)
		if err != nil {
			return err
		}
		_ = datastore.Get(c, k, &e)
	}
	return nil
}
