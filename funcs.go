package aetools

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"

	"appengine"
	"appengine/datastore"
)

const (
	// DateTimeFormat is used to store and load time.Time objects
	DateTimeFormat = time.RFC3339
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
	// LoadSync is an aetools.Options that enforces data to sync
	// after it get loaded into the datastore.
	// This Options will cause a datastore.Get to happen
	// for each entity loaded.
	LoadSync = &Options{
		GetAfterPut: true,
	}
)

// Options allows callees to specify parameters to the Load function.
type Options struct {
	// GetAfterPut indicates if we must force the Datastore to load
	// entities to be visible for non-ancestor queries, by issuing a
	// Get by key.
	// Not used when loading.
	GetAfterPut bool

	// The size for batch operations when loading/dumping
	BatchSize int

	// Kind is used to specify the kind when dumping.
	// Not used when loading.
	Kind string

	// PrettyPrint is used to specify if the dump should beaultify the output.
	// Not used when loading.
	PrettyPrint bool
}

// DumpOptions is deprecated. Use Options instead.
type DumpOptions struct {
	Options
}

// LoadJSON is a convenient wrapper to call Load using a JSON string in memory,
// wrapped by a strings.Reader. The error result from Load, if any, is returned.
func LoadJSON(c appengine.Context, s string, o *Options) error {
	return Load(c, strings.NewReader(s), o)
}

// Load reads the JSON representation of entities from the io.Reader "r",
// and stores them in the Datastore using the given appengine.Context.
// The Options parameter allows you to configure how the dump will work.
// If there is any parsing erros, improper format, or datastore failures
// during the process, that error is returned and processing stops. The
// error may be returned after some entities were loaded: there is no
// parsing cache.
func Load(c appengine.Context, r io.Reader, o *Options) error {
	entities, err := DecodeEntities(c, r)
	if err != nil {
		return err
	}
	if len(entities) == 0 {
		c.Infof("Skipping load of 0 entities")
		return nil
	}
	batchSize := o.BatchSize
	if batchSize <= 0 {
		batchSize = 50
	}
	for start, end := 0, 0; start < len(entities); {
		end += batchSize
		if end > len(entities) {
			end = len(entities)
		}
		keys := make([]*datastore.Key, 0, end-start)
		values := make([]datastore.PropertyList, 0, cap(keys))

		for _, e := range entities[start:end] {
			keys = append(keys, e.Key)
			values = append(values, e.Properties)
		}

		keys, err = datastore.PutMulti(c, keys, values)
		if err != nil {
			return err
		}
		c.Infof("Loaded %d entities ...", len(keys))

		if o.GetAfterPut {
			c.Infof("Making a read to force consistency ...")
			l := make([]Entity, len(keys))
			err := datastore.GetMulti(c, keys, l)
			if err != nil {
				return err
			}
		}

		start = end
	}

	return nil
}

// DumpJSON is a convenient wrapper that captures the generated JSON from Dump
// in memory, and return it as a string. If Dump returns an error, an empty
// string and the error are returned.
func DumpJSON(c appengine.Context, o *Options) (string, error) {
	var w bytes.Buffer
	err := Dump(c, &w, o)
	if err != nil {
		return "", err
	}
	return w.String(), nil
}

// Dump exports entities from the context c using the specified Options o and
// writing the generated JSON representations to the io.Writer w. You can configure
// how the dump will run by using the Options parameter. If there is an error
// generating the output, or writting to the writer, it is returned. This method
// may return an error after writting bytes to w: the output is not buffered.
func Dump(c appengine.Context, w io.Writer, o *Options) error {
	var (
		comma        = []byte(",")
		openBracket  = []byte("[")
		closeBracket = []byte("]")
		lineFeed     = []byte("\n")
		indent       = "  "
	)

	w.Write(openBracket)
	count := 0
	last := 0
	batchSize := o.BatchSize
	if batchSize <= 0 {
		batchSize = 100
	}

	q := datastore.NewQuery(o.Kind).Limit(batchSize)
	for i := q.Run(c); ; {
		var e Entity
		k, err := i.Next(&e)
		e.Key = k
		if err == datastore.Done {
			c.Infof("datastore.Done: last=%d, count=%d", last, count)
			if last == count || count-last < batchSize {
				break
			}
			// This 100 batch is done, but more can be found in the next one
			last = count
			cur, err := i.Cursor()
			if err != nil {
				return err
			}
			c.Infof("restarting the query: cursor=%v", cur)
			i = datastore.NewQuery(o.Kind).Limit(batchSize).Start(cur).Run(c)
			continue
		}
		if err != nil {
			return err
		}
		if count > 0 {
			w.Write(comma)
			w.Write(lineFeed)
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
	w.Write(closeBracket)
	return nil
}

// EncodeEntities serializes the parameter into a JSON string.
func EncodeEntities(entities []Entity, w io.Writer) error {
	for i, e := range entities {
		err := encodeEntity(e, w)
		if err != nil {
			return fmt.Errorf("aetools: Unable to encode position %d: %s", i, err.Error())
		}
	}
	return nil
}

// DecodeEntities deserielizes the parameter from a JSON string
func DecodeEntities(c appengine.Context, r io.Reader) ([]Entity, error) {
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

// parseJSONArray parses a JSON array and returns it's value.
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

// encodeEntity serializes the given Entity into the provided writer.
func encodeEntity(e Entity, w io.Writer) error {
	b, err := e.MarshalJSON()
	if err != nil {
		return err
	}
	_, err = w.Write(b)
	return err
}

// decodeEntity decodes the map as an Entity struct.
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

// invalidIDError create an error for an invalid ID type.
func invalidIDError(id interface{}) error {
	return fmt.Errorf("aetest: invalid key id/name '%v' (type %T)", id, reflect.TypeOf(id))
}
