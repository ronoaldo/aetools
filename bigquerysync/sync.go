package bigquerysync

import (
	"bytes"
	"fmt"
	"log"
	"sort"
	"strings"

	"appengine"
	"appengine/datastore"

	"ronoaldo.gopkg.net/aetools"
)

const (
	MaxErrorsPerSync = 10
	BatchSize        = 81
)

var (
	ScatterProperty = "__scatter__"
)

// Errors collects all errors during the igestion job
// for reporting
type Errors []error

func (e Errors) Error() string {
	b := new(bytes.Buffer)
	l := len(e)
	for i, err := range e {
		b.WriteString(err.Error())
		if i < l {
			b.WriteRune(';')
		}
	}
	return b.String()
}

// SyncKeyRange sinchronizes the specified key range using the provided
// appengine context. The kind for start and end keys must be the same.
// The method returns error if start key is nil. If end key is nil, all
// entities starting from starkey are processed.
//
// The calee is responsible for checking if the last key returned is equal
// to the end parameter, and eventually reschedule the syncronization:
//
//	start, end = startKey(), endKey()
//	count, start, err := SyncKeyRange(c, proj, dataset, start, end)
//	if err != nil {
//		// Handle errors
//	} else if !start.Equal(end) {
//		// Reschedule from new start
//	}
//
// The above sample code ilustrates how to handle the results.
func SyncKeyRange(c appengine.Context, project, dataset string, start, end *datastore.Key) (int, *datastore.Key, error) {
	if start == nil {
		return 0, nil, fmt.Errorf("bigquerysync: invalid nil start")
	}
	var (
		errors   = make(Errors, 0)
		ingested = 0
		done     = false
		cur      datastore.Cursor
		last     *datastore.Key
		buff     = make([]*aetools.Entity, 0, BatchSize)
	)
	q := createQuery(start, end, cur)
	for it := q.Run(c); ; {
		e := new(aetools.Entity)

		// TODO(ronoaldo): make this for loop consume entities
		// from a goroutine channel, so it is easier to retry and buffer
		// the goroutine instead of this complex logic.
		key, err := it.Next(e)
		if err == datastore.Done {
			done = true
			break
		}
		if err != nil {
			if strings.Contains(err.Error(), "datastore operation timed out") {
				c.Infof("Continuing from cursor '%s', due to error %s", cur, err.Error())
				q := createQuery(start, end, cur)
				it = q.Run(c)
				continue
			}
			errors = append(errors, err)
			c.Warningf("Error loading next entity: %s", err.Error())
			break
		}

		e.Key, last = key, key
		buff = append(buff, e)

		cur, err = it.Cursor()
		if err != nil {
			errors = append(errors, err)
			c.Warningf("Error fething cursor: %s", err.Error())
		}

		if err != nil {
			errors = append(errors, err)
			c.Warningf("Error loading ingesting %s into BigQuery: %s", e.Key, err.Error())
		} else {
			ingested++
		}

		if len(errors) > MaxErrorsPerSync {
			done = true
			break
		}
		if len(buff) == BatchSize {
			break
		}
	}
	// Due to appengine/urlfetch payload limits, we split the BatchSize in 9-entity
	// batches
	for i, j := 0, 9; i < len(buff)+9; i, j = i+9, j+9 {
		if j > len(buff) {
			j = len(buff)
		}
		if i >= j {
			break
		}
		c.Infof("Ingesting %d entities into %s:%s [%d:%d]", len(buff), project, dataset, i, j)
		err := IngestToBigQuery(c, project, dataset, buff[i:j])
		if err != nil {
			errors = append(errors, err)
		}
	}

	if done {
		last = end
	}

	if len(errors) > 0 {
		return ingested, last, errors
	}

	return ingested, last, nil
}

type KeyRange struct {
	Start *datastore.Key
	End   *datastore.Key
}

// cmpStr compares two strings returning -1, 0, or 1 if
// s is less than, equal or grather than other.
func cmpStr(s, other string) int {
	if s < other {
		return -1
	} else if s > other {
		return 1
	}
	return 0
}

// cmpInt compares two int64 returning -1, 0, or 1 if i is
// less than, equal or grather than other.
func cmpInt(i, other int64) int {
	if i < other {
		return -1
	} else if i > other {
		return 1
	}
	return 0
}

// cmpKey compares k and other, returning -1, 0 or 1 if k is
// less than, equal or grather than other. The algorithm doesn't
// takes into account any ancestors in the two keys. The order
// of comparision is AppID, Kind, IntID and StringID. Keys with
// integer identifiers are smaller than string identifiers.
func cmpKey(k, other *datastore.Key) int {
	if k == other {
		return 0
	}
	if r := cmpStr(k.AppID(), other.AppID()); r != 0 {
		return r
	}
	if r := cmpStr(k.Kind(), other.Kind()); r != 0 {
		return r
	}
	if k.IntID() != 0 {
		if other.IntID() == 0 {
			return -1
		}
		return cmpInt(k.IntID(), other.IntID())
	}
	if other.IntID() != 0 {
		return 1
	}
	return cmpStr(k.StringID(), other.StringID())
}

// KeyPath takes a datastore.Key and decomposes its ancestor path
// as a slice of keys, where the first ancestor is at position 0.
func KeyPath(k *datastore.Key) []*datastore.Key {
	path := make([]*datastore.Key, 0)
	for p := k; p != nil; p = p.Parent() {
		path = append(path, nil)
		copy(path[1:], path[0:])
		path[0] = p
	}
	return path
}

// CompareKeys compares k and other, returning -1, 0, 1 if k is less than
// equal or grather than other, taking into account the full ancestor path.
func CompareKeys(k, other *datastore.Key) int {
	if k == other {
		return 0
	}

	thisPath := KeyPath(k)
	otherPath := KeyPath(other)

	for i, thisKey := range thisPath {
		if i < len(otherPath) {
			otherKey := otherPath[i]
			result := cmpKey(thisKey, otherKey)
			if result != 0 {
				return result
			}
		} else {
			return 1
		}
	}

	if len(otherPath) > len(thisPath) {
		return -1
	}
	return 0
}

// byKey implements sort.Interface to sort keys by they path.
type byKey []*datastore.Key

func (b byKey) Len() int           { return len(b) }
func (b byKey) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b byKey) Less(i, j int) bool { return CompareKeys(b[i], b[j]) == -1 }

// KeyRangesForKind generates a set of KeyRanges, attempting to make them uniformly
// distributed by using the __scatter__ property implementation.
func KeyRangesForKind(c appengine.Context, kind string) []KeyRange {
	// TODO(ronoaldo): compute rangeLen using datastore statistics
	rangeLen := 64
	// Start key is the first entity key
	sq := datastore.NewQuery(kind).Order("__key__").KeysOnly().Limit(1)
	it := sq.Run(c)
	start, err := it.Next(nil)
	if err != nil || start == nil {
		// No entities found, return empty range
		return []KeyRange{}
	}
	c.Infof("Found start key %s", start)
	// Find scatters to build ranges
	q := datastore.NewQuery(kind).Order(ScatterProperty).KeysOnly().Limit(rangeLen)
	keys := make([]*datastore.Key, 0, rangeLen)
	for it := q.Run(c); ; {
		k, err := it.Next(nil)
		if err == datastore.Done {
			break
		}
		if err != nil {
			c.Infof("Error iterating over scatters: %s", err.Error())
			break
		}
		keys = append(keys, k)
	}
	// No scatters, single range
	if len(keys) < 1 {
		return []KeyRange{KeyRange{start, nil}}
	}
	// Sort by keys and build the key ranges, leaving the final range open
	sort.Sort(byKey(keys))
	ranges := make([]KeyRange, 0, len(keys))
	for _, k := range keys {
		r := KeyRange{start, k}
		ranges = append(ranges, r)
		start = k
	}
	ranges = append(ranges, KeyRange{keys[len(keys)-1], nil})
	return ranges
}

// createQuery builds a range query using start and end. It works
// for [start,end[, [start,nil] and [start,start] intervals. The
// returned query is sorted by __key__ and limited to BatchSize.
func createQuery(start, end *datastore.Key, cur datastore.Cursor) *datastore.Query {
	q := datastore.NewQuery(start.Kind())

	if start.Equal(end) {
		q = q.Filter("__key__ =", start)
	} else {
		q = q.Filter("__key__ >=", start)
		if end != nil {
			q = q.Filter("__key__ <", end)
		}
	}

	if cur.String() != "" {
		q = q.Start(cur)
	}

	q = q.Order("__key__")
	return q
}

// decodeKey safely decodes the specified key string, returning
// nil if there is a decoding error.
func decodeKey(k string) *datastore.Key {
	if k == "" {
		return nil
	}
	key, err := datastore.DecodeKey(k)
	if err != nil {
		// TODO(ronoaldo): log to gae console
		log.Printf("Unable to decode key %s: %s", k, err.Error())
		return nil
	}
	return key
}
