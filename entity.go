// Copyright 2014 Ronoaldo JLP <ronoaldo@gmail.com>
// Licensed under the Apache License, Version 2.0

package aetools

import (
	"google.golang.org/appengine/datastore"
)

// Entity is a small wrapper around datastore.PropertyList
// to also hold a *datastore.Key.
type Entity struct {
	Key        *datastore.Key
	Properties datastore.PropertyList
}

// Load decodes all properties into Entity.Propertioes object,
// implementing the datastore.PropertyLoadSaver interface.
func (e *Entity) Load(p []datastore.Property) error {
	return e.Properties.Load(p)
}

// Save encodes all properties from Entity.Properties object,
// implmenting the datastore.PropetyLoadSaver interface.
func (e *Entity) Save() ([]datastore.Property, error) {
	return e.Properties.Save()
}

// Add append p to the Properties attribute.
func (e *Entity) Add(p datastore.Property) {
	e.Properties = append(e.Properties, p)
}

// Get returns the property value of the given name.
// If no property with that name exists, returns nil. It also
// returns nil if the property Value attribute is nil.
func (e *Entity) Get(name string) interface{} {
	for _, p := range e.Properties {
		if p.Name == name {
			return p.Value
		}
	}
	return nil
}

// GetInt returns the int value of the named property,
// and returns the zero value (0) if the property is not found,
// if its value is nil or if its type is not int, int32 or int64.
func (e *Entity) GetInt(name string) int64 {
	v := e.Get(name)
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

// GetFloat returns the string value of the named property,
// and returns the zero value (0.0) if the property is not found,
// if its value is nil or if its type is not float32 or float64.
func (e *Entity) GetFloat(name string) float64 {
	v := e.Get(name)
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

// GetString returns the string value of the named property,
// and returns the zero value ("") if the property is not found,
// if its value is nil or if its type is not string.
func (e *Entity) GetString(name string) string {
	v := e.Get(name)
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

// GetBool returns the string value of the named property,
// and returns the zero value (false) if the property is not found,
// if its value is nil, or if its type is not bool.
func (e *Entity) GetBool(name string) bool {
	v := e.Get(name)
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
