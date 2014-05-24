package aetools

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"

	"appengine"
	"appengine/datastore"
)

const (
	// DateTimeFormat is used to store and load time.Time objects
	DateTimeFormat = "2006-01-02 15:04:05.000 -0700"
)

var (
	// ErrInvalidRootElement is returned when the root element is not a valid JSON Array.
	ErrInvalidRootElement = errors.New("aetools: root object is not an array")
	// ErrInvalidElementType is retunred when the element is not a JSON Object.
	ErrInvalidElementType = errors.New("aetools: element is not a JSON object")
	// ErrInvalidPropertiesElement is returned when the field to be decoded is not valid.
	ErrInvalidPropertiesElement = errors.New("aetools: element's properties field is invalid")
	// ErrNoKeyElement is returned for an entity with missing key information.
	ErrNoKeyElement = errors.New("aetools: element's key field is not present")
	// ErrInvalidKeyElement is returned when the key is not properly encoded.
	ErrInvalidKeyElement = errors.New("aetools: element's key field is invalid")
)

var (
	LoadSync = &Options{
		GetAfterPut: true,
	}
)

type Options struct {
	// GetAfterPut indicates if we must force the Datastore to load
	// entities to be visible for non-ancestor queries, by issuing a
	// Get by key.
	GetAfterPut bool
}

// DumpOptions are options to configure how the entities are dumped.
type DumpOptions struct {
	Kind        string // The entity kind. Defaults to all entities ("").
	PrettyPrint bool   // If we should pretty print each line, defaults to false.
}

// LoadJSON loads entities from the JSON encoded string.
func LoadJSON(c appengine.Context, j string, o *Options) error {
	return LoadFixtures(c, strings.NewReader(j), o)
}

// LoadFixtures load the Json representation of entities from
// the io.Reader into the Datastore, using the given appengine.Context.
func LoadFixtures(c appengine.Context, r io.Reader, o *Options) error {
	entities, err := decodeEntities(c, r)
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

	if o.GetAfterPut {
		l := make([]Entity, len(keys))
		err := datastore.GetMulti(c, keys, l)
		if err != nil {
			return err
		}
	}

	return nil
}

func DumpJSON(c appengine.Context, o *DumpOptions) (string, error) {
	var w bytes.Buffer
	err := DumpFixtures(c, &w, o)
	if err != nil {
		return "", err
	}
	return w.String(), err
}

// DumpFixtures exports all entities, as JSON, writing the results in the
// specified writer. You can configure how the dump will run by using the
// DumpOptions parameter.
func DumpFixtures(c appengine.Context, w io.Writer, o *DumpOptions) error {
	var (
		comma  = []byte(",")
		op_b   = []byte("[")
		cl_b   = []byte("]")
		lf_b   = []byte("\n")
		indent = "  "
	)

	w.Write(op_b)
	count := 0

	q := datastore.NewQuery(o.Kind)
	for i := q.Run(c); ; {
		var e Entity
		k, err := i.Next(&e)
		e.Key = k
		if err == datastore.Done {
			break
		}
		if err != nil {
			return err
		}
		if count > 0 {
			w.Write(comma)
			w.Write(lf_b)
		}
		var b []byte
		if o.PrettyPrint {
			b, err = json.MarshalIndent(&e, "", indent)
		} else {
			b, err = json.Marshal(&e)
		}
		if err != nil {
			return err
		}
		w.Write(b)
		count++
	}
	w.Write(cl_b)
	return nil
}

// encodeEntities serializes the parameter into a JSON string.
func encodeEntities(entities []Entity, w io.Writer) error {
	for i, e := range entities {
		err := encodeEntity(e, w)
		if err != nil {
			return fmt.Errorf("aetools: Unable to encode position %d: %s", i, err.Error())
		}
	}
	return nil
}

func decodeEntities(c appengine.Context, r io.Reader) ([]Entity, error) {
	a, err := parseJSONArray(r)
	if err != nil {
		return nil, err
	}

	var result []Entity

	for _, i := range a {
		m, ok := i.(map[string]interface{})
		if !ok {
			return nil, ErrInvalidElementType
		}

		e, err := decodeEntity(c, m)
		if err != nil {
			return nil, err
		}

		result = append(result, *e)
	}

	return result, nil
}

func parseJSONArray(r io.Reader) ([]interface{}, error) {
	d := json.NewDecoder(r)
	d.UseNumber()

	//Generic decode into an empty interface
	var i interface{}
	err := d.Decode(&i)
	if err != nil {
		return nil, err
	}

	//Chek casting to array of interfaces, so we make sure the Json
	//is a list of entities.
	a, ok := i.([]interface{})
	if !ok {
		return nil, ErrInvalidRootElement
	}
	return a, nil
}

func encodeEntity(e Entity, w io.Writer) error {
	b, err := e.MarshalJSON()
	if err != nil {
		return err
	}
	_, err = w.Write(b)
	return err
}

func decodeEntity(c appengine.Context, m map[string]interface{}) (*Entity, error) {
	var e Entity
	var err error

	for k, v := range m {
		if k == "__key__" {
			e.Key, err = decodeKey(c, v)
			if err != nil {
				return nil, err
			}
		} else {
			switch v.(type) {
			case []interface{}:
				l := v.([]interface{})
				for _, v := range l {
					err = decodeProperty(c, k, v, &e)
					if err != nil {
						return nil, err
					}
					e.Properties[len(e.Properties)-1].Multiple = true
				}
			default:
				err = decodeProperty(c, k, v, &e)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	return &e, nil
}

func invalidIDError(id interface{}) error {
	return fmt.Errorf("aetest: invalid key id/name '%v' (type %T)", id, reflect.TypeOf(id))
}
