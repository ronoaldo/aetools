package bigquerysync_test

import (
	"encoding/json"
	"testing"

	"ronoaldo.gopkg.net/aetools"
	"ronoaldo.gopkg.net/aetools/aestubs"
	"ronoaldo.gopkg.net/aetools/bigquerysync"

	"appengine/datastore"
)

var datastoreStats = `
[{
	"__key__": [
		"__Stat_Kind__",
		"Account"
	],
	"count": 123,
	"entity_bytes": 45678,
	"kind_name": "Account",
	"builtin_index_bytes": 45678,
	"builtin_index_count": 123,
	"composite_index_count": 0,
	"composite_index_bytes": 0,
	"timestamp": {
		"type": "date",
		"value": "2014-01-01T10:00:00-03:00"
	}
},{
	"__key__": [
		"__Stat_PropertyType_PropertyName_Kind__",
		"String_Emails_Account"
	],
	"count": 246,
	"bytes": 456789,
	"property_type": "String",
	"property_name": "Emails",
	"kind_name": "Account",
	"builtin_index_bytes": 456789,
	"builtin_index_count": 246,
	"timestamp": {
		"type": "date",
		"value": "2014-01-01T10:00:00-03:00"
	}
},{
	"__key__": [
		"__Stat_PropertyType_PropertyName_Kind__",
		"Date/Time_CreationDate_Account"
	],
	"count": 123,
	"bytes": 45678,
	"property_type": "Date/Time",
	"property_name": "CreationDate",
	"kind_name": "Account",
	"builtin_index_bytes": 45678,
	"builtin_index_count": 123,
	"timestamp": {
		"type": "date",
		"value": "2014-01-01T10:00:00-03:00"
	}
}]`

func TestDecodeStatByProperty(t *testing.T) {
	c := aestubs.NewContext(nil, t)

	err := aetools.LoadJSON(c, datastoreStats, aetools.LoadSync)
	if err != nil {
		t.Log("Unable to load fixtures")
		t.Fatal(err)
	}

	p := new(bigquerysync.StatByProperty)
	name := "Date/Time_CreationDate_Account"
	k := datastore.NewKey(c, bigquerysync.StatByPropertyKind, name, 0, nil)
	err = datastore.Get(c, k, p)
	if err != nil {
		t.Fatal(err)
	}

	if p.Count != 123 {
		t.Errorf("Unexpected p.Count: %d, expecting 123", p.Count)
	}
	if p.Bytes != 45678 {
		t.Errorf("Unexpected p.Bytes: %d, expecting 45678", p.Bytes)
	}
	if p.Type != "Date/Time" {
		t.Errorf("Unexpected p.Type: '%s', expecting 'Date/Time'", p.Type)
	}
	if p.Kind != "Account" {
		t.Errorf("Unexpected p.Account: '%s', expecting 'Account'", p.Kind)
	}
	if p.IndexBytes != 45678 {
		t.Errorf("Unexpected p.IndexBytes: %d, expecting 45678", p.IndexBytes)
	}
	if p.IndexCount != 123 {
		t.Errorf("Unexpected p.IndexCount: %d, expecting 123", p.IndexCount)
	}
	if p.Timestamp.IsZero() {
		t.Errorf("Decoded timestamp is zero")
	}
}

func TestInferTableSchema(t *testing.T) {
	c := aestubs.NewContext(nil, t)
	err := aetools.LoadJSON(c, datastoreStats, aetools.LoadSync)

	s, err := bigquerysync.SchemaForKind(c, "Account")
	if err != nil {
		t.Fatal(err)
	}
	j, _ := json.Marshal(s)
	t.Logf("Decoded schema: '%s'", string(j))

	if len(s.Fields) != 4 {
		t.Errorf("Unexpected field len: %d, expected 4", len(s.Fields))
	}
	for i, f := range s.Fields {
		if f.Name == "" {
			t.Errorf("Name of field %d is empty", i)
		}
		if f.Type == "" {
			t.Errorf("Type of field %d is empty", i)
		}
	}
}
