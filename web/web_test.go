package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
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
