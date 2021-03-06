// Copyright 2014 Ronoaldo JLP <ronoaldo@gmail.com>
// Licensed under the Apache License, Version 2.0

package aetools

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// MarshalJSON implements the json.Marshaller interface by dumping
// the entity key and properties.
func (e *Entity) MarshalJSON() ([]byte, error) {
	m, err := e.Map()
	if err != nil {
		return nil, err
	}
	b := new(bytes.Buffer)
	enc := json.NewEncoder(b)
	err = enc.Encode(m)
	return b.Bytes(), err
}

// Map converts the entity data into a JSON compatible map.
// Usefull to manually encode the entity using different
// marshallers than the JSON built-in.
func (e *Entity) Map() (map[string]interface{}, error) {
	m := make(map[string]interface{})
	m["__key__"] = encodeKey(e.Key)

	add := func(multi bool, n string, v interface{}) {
		if multi {
			a := m[n].([]interface{})
			m[n] = append(a, v)
		} else {
			m[n] = v
		}
	}

	for _, p := range e.Properties {
		// Check if multi property is consistent so it's safe to append
		if p.Multiple {
			if _, ok := m[p.Name]; !ok {
				m[p.Name] = make([]interface{}, 0)
			} else {
				if _, ok := m[p.Name].([]interface{}); !ok {
					return nil, fmt.Errorf("aetools: %s with invalid Multiple values", p.Name)
				}
			}
		}

		switch p.Value.(type) {
		case int, int32, int64:
			if p.NoIndex {
				add(p.Multiple, p.Name, toMap("int", p.NoIndex, p.Value))
			} else {
				add(p.Multiple, p.Name, p.Value)
			}
		case float32, float64:
			if p.NoIndex {
				add(p.Multiple, p.Name, toMap("float", p.NoIndex, float(p.Value.(float64))))
			} else {
				add(p.Multiple, p.Name, float(p.Value.(float64)))
			}
		case string:
			if p.NoIndex {
				add(p.Multiple, p.Name, toMap("string", p.NoIndex, p.Value))
			} else {
				add(p.Multiple, p.Name, p.Value)
			}
		case bool:
			if p.NoIndex {
				add(p.Multiple, p.Name, toMap("bool", p.NoIndex, p.Value))
			} else {
				add(p.Multiple, p.Name, p.Value)
			}
		case *datastore.Key:
			v := toMap("key", p.NoIndex, encodeKey(p.Value.(*datastore.Key)))
			add(p.Multiple, p.Name, v)
		case appengine.BlobKey:
			v := toMap("blobkey", p.NoIndex, string(p.Value.(appengine.BlobKey)))
			add(p.Multiple, p.Name, v)
		case time.Time:
			s := p.Value.(time.Time).Format(DateTimeFormat)
			v := toMap("date", p.NoIndex, s)
			add(p.Multiple, p.Name, v)
		case []byte:
			s := base64.URLEncoding.EncodeToString(p.Value.([]byte))
			v := toMap("blob", p.NoIndex, s)
			add(p.Multiple, p.Name, v)
		default:
			if p.Value != nil {
				return nil, fmt.Errorf("aetools: invalid property value %s: %#v", p.Name, p.Value)
			}

			if p.NoIndex {
				add(p.Multiple, p.Name, toMap("", p.NoIndex, p.Value))
			} else {
				add(p.Multiple, p.Name, p.Value)
			}
		}
	}
	return m, nil
}

// KeyPath returns a representation of a Key as a string with each value
// in the key path separated by a coma.
// The key representation has all ancestors,
// but has no information about namespaces or ancestors.
func KeyPath(k *datastore.Key) string {
	b := new(bytes.Buffer)
	path := encodeKey(k)
	for i := range path {
		fmt.Fprintf(b, "%v", path[i])
		if i < len(path)-1 {
			fmt.Fprint(b, ",")
		}
	}
	return b.String()
}

type float float64

var hasDecimalPoint = regexp.MustCompile(".*[.eE].*")

func (f float) MarshalJSON() ([]byte, error) {
	var b bytes.Buffer
	if math.IsInf(float64(f), 0) || math.IsNaN(float64(f)) {
		return nil, fmt.Errorf("aetools: unsuported float value: %#v", f)
	}
	b.Write(strconv.AppendFloat(b.Bytes(), float64(f), 'g', -1, 64))
	if !hasDecimalPoint.Match(b.Bytes()) {
		b.WriteString(".0")
	}
	return b.Bytes(), nil
}

func decodeJSONPrimitiveValue(v interface{}, p *datastore.Property) error {
	switch v.(type) {
	case json.Number:
		n := v.(json.Number)
		if strings.Contains(n.String(), ".") {
			// float64
			p.Value, _ = n.Float64()
		} else {
			// int64
			p.Value, _ = n.Int64()
		}
	case string:
		p.Value = v.(string)
	case bool:
		p.Value = v.(bool)
	case nil:
		p.Value = nil
	default:
		return fmt.Errorf("Invalid primitive value: %#v", v)
	}
	return nil
}

func encodeKey(k *datastore.Key) []interface{} {
	path := make([]*datastore.Key, 0)

	tmp := k
	for tmp != nil {
		path = append(path, tmp)
		tmp = tmp.Parent()
	}

	r := make([]interface{}, 0, 2*len(path))
	for i := len(path) - 1; i >= 0; i-- {
		tmp = path[i]

		r = append(r, tmp.Kind())
		if !tmp.Incomplete() {
			if tmp.StringID() != "" {
				r = append(r, tmp.StringID())
			} else {
				r = append(r, tmp.IntID())
			}
		} else {
			r = append(r, nil)
		}
	}

	return r
}

func decodeKey(c context.Context, v interface{}) (*datastore.Key, error) {
	var result, ancestor *datastore.Key
	p, ok := v.([]interface{})
	if !ok {
		return nil, ErrInvalidKeyElement
	}

	for i := 0; i < len(p); i += 2 {
		kind := p[i].(string)
		id := p[i+1]
		switch id.(type) {
		case string:
			result = datastore.NewKey(c, kind, id.(string), 0, ancestor)
		case json.Number:
			n, err := id.(json.Number).Int64()
			if err != nil {
				return nil, invalidIDError(id)
			}
			result = datastore.NewKey(c, kind, "", n, ancestor)
		default:
			return nil, invalidIDError(id)
		}

		ancestor = result
	}

	return result, nil
}

func toMap(t string, noIndex bool, v interface{}) map[string]interface{} {
	m := make(map[string]interface{}, 3)
	m["value"] = v
	m["type"] = t
	m["indexed"] = !noIndex
	if t == "" {
		delete(m, "type")
	}
	return m
}

func decodeProperty(c context.Context, k string, v interface{}, e *Entity) error {
	var p datastore.Property
	p.Name = k

	var err error

	switch v.(type) {
	// Try to decode property object
	case map[string]interface{}:
		// Decode custom type
		m := v.(map[string]interface{})

		t, ok := m["type"]
		if !ok {
			t = "primitive"
		}

		if index, ok := m["indexed"]; ok {
			if i, ok := index.(bool); ok {
				p.NoIndex = !i
			}
		}

		switch t {
		case "key":
			key, err := decodeKey(c, m["value"])
			if err != nil {
				return err
			}
			p.Value = key
		case "blobkey":
			v, ok := m["value"].(string)
			if !ok {
				return newDecodePropertyError(k, "blobkey", v)
			}
			p.Value = appengine.BlobKey(v)
		case "blob":
			v, ok := m["value"].(string)
			if !ok {
				return newDecodePropertyError(k, "date", v)
			}
			p.Value, err = base64.URLEncoding.DecodeString(v)
			if err != nil {
				return err
			}
		case "date":
			v, ok := m["value"].(string)
			if !ok {
				return newDecodePropertyError(k, "date", v)
			}
			var dt time.Time
			dt, err = time.Parse(DateTimeFormat, v)
			if err != nil {
				return newDecodePropertyError(k, "date", err)
			}
			p.Value = dt.UTC()
		default:
			if v, ok := m["value"]; ok {
				err = decodeJSONPrimitiveValue(v, &p)
			} else {
				err = fmt.Errorf("aetools: complex property %s without 'value' attribute", k)
			}
		}

	default:
		err = decodeJSONPrimitiveValue(v, &p)
	}

	if err == nil {
		e.Properties = append(e.Properties, p)
	}
	return err
}

func newDecodePropertyError(name, ptype string, raw interface{}) error {
	return fmt.Errorf("aetools: can't decode %s, value is not %s: %s", name, ptype, raw)
}
