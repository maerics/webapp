package web

import "testing"

func TestPreferJson(t *testing.T) {
	for i, eg := range []struct {
		acceptHeader string
		expected     bool
	}{
		{"", false},
		{"application/json", true},
		{"text/json", true},
		{"application/html", false},
		{"text/html", false},
		{"text/json,text/html", true},
		{"text/html, text/json", false},
	} {
		assertEqual(t, i, eg.expected, preferJson(eg.acceptHeader))
	}
}
