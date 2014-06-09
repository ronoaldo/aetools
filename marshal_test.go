package aetools

import (
	"testing"
)

func TestMarshalFloat(t *testing.T) {
	cases := []struct {
		Value    float64
		Expected string
	}{
		{1, "1.0"},
		{1.01, "1.01"},
		{1e7, "1e+07"},
	}
	for i, c := range cases {
		b, err := float(c.Value).MarshalJSON()
		if err != nil {
			t.Fatal(err)
		}
		if string(b) != c.Expected {
			t.Errorf("%d: Unexpected float value %s, expected: %s", i, string(b), c.Expected)
		}
	}
}
