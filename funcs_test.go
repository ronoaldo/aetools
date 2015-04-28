package aetools

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/drhodes/golorem"

	"appengine"
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
	HTMLDesc    string    `datastore:"htmlDesc"`
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
	err = Dump(c, w, &Options{Kind: "User", GetAfterPut: true})
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Dump output: ", w)

	err = Load(c, w, &Options{
		GetAfterPut: true,
		BatchSize:   50,
	})
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
		e.Add(datastore.Property{
			Name:  "name",
			Value: fmt.Sprintf("Test Entity #%d", id),
		})
		for j := 0; j < 3; j++ {
			e.Add(datastore.Property{
				Name:     "tags",
				Value:    fmt.Sprintf("tag%d", j),
				Multiple: true,
			})
		}
		e.Add(datastore.Property{
			Name:  "active",
			Value: i%2 == 0,
		})
		e.Add(datastore.Property{
			Name:  "height",
			Value: i * 10,
		})
		entities = append(entities, e)
	}

	p := encodeKey(entities[0].Key)
	t.Logf("encodeKey: from %s to %#v", entities[0].Key, p)

	w := new(bytes.Buffer)
	err = EncodeEntities(entities, w)
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

	r, err := DecodeEntities(c, bytes.NewReader(fixture))
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

	err = Load(c, bytes.NewReader(fixture), &Options{GetAfterPut: true})
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
	b, _ := time.Parse("2006-01-02 15:04:05.999 -0700", "1986-07-19 03:00:00.000 -0000")
	if p.Birthday.UTC() != b.UTC() {
		t.Errorf("Unexpected p.Birthday: %v, expected %v", p.Birthday, b.UTC())
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

func TestBatchSizeOnDump(t *testing.T) {
	c, err := aetest.NewContext(nil)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	for _, i := range []int{10, 20, 50, 99, 100, 101} {
		t.Logf("Testing %d entities ...", i)
		if err := createSampleEntities(c, i); err != nil {
			t.Fatal(err)
		}
		w := new(bytes.Buffer)
		err := Dump(c, w, &Options{Kind: "User", PrettyPrint: false})
		if err != nil {
			t.Fatal(err)
		}
		count := strings.Count(w.String(), "__key__")
		if count != i {
			t.Errorf("Unexpected number of __key__'s %d: expected %d", count, i)
		}
		// t.Logf(w.String())
		// Check if we have all keys
		for id := 1; id <= i; id++ {
			sep := fmt.Sprintf(`["User",%d]`, id)
			occ := strings.Count(w.String(), sep)
			if occ != 1 {
				t.Errorf("Unexpected ocorrences of entity id %d: %d, expected 1", id, occ)
			}
		}
	}
}

func TestBatchSizeWhenLoading(t *testing.T) {
	c, err := aetest.NewContext(nil)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	// Zero-case check for load bounds
	if err := Load(c, strings.NewReader("[]"), LoadSync); err != nil {
		t.Errorf("Failed to load empty array: %v", err)
	}

	for _, i := range []int{100, 500, 600} {
		json := new(bytes.Buffer)
		fmt.Fprint(json, "[")
		for j := 1; j < i; j++ {
			fmt.Fprintf(json, `{
				"__key__" : ["Test%d", %d],
				"blob" : {
					"type" : "string",
					"indexed" : false,
					"value" : "%s"
				}
			},`, i, j, strings.Repeat("0", 10*1024))
		}
		fmt.Fprintf(json, `{"__key__" : ["Test%d", 0]}]`, i)
		t.Logf("Loading %d entities ...", i)
		err := Load(c, json, LoadSync)
		if err != nil {
			t.Errorf("Error loaing %d entities: %v", i, err)
		}
		if count, err := datastore.NewQuery(fmt.Sprintf("Test%d", i)).Count(c); err != nil {
			t.Errorf("Error checking the persisted entities: %v", err)
		} else if count != i {
			t.Errorf("Entity count minsmatch: %d, expected %d", count, i)
		}
	}
}

func createSampleEntities(c appengine.Context, size int) error {
	buff := make([]Entity, 0, 10)
	keys := make([]*datastore.Key, 0, 10)
	for i := 1; i <= size; i++ {
		k := datastore.NewKey(c, "User", "", int64(i), nil)
		e := Entity{Key: k}
		e.Add(datastore.Property{Name: "Title", Value: lorem.Sentence(5, 10)})
		e.Add(datastore.Property{
			Name:    "SubTitle",
			Value:   lorem.Sentence(3, 5),
			NoIndex: true,
		})
		e.Add(datastore.Property{
			Name:    "Description",
			Value:   lorem.Paragraph(3, 5),
			NoIndex: true,
		})
		e.Add(datastore.Property{Name: "Size", Value: int64(32)})
		for j := 0; j < 5; j++ {
			e.Add(datastore.Property{
				Name:     "Tags",
				Value:    lorem.Word(5, 10),
				Multiple: true,
			})
		}
		e.Add(datastore.Property{Name: "Price", Value: float64(123.45)})
		for j := 0; j < 10; j++ {
			e.Add(datastore.Property{
				Name:     "PriceHistory",
				Value:    float64(123.45) - float64(j),
				Multiple: true,
			})
		}
		e.Add(datastore.Property{Name: "Favicon", Value: icon, NoIndex: true})
		e.Add(datastore.Property{Name: "FaviconSource", Value: blobKey})
		for j := 1; j <= 3; j++ {
			e.Add(datastore.Property{
				Name:     "Friends",
				Value:    datastore.NewKey(c, "Friend", "", int64(j), k),
				Multiple: true,
			})
		}
		buff = append(buff, e)
		keys = append(keys, k)

		if len(buff) == 10 {
			_, err := datastore.PutMulti(c, keys, buff)
			if err != nil {
				return err
			}
			_ = datastore.GetMulti(c, keys, buff)

			buff = make([]Entity, 0, 10)
			keys = make([]*datastore.Key, 0, 10)
		}
	}
	if len(buff) > 0 {
		k, err := datastore.PutMulti(c, keys, buff)
		if err != nil {
			return err
		}
		_ = datastore.GetMulti(c, k, buff)
	}
	return nil
}
