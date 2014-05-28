package bigquerysync_test

import (
	"testing"

	"ronoaldo.gopkg.net/aetools/bigquerysync"

	"appengine/datastore"
)

func TestSyncKeyRangeWithOpenEnd(t *testing.T) {
	c := SetupEnv(t)
	defer c.Close()

	var start, end *datastore.Key
	start = datastore.NewKey(c, "Sample", "", 1, nil)

	ingested, last, err := bigquerysync.SyncKeyRange(c, "project", "dataset", start, end)
	if err != nil {
		t.Errorf("Unexpected failure: %s", err.Error())
	}
	if ingested != 3 {
		t.Errorf("Unexpected ammount of ingested entities: %d, expected: %d", ingested, 3)
	}
	if !last.Equal(end) {
		t.Errorf("Unexpected last %s (expected %s)", last, end)
	}
}

func TestSyncExplicitKeyRange(t *testing.T) {
	c := SetupEnv(t)
	defer c.Close()

	start := datastore.NewKey(c, "Sample", "", 1, nil)
	end := datastore.NewKey(c, "Sample", "", 3, nil)

	ingested, last, err := bigquerysync.SyncKeyRange(c, "project", "dataset", start, end)
	if err != nil {
		t.Errorf("Unexpected failure: %s", err.Error())
	}
	if ingested != 2 {
		t.Errorf("Unexpected ammount of ingested entities: %d, expected %d", ingested, 2)
	}
	if !last.Equal(end) {
		t.Errorf("Unexpected last %s (expected %s)", last, end)
	}
}

func TestSyncSingleEntity(t *testing.T) {
	c := SetupEnv(t)
	defer c.Close()

	start := datastore.NewKey(c, "Sample", "", 2, nil)
	end := datastore.NewKey(c, "Sample", "", 2, nil)

	ingested, last, err := bigquerysync.SyncKeyRange(c, "project", "dataset", start, end)
	if err != nil {
		t.Errorf("Unexpected failure: %s", err.Error())
	}
	if ingested != 1 {
		t.Errorf("More than one entities ingested: %d", ingested)
	}
	if !last.Equal(end) {
		t.Errorf("Unexpected last %s (expected %s)", last, end)
	}
}

func TestSyncInvalidKeyIntervals(t *testing.T) {
	c := SetupEnv(t)
	defer c.Close()

	var start, end *datastore.Key
	ingested, last, err := bigquerysync.SyncKeyRange(c, "project", "dataset", start, end)
	if err == nil {
		t.Errorf("Missing error for nil start key")
	}
	if ingested != 0 {
		t.Errorf("Ingested %d entities for nil start and end keys", ingested)
	}
	if last != nil {
		t.Errorf("Unexpected last %s (expected nil)", last)
	}
}
