package bigquerysync_test

import (
	"testing"

	"ronoaldo.gopkg.net/aetools/bigquerysync"

	"appengine/datastore"
)

func TestSyncKeyRangeWithOpenEnd(t *testing.T) {
	c := SetupEnv(t)
	defer c.Close()

	var kstart, kend *datastore.Key
	kstart = datastore.NewKey(c, "Sample", "", 1, nil)

	ingested, klast, err := bigquerysync.SyncKeyRange(c, kstart, kend)
	if err != nil {
		t.Errorf("Unexpected failure: %s", err.Error())
	}
	if ingested != 3 {
		t.Errorf("Unexpected ammount of ingested entities: %d, expected: %d", ingested, 3)
	}
	if !klast.Equal(kend) {
		t.Errorf("Unexpected klast %s (expected %s)", klast, kend)
	}
}

func TestSyncExplicitKeyRange(t *testing.T) {
	c := SetupEnv(t)
	defer c.Close()

	kstart := datastore.NewKey(c, "Sample", "", 1, nil)
	kend := datastore.NewKey(c, "Sample", "", 3, nil)

	ingested, klast, err := bigquerysync.SyncKeyRange(c, kstart, kend)
	if err != nil {
		t.Errorf("Unexpected failure: %s", err.Error())
	}
	if ingested != 2 {
		t.Errorf("Unexpected ammount of ingested entities: %d, expected %d", ingested, 2)
	}
	if !klast.Equal(kend) {
		t.Errorf("Unexpected klast %s (expected %s)", klast, kend)
	}
}

func TestSyncSingleEntity(t *testing.T) {
	c := SetupEnv(t)
	defer c.Close()

	kstart := datastore.NewKey(c, "Sample", "", 2, nil)
	kend := datastore.NewKey(c, "Sample", "", 2, nil)

	ingested, klast, err := bigquerysync.SyncKeyRange(c, kstart, kend)
	if err != nil {
		t.Errorf("Unexpected failure: %s", err.Error())
	}
	if ingested != 1 {
		t.Errorf("More than one entities ingested: %d", ingested)
	}
	if !klast.Equal(kend) {
		t.Errorf("Unexpected klast %s (expected %s)", klast, kend)
	}
}

func TestSyncInvalidKeyIntervals(t *testing.T) {
	c := SetupEnv(t)
	defer c.Close()

	var kstart, kend *datastore.Key
	ingested, klast, err := bigquerysync.SyncKeyRange(c, kstart, kend)
	if err == nil {
		t.Errorf("Missing error for nil start key")
	}
	if ingested != 0 {
		t.Errorf("Ingested %d entities for nil start and end keys", ingested)
	}
	if klast != nil {
		t.Errorf("Unexpected klast %s (expected nil)", klast)
	}
}
