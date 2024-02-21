package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func init() {
	// XXX: find better pushd/popd pattern, eg https://stackoverflow.com/a/28200371/244128
	os.Chdir("..")
}

func TestStatusRoute(t *testing.T) {
	config := Config{Mode: "test"}
	server, err := NewServer(config, nil)
	tmust(t, err)

	res := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/_status", nil)
	tmust(t, err)
	server.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)

	var body StatusDTO
	tmust(t, json.Unmarshal(res.Body.Bytes(), &body))
	assert.Equal(t, "ok", body.Status)
	assert.Equal(t, "test", body.Mode)
	assert.Equal(t, "GET", body.HTTP.Method)
	assert.Equal(t, "/_status", body.HTTP.URL)
}

func Test404(t *testing.T) {
	server, err := NewServer(Config{}, nil)
	tmust(t, err)

	res := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/notfound", nil)
	tmust(t, err)
	server.ServeHTTP(res, req)

	assert.Equal(t, 404, res.Code)
	body := res.Body.String()
	assert.True(t, must1(regexp.MatchString(`\b404\b`, body)), `matches 404`)
	assert.True(t, must1(regexp.MatchString(`\bNot found\b`, body)), `matches "Not found"`)
	assert.Equal(t, "text/html; charset=utf-8", res.Header().Get("Content-Type"))
}

func Test404PreferJson(t *testing.T) {
	server, err := NewServer(Config{}, nil)
	tmust(t, err)

	res := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/notfound", nil)
	tmust(t, err)
	req.Header.Add("Accept", "application/json,*")
	server.ServeHTTP(res, req)

	assert.Equal(t, 404, res.Code)
	assert.Equal(t, "text/json; charset=utf-8", res.Header().Get("Content-Type"))
	assert.Equal(t, `{"error":"not found"}`, res.Body.String())
}

func Test500(t *testing.T) {
	server, err := NewServer(Config{}, nil)
	tmust(t, err)

	res := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/panic", nil)
	tmust(t, err)
	server.ServeHTTP(res, req)

	assert.Equal(t, 500, res.Code)
	assert.Equal(t, "text/html; charset=utf-8", res.Header().Get("Content-Type"))
	body := res.Body.String()
	assert.True(t, must1(regexp.MatchString(`\b500\b`, body)), `matches 500`)
	assert.True(t, must1(regexp.MatchString(`\bInternal server error\b`, body)), `matches "Internal server error"`)
}

func Test500PreferJson(t *testing.T) {
	server, err := NewServer(Config{}, nil)
	tmust(t, err)

	res := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/panic", nil)
	tmust(t, err)
	req.Header.Add("Accept", "application/json,*")
	server.ServeHTTP(res, req)

	assert.Equal(t, 500, res.Code)
	assert.Equal(t, "text/json; charset=utf-8", res.Header().Get("Content-Type"))
	assert.Equal(t, `{"error":"internal server error"}`, res.Body.String())
}

func tmust(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}
