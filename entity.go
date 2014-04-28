package aetools

import (
	"appengine"
	"appengine/datastore"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"
)

// Entity is a small wrapper around datastore.PropertyList
// to also hold the a *datastore.Key.
type Entity struct {
	Key        *datastore.Key
	Properties datastore.PropertyList
}

// Load decodes all properties into Entity.Propertioes object,
// implementing the datastore.PropertyLoadSaver interface.
func (e *Entity) Load(c <-chan datastore.Property) error {
	return e.Properties.Load(c)
}

// Save encodes all properties from Entity.Properties object,
// implmenting the datastore.PropetyLoadSaver interface.
func (e *Entity) Save(c chan<- datastore.Property) error {
	return e.Properties.Save(c)
}

// AddProperty append p to the Properties attribute.
func (e *Entity) AddProperty(p datastore.Property) {
	e.Properties = append(e.Properties, p)
}

// GetProperty returns the property value of the given name.
// If no property with that name exists, returns nil. It also
// returns nil if the property Value attribute is nil.
func (e *Entity) GetProperty(name string) interface{} {
	for _, p := range e.Properties {
		if p.Name == name {
			return p.Value
		}
	}
	return nil
}

// GetIntProperty returns the int value of the named property,
// and returns the zero value (0) if the property is not found,
// if its value is nil or if its type is not int, int32 or int64.
func (e *Entity) GetIntProperty(name string) int64 {
	v := e.GetProperty(name)
	if v == nil {
		return 0
	}

	switch v.(type) {
	case int, int32, int64:
		return v.(int64)
	default:
		return 0
	}
}

// GetFloatProperty returns the string value of the named property,
// and returns the zero value (0.0) if the property is not found,
// if its value is nil or if its type is not float32 or float64.
func (e *Entity) GetFloatProperty(name string) float64 {
	v := e.GetProperty(name)
	if v == nil {
		return 0.0
	}

	switch v.(type) {
	case float32, float64:
		return v.(float64)
	default:
		return 0.0
	}
}

// GetStringProperty returns the string value of the named property,
// and returns the zero value ("") if the property is not found,
// if its value is nil or if its type is not string.
func (e *Entity) GetStringProperty(name string) string {
	v := e.GetProperty(name)
	if v == nil {
		return ""
	}

	switch v.(type) {
	case string:
		return v.(string)
	default:
		return ""
	}
}

// GetBoolProperty returns the string value of the named property,
// and returns the zero value (false) if the property is not found,
// if its value is nil, or if its type is not bool.
func (e *Entity) GetBoolProperty(name string) bool {
	v := e.GetProperty(name)
	if v == nil {
		return false
	}

	switch v.(type) {
	case bool:
		return v.(bool)
	default:
		return false
	}
}

func (e *Entity) MarshalJSON() ([]byte, error) {
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
		// Check if multi property is consistent so it's safe to a
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
				add(p.Multiple, p.Name, toMap("float", p.NoIndex, p.Value))
			} else {
				add(p.Multiple, p.Name, p.Value)
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
			return nil, fmt.Errorf("aetools: invalid property value %s: %#v", p.Name, p.Value)
		}
	}

	b := new(bytes.Buffer)
	enc := json.NewEncoder(b)
	err := enc.Encode(m)

	return b.Bytes(), err
}

func toMap(t string, noIndex bool, v interface{}) map[string]interface{} {
	var m map[string]interface{}
	m["value"] = v
	m["type"] = t
	m["indexed"] = !noIndex
	return m
}
