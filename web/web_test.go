package web

import (
	"testing"
)

func assertEqual(t *testing.T, exampleIndex int, expected, actual any) {
	if actual != expected {
		t.Errorf("example %#v, wanted %#v, got %#v",
			exampleIndex+1, expected, actual)
	}
}
