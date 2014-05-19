// Package bigquerysync allow the AppEngine Datastore to be synced
// with Google BigQuery.
package bigquerysync

import (
	"log"
	"net/http"

	"appengine"

	"ronoaldo.gopkg.net/aetools"

	"appengine/datastore"
)

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
// appengine context.
func SyncKeyRange(c appengine.Context, kstart, kend *datastore.Key) {
	q := datastore.NewQuery(kstart.Kind())

	if kstart.Equal(kend) {
		q = q.Filter("key =", kstart)
	} else {
		q = q.Filter("key >=", kstart)
		if kend != nil {
			q = q.Filter("<", kend)
		}
	}
	q = q.Order("key").Limit(10000)

	klast := ""
	done := false
	for it := q.Run(c); ; {
		e := new(aetools.Entity)

		key, err := it.Next(e)
		if err == datastore.Done {
			done = true
			break
		}
		if err != nil {
			c.Warningf("Error loading next entity: %s", err.Error())
			continue
		}

		e.Key = key
		klast = key.Encode()

		err = IngestToBigQuery(c, e)
		if err != nil {
			c.Warningf("Error loading ingesting %s into BigQuery: %s", e.Key, err.Error())
		}
	}

	if !done {
		if kend == nil {
			ScheduleSync(klast, "")
		} else {
			ScheduleSync(klast, kend.Encode())
		}
	}
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

// IngestToBigQuery takes an aetools.Entity, and ingest it's JSON representation
// into the configured BigQuery table, via streaming.
func IngestToBigQuery(c appengine.Context, e *aetools.Entity) error {
	j, _ := e.MarshalJSON()
	log.Printf("Ingest %s", j)
	return nil
}

// ScheduleSync is a function that schedules a new iteration of SynSyncEntityHandler,
// using the configured task queue.
var ScheduleSync = func(kstart, kend string) {
	log.Printf("Reschedule from %s to %s", kstart, kend)
}
