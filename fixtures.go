package aetools

import (
	"appengine"
	"appengine/datastore"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
	"time"
)

const (
	DateTimeFormat = "2006-01-02 15:04:05.000 -0700"
)

var (
	ErrInvalidRootElement       = errors.New("aetools: root object is not an array.")
	ErrInvalidElementType       = errors.New("aetools: element is not a json object.")
	ErrInvalidPropertiesElement = errors.New("aetools: element's properties field is invalid.")
	ErrNoKeyElement             = errors.New("aetools: element's key field is not present.")
	ErrInvalidKeyElement        = errors.New("aetools: element's key field is invalid.")
)

type Entity struct {
	Key        *datastore.Key
	Properties datastore.PropertyList
}

func LoadFixture(c appengine.Context, r io.Reader) error {
	entities, err := DecodeFixture(c, r)
	if err != nil {
		return err
	}

	keys := make([]*datastore.Key, 0, len(entities))
	values := make([]datastore.PropertyList, 0, len(entities))

	for _, e := range entities {
		keys = append(keys, e.Key)
		values = append(values, e.Properties)
	}

	keys, err = datastore.PutMulti(c, keys, values)
	if err != nil {
		return err
	}

	return nil
}

func DecodeFixture(c appengine.Context, r io.Reader) ([]Entity, error) {
	d := json.NewDecoder(r)
	d.UseNumber()

	var f interface{}
	err := d.Decode(&f)
	if err != nil {
		return nil, err
	}

	a, ok := f.([]interface{})
	if !ok {
		return nil, ErrInvalidRootElement
	}

	result := make([]Entity, 0)

	for _, f := range a {
		m, ok := f.(map[string]interface{})
		if !ok {
			return nil, ErrInvalidElementType
		}

		e, err := DecodeEntity(c, m)
		if err != nil {
			return nil, err
		}

		result = append(result, *e)
	}

	return result, nil
}

func DecodeEntity(c appengine.Context, m map[string]interface{}) (*Entity, error) {
	var e Entity
	var err error

	if v, ok := m["key"]; ok {
		e.Key, err = DecodeKeyPath(c, v)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, ErrNoKeyElement
	}

	prop, ok := m["properties"].(map[string]interface{})
	if !ok {
		return nil, ErrInvalidPropertiesElement
	}

	for k, v := range prop {
		switch v.(type) {
		case []interface{}:
			l := v.([]interface{})
			for _, v := range l {
				err = DecodeProperty(k, v, &e)
				if err != nil {
					return nil, err
				}
				e.Properties[len(e.Properties)-1].Multiple = true
			}
		default:
			err = DecodeProperty(k, v, &e)
			if err != nil {
				return nil, err
			}
		}
	}

	return &e, nil
}

func DecodeProperty(k string, v interface{}, e *Entity) error {
	var p datastore.Property
	p.Name = k

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

	case map[string]interface{}:
		// Decode custom type
		m := v.(map[string]interface{})

		k, ok := m["type"].(string)
		if !ok {
			return ErrInvalidPropertiesElement
		}

		switch k {
		case "date":
			v, ok := m["value"].(string)
			if !ok {
				return ErrInvalidPropertiesElement
			}
			t, err := time.Parse(DateTimeFormat, v)
			if err != nil {
				return ErrInvalidPropertiesElement
			}
			p.Value = t
		default:
			return ErrInvalidPropertiesElement
		}
	default:
		return ErrInvalidPropertiesElement
	}

	e.Properties = append(e.Properties, p)
	return nil
}

func DecodeKeyPath(c appengine.Context, v interface{}) (*datastore.Key, error) {
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
				return nil, invalidIdError(id)
			}
			result = datastore.NewKey(c, kind, "", n, ancestor)
		default:
			return nil, invalidIdError(id)
		}

		ancestor = result
	}

	log.Printf("Decoded key %#v", *result)
	return result, nil
}

func invalidIdError(id interface{}) error {
	return errors.New(fmt.Sprintf("aetest: invalid key id/name '%v' (type %T)", id))
}
