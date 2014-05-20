// Package bigquerysync allow the AppEngine Datastore to be synced
// with Google BigQuery.
package bigquerysync

import (
	"bytes"
	"fmt"
	"log"
	"net/http"

	"appengine"
	"appengine/datastore"

	"ronoaldo.gopkg.net/aetools"
)

const (
	MaxErrorsPerSync = 10
	BatchSize        = 1000
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

// SyncEntityHandler syncrhonizes the entities starting from startKey,
// util endKey, exclusive. If no endKey is specified, the open end results
// in all entities from startKey beign synced. If startKey and endKey are
// equal, only a single entity is processed, if found.
func SyncEntityHandler(w http.ResponseWriter, r *http.Request) {
	var (
		kstart = decodeKey(r.FormValue("startKey"))
		kend   = decodeKey(r.FormValue("endKey"))
	)

	if kstart == nil {
		http.Error(w, "Start key canno't be nil.", http.StatusBadRequest)
		return
	}

	c := appengine.NewContext(r)
	SyncKeyRange(c, kstart, kend)
}

// SyncKeyRange sinchronizes the specified key range using the provided
// appengine context. The kinds for both kstart and kend must be the same.
func SyncKeyRange(c appengine.Context, kstart, kend *datastore.Key) (int, error) {
	var (
		q, err   = createQuery(kstart, kend)
		errors   = make(Errors, 0)
		ingested = 0
		done     = false

		klast *datastore.Key
	)

	if err != nil {
		return 0, err
	}

	for it := q.Run(c); ; {
		e := new(aetools.Entity)

		key, err := it.Next(e)
		if err == datastore.Done {
			done = true
			break
		}
		if err != nil {
			errors = append(errors, err)
			c.Warningf("Error loading next entity: %s", err.Error())
			break
		}

		e.Key, klast = key, key

		err = IngestToBigQuery(c, e)
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
	}

	if !done {
		ScheduleSync(c, klast, kend)
	}

	if len(errors) > 0 {
		return ingested, errors
	}

	return ingested, nil
}

// createQuery builds a range query using start and end. It works
// for [start,end[, [start,nil] and [start,start] intervals. The
// returned query is sorted by __key__ and limited to BatchSize.
func createQuery(kstart, kend *datastore.Key) (*datastore.Query, error) {
	if kstart == nil {
		return nil, fmt.Errorf("Invalid nil kstart")
	}
	q := datastore.NewQuery(kstart.Kind())

	if kstart.Equal(kend) {
		q = q.Filter("__key__ =", kstart)
	} else {
		q = q.Filter("__key__ >=", kstart)
		if kend != nil {
			q = q.Filter("__key__ <", kend)
		}
	}

	q = q.Order("__key__").Limit(BatchSize)
	return q, nil
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

// ScheduleSync is a function that schedules a new iteration of SyncEntityHandler,
// using the configured task queue.
var ScheduleSync = func(c appengine.Context, kstart, kend *datastore.Key) {
	c.Debugf("Reschedule from %s to %s", kstart, kend)
}
