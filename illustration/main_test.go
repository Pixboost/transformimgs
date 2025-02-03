package main

import "testing"

func TestPercentage(t *testing.T) {

	b := 1234
	a := uint(float32(b) * 0.1)
	if uint(b/10) != a {
		t.Errorf("%d != %d", 1234/10, uint(float32(b)*0.1))
	}

	if a != uint(123) {
		t.Errorf("%d != %d", a, uint(123))
	}
}
