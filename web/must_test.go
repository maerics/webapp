package web

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPreferJson(t *testing.T) {
	type h http.Header

	for _, eg := range []struct {
		acceptHeader h
		expected     bool
	}{
		{h{}, false},
		{h{"Accept": {"application/json"}}, true},
		{h{"Accept": {"text/json"}}, true},
		{h{"Accept": {"application/html"}}, false},
		{h{"Accept": {"text/html"}}, false},
		{h{"Accept": {"text/json,text/html"}}, true},
		{h{"Accept": {"text/html, text/json"}}, false},
		{h{"Accept": {"text/html"}, "Content-Type": {"text/json"}}, false},
		{h{"Accept": {"text/html, *"}, "Content-Type": {"text/json"}}, false},
		{h{"Accept": {"*"}, "Content-Type": {"text/json"}}, true},
		{h{"Content-Type": {"text/json"}}, true},
	} {
		assert.Equal(t, eg.expected, preferJson(http.Header(eg.acceptHeader)))
	}
}
