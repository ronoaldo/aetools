// Copyright 2014 Ronoaldo JLP <ronoaldo@gmail.com>
// Licensed under the Apache License, Version 2.0

package bigquerysync_test

import (
	"testing"

	"github.com/ronoaldo/aetools"
	"github.com/ronoaldo/aetools/bigquerysync"

	"google.golang.org/appengine/aetest"
	"google.golang.org/appengine/datastore"
)

func init() {
	bigquerysync.ScatterProperty = "_scatter__"
}

func TestSyncKeyRangeWithOpenEnd(t *testing.T) {
	c := SetupEnv(t)
	defer c.Close()

	var start, end *datastore.Key
	start = datastore.NewKey(c, "Sample", "", 1, nil)

	ingested, last, err := bigquerysync.SyncKeyRange(c, "project", "dataset", start, end, "")
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

	ingested, last, err := bigquerysync.SyncKeyRange(c, "project", "dataset", start, end, "")
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

	ingested, last, err := bigquerysync.SyncKeyRange(c, "project", "dataset", start, end, "")
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
	ingested, last, err := bigquerysync.SyncKeyRange(c, "project", "dataset", start, end, "")
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

func TestKeyRangeForKind(t *testing.T) {
	c, clean, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer clean()
	// No entities: empty range
	ranges := bigquerysync.KeyRangesForKind(c, "RangeTest")
	if len(ranges) != 0 {
		t.Errorf("Unexpected ranges returned: %d, expected 0: %#v", len(ranges), ranges)
	}
	// Setup datastore - __scatter__ is replaced with _scatter__ for testing
	aetools.LoadJSON(c, SampleEntities, aetools.LoadSync)
	// No scatter, single range
	ranges = bigquerysync.KeyRangesForKind(c, "Sample")
	if len(ranges) != 1 {
		t.Errorf("Unexpected ranges without scatters: %v, expected length 1", ranges)
	} else {
		if ranges[0].Start == nil {
			t.Errorf("Unexpected nil start for Sample")
		} else {
			if ranges[0].Start.IntID() != 1 || ranges[0].Start.Kind() != "Sample" {
				t.Errorf("Unexpected start key for Sample: %#v", ranges[0].Start)
			}
			if ranges[0].End != nil {
				t.Errorf("Unexpected end key for Sample: %#v, expected nil", ranges[0].End)
			}
		}
	}
	// Scattered entities: sorted key ranges expected
	ranges = bigquerysync.KeyRangesForKind(c, "RangeTest")
	expected := []struct {
		Start int64
		End   int64
	}{
		{1, 30},
		{30, 50},
		{50, 1000},
		{1000, 0},
	}
	if len(ranges) != len(expected) {
		t.Errorf("Unexpected ranges with scatter: %d, expected 3", len(ranges))
	} else {
		for i, e := range expected {
			r := ranges[i]
			if r.Start == nil {
				t.Errorf("Unexpected nil start at range %d", i)
			} else if r.Start.IntID() != e.Start {
				t.Errorf("Unexpected start at range %d: %#v, expected %d", i, r.Start.IntID(), e.Start)
			}
			if e.End == 0 {
				if r.End != nil {
					t.Errorf("Unexpected end at range %d: %#v, expected nil", i, r.End)
				}
			} else if r.End == nil {
				t.Errorf("Unexpected nil end at range %d: %#v, expected %v", i, r.End, e.End)
			} else if r.End.IntID() != e.End {
				t.Errorf("Unexpected end at range %d: %#v, expected %d", i, r.End.IntID(), e.End)
			}
		}
		// Check if all entity keys match
		for i, r := range ranges {
			if r.Start != nil {
				if r.Start.Kind() != "RangeTest" {
					t.Errorf("Unexpected kind at range %d: %s", i, r.Start.Kind())
				}
			}
			if r.End != nil {
				if r.End.Kind() != "RangeTest" {
					t.Errorf("Unexpected kind at range %d: %s", i, r.End.Kind())
				}
			}
		}
	}
}

func TestCompareKeys(t *testing.T) {
	c, clean, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer clean()

	A1 := datastore.NewKey(c, "A", "", 1, nil)
	A2 := datastore.NewKey(c, "A", "", 2, nil)
	Aa := datastore.NewKey(c, "A", "a", 0, nil)

	B1 := datastore.NewKey(c, "B", "", 1, A1)
	B2 := datastore.NewKey(c, "B", "", 2, A1)
	Ba := datastore.NewKey(c, "B", "a", 0, A1)

	compare := []struct {
		a *datastore.Key
		b *datastore.Key
		r int
	}{
		{A1, A1, 0},
		{A1, A2, -1},
		{A2, A1, 1},

		{A1, Aa, -1},
		{Aa, A1, 1},
		{Aa, Aa, 0},

		{A1, B1, -1},
		{B1, A1, 1},
		{B1, B1, 0},

		{B1, B2, -1},
		{B2, B1, 1},
		{B2, B2, 0},
		{Ba, Ba, 0},
	}

	for _, exp := range compare {
		r := bigquerysync.CompareKeys(exp.a, exp.b)
		if r != exp.r {
			t.Errorf("Error in CompareKey(%v, %v) = %d, expected %d", exp.a, exp.b, r, exp.r)
		}
	}
}
