package bigquerysync_test

import (
	"strings"

	"ronoaldo.gopkg.net/aetools"
	"ronoaldo.gopkg.net/aetools/bigquerysync"

	"testing"

	"appengine/aetest"
	"appengine/datastore"
)

func TestSyncKeyRangeWithOpenEnd(t *testing.T) {
	c := setup(t)
	defer c.Close()

	var kstart, kend *datastore.Key
	kstart = datastore.NewKey(c, "Sample", "", 1, nil)

	ingested, err := bigquerysync.SyncKeyRange(c, kstart, kend)
	if err != nil {
		t.Errorf("Unexpected failure: %s", err.Error())
	}
	if ingested != 3 {
		t.Errorf("Unexpected ammount of ingested entities: %d, expected: %d", ingested, 3)
	}
}

func TestSyncExplicitKeyRange(t *testing.T) {
	c := setup(t)
	defer c.Close()

	kstart := datastore.NewKey(c, "Sample", "", 1, nil)
	kend := datastore.NewKey(c, "Sample", "", 3, nil)

	ingested, err := bigquerysync.SyncKeyRange(c, kstart, kend)
	if err != nil {
		t.Errorf("Unexpected failure: %s", err.Error())
	}
	if ingested != 2 {
		t.Errorf("Unexpected ammount of ingested entities: %d, expected %d", ingested, 2)
	}
}

func TestSyncSingleEntity(t *testing.T) {
	c := setup(t)
	defer c.Close()

	kstart := datastore.NewKey(c, "Sample", "", 2, nil)
	kend := datastore.NewKey(c, "Sample", "", 2, nil)

	ingested, err := bigquerysync.SyncKeyRange(c, kstart, kend)
	if err != nil {
		t.Errorf("Unexpected failure: %s", err.Error())
	}
	if ingested != 1 {
		t.Errorf("More than one entities ingested: %d", ingested)
	}
}

func TestSyncInvalidKeyIntervals(t *testing.T) {
	c := setup(t)
	defer c.Close()

	var kstart, kend *datastore.Key
	ingested, err := bigquerysync.SyncKeyRange(c, kstart, kend)
	if err == nil {
		t.Errorf("Missing error for nil start key")
	}
	if ingested != 0 {
		t.Errorf("Ingested %d entities for nil start and end keys", ingested)
	}
}

func setup(t *testing.T) aetest.Context {
	c, err := aetest.NewContext(nil)
	if err != nil {
		t.Fatal(err)
	}

	err = aetools.LoadFixtures(c, strings.NewReader(SampleEntities), &aetools.Options{true})
	if err != nil {
		c.Close()
		t.Fatal(err)
	}

	return c
}
