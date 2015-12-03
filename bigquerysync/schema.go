// Copyright 2014 Ronoaldo JLP <ronoaldo@gmail.com>
// Licensed under the Apache License, Version 2.0

package bigquerysync

import (
	"fmt"
	"golang.org/x/net/context"
	"regexp"
	"sort"
	"strings"
	"time"

	"google.golang.org/api/bigquery/v2"
	"google.golang.org/appengine/datastore"
)

const (
	StatByPropertyKind = "__Stat_PropertyType_PropertyName_Kind__"
	StatByKindKind     = "__Stat_Kind__"
)

// StatByProperty holds the statistic information about an
// entity property.
type StatByProperty struct {
	Count      int64     `datastore:"count"`
	Bytes      int64     `datastore:"bytes"`
	Type       string    `datastore:"property_type"`
	Name       string    `datastore:"property_name"`
	Kind       string    `datastore:"kind_name"`
	IndexBytes int64     `datastore:"builtin_index_bytes"`
	IndexCount int64     `datastore:"builtin_index_count"`
	Timestamp  time.Time `datastore:"timestamp"`
}

// StatByKind holds the statistic information about an entity kind.
type StatByKind struct {
	Count               int64     `datastore:"count"`
	EntityBytes         int64     `datastore:"entity_bytes"`
	Kind                string    `datastore:"kind_name"`
	IndexBytes          int64     `datastore:"builtin_index_bytes"`
	IndexCount          int64     `datastore:"builtin_index_count"`
	CompositeIndexBytes int64     `datastore:"composite_index_count"`
	CompositeIndexCount int64     `datastore:"composite_index_bytes"`
	Timestamp           time.Time `datastore:"timestamp"`
}

// SchemaForKind guess the schema based on the datastore
// statistics for the specified entity kind.
func SchemaForKind(c context.Context, kind string) (*bigquery.TableSchema, error) {
	var (
		k         *datastore.Key
		err       error
		kindStats *StatByKind
	)
	schema := bigquery.TableSchema{
		Fields: make([]*bigquery.TableFieldSchema, 0),
	}

	// Query for kind stats
	k = datastore.NewKey(c, StatByKindKind, kind, 0, nil)
	kindStats = new(StatByKind)
	err = datastore.Get(c, k, kindStats)
	if err != nil && !missingFieldErr(err) {
		return nil, fmt.Errorf("no stats for '%s': %s", kind, err.Error())
	}
	// Parse fields
	q := datastore.NewQuery(StatByPropertyKind).
		Filter("kind_name =", kind)
	for it := q.Run(c); ; {
		s := new(StatByProperty)
		k, err = it.Next(s)
		if err == datastore.Done {
			break
		}
		if err != nil && !missingFieldErr(err) {
			err := fmt.Errorf("can't load property stats %s: %s", kind, err.Error())
			return nil, err
		}
		fName := MakeFieldName(s.Name)
		if !containsField(&schema, fName) {
			f := new(bigquery.TableFieldSchema)
			f.Name = fName

			switch s.Type {
			case "Blob", "BlobKey", "Category", "Email", "IM", "Key", "Link",
				"PhoneNumber", "PostalAddress", "Rating", "ShortBlob", "String":
				f.Type = "STRING"
			case "Date/Time":
				f.Type = "TIMESTAMP"
			case "Boolean":
				f.Type = "BOOLEAN"
			case "Float":
				f.Type = "FLOAT"
			case "Integer":
				f.Type = "INTEGER"
			}
			if s.Count > kindStats.Count {
				// More property values than entities: must be repeated
				// Repeated are serialized as json strings
				f.Type = "STRING"
			}
			if f.Type != "" {
				schema.Fields = append(schema.Fields, f)
			}
		}
	}
	// Include key
	schema.Fields = append(schema.Fields, &bigquery.TableFieldSchema{
		Name: "__key__",
		Type: "STRING",
	})
	schema.Fields = append(schema.Fields, &bigquery.TableFieldSchema{
		Name: "__timestamp__",
		Type: "TIMESTAMP",
	})
	sort.Sort(byName(schema.Fields))
	return &schema, nil
}

// byName implements the sort.Interface ordering schema fields by name.
type byName []*bigquery.TableFieldSchema

func (b byName) Len() int           { return len(b) }
func (b byName) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b byName) Less(i, j int) bool { return b[i].Name < b[j].Name }

// invalidFieldChars is a regexp to filter out the invalid chars in field names.
var invalidFieldChars = regexp.MustCompile("[^a-zA-Z0-9_]")

// MakeFieldName returns a string replacing invalid field name chars by "_".
func MakeFieldName(propName string) string {
	f := invalidFieldChars.ReplaceAllString(propName, "_")
	return f
}

// missingFieldError checks if the given error is a missing struct field error.
func missingFieldErr(err error) bool {
	return strings.Contains(err.Error(), "no such struct field")
}

// containsField Checks if we have a field detected already
func containsField(s *bigquery.TableSchema, n string) bool {
	for _, f := range s.Fields {
		if f.Name == n {
			return true
		}
	}
	return false
}
