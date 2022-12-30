package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/assert/v2"
)

func TestStatusRoute(t *testing.T) {
	config := Config{Environment: "test"}
	server, err := NewServer(config, nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/_status", nil)
	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var body StatusDTO
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "ok", body.Status)
	assert.Equal(t, "test", body.Env)
	assert.Equal(t, "GET", body.HTTP.Method)
	assert.Equal(t, "/_status", body.HTTP.URL)
}

func assertEqual(t *testing.T, exampleIndex int, expected, actual any) {
	if actual != expected {
		t.Errorf("example %#v, wanted %#v, got %#v",
			exampleIndex+1, expected, actual)
	}
}
