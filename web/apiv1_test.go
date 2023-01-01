package web

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"webapp/db"
	"webapp/models"

	"github.com/maerics/goutil"
	"github.com/stretchr/testify/assert"
)

func InitDB(t *testing.T) *Server {
	testdb, err := db.Connect(goutil.MustEnv(db.Env_TEST_DATABASE_URL))
	tmust(t, err)
	tmust(t, testdb.Migrate())

	_, err = testdb.Exec("DELETE FROM users")
	tmust(t, err)

	server, err := NewServer(Config{}, testdb)
	tmust(t, err)
	return server
}

func TestUsersApiNoAuth(t *testing.T) {
	server := InitDB(t)

	newReq := func(method, route string, body io.Reader) func() (*http.Request, error) {
		return func() (*http.Request, error) { return http.NewRequest(method, route, body) }
	}

	for _, requestFunc := range []func() (*http.Request, error){
		newReq("GET", "/api/v1/users", nil),
		newReq("PUT", "/api/v1/users", strings.NewReader(`{"name":"Alice"}`)),
		newReq("POST", "/api/v1/users/1", strings.NewReader(`{"name":"Bob"}`)),
		newReq("DELETE", "/api/v1/users/1", nil),
	} {
		req, err := requestFunc()
		tmust(t, err)
		res := httptest.NewRecorder()
		server.ServeHTTP(res, req)
		assert.Equal(t, 401, res.Code)
	}
}

func TestListUsersEmptyDB(t *testing.T) {
	server := InitDB(t)

	req, err := http.NewRequest("GET", "/api/v1/users", nil)
	req.SetBasicAuth("admin", "secret")
	tmust(t, err)
	res := httptest.NewRecorder()
	server.ServeHTTP(res, req)

	assert.Equal(t, 200, res.Code)
	assert.Equal(t, "[]", res.Body.String())
}

func TestListUsersWithOneUser(t *testing.T) {
	server := InitDB(t)

	_, err := server.DB.Exec("INSERT INTO users (name) VALUES ('Alice')")
	tmust(t, err)

	req, err := http.NewRequest("GET", "/api/v1/users", nil)
	req.SetBasicAuth("admin", "secret")
	tmust(t, err)
	res := httptest.NewRecorder()
	server.ServeHTTP(res, req)

	assert.Equal(t, 200, res.Code)
	var users []models.User
	tmust(t, json.Unmarshal(res.Body.Bytes(), &users))
	assert.Equal(t, 1, len(users))
	user := users[0]
	assert.LessOrEqual(t, 1, user.Id)
	assert.Equal(t, "Alice", user.Name)

	epsilon, err := time.ParseDuration("5s")
	tmust(t, err)
	assert.GreaterOrEqual(t, epsilon, time.Now().UTC().Sub(*user.CreatedAt))
	assert.GreaterOrEqual(t, epsilon, time.Now().UTC().Sub(*user.UpdatedAt))
	assert.Equal(t, user.CreatedAt, user.UpdatedAt)
}

func TestCreateUser(t *testing.T) {
	server := InitDB(t)

	reqBody := `{"name":"Bob"}`
	req, err := http.NewRequest("PUT", "/api/v1/users", strings.NewReader(reqBody))
	req.SetBasicAuth("admin", "secret")
	tmust(t, err)
	res := httptest.NewRecorder()
	server.ServeHTTP(res, req)

	assert.Equal(t, 200, res.Code)
	var user models.User
	tmust(t, json.Unmarshal(res.Body.Bytes(), &user))
	assert.LessOrEqual(t, 1, user.Id)
	assert.Equal(t, "Bob", user.Name)

	epsilon, err := time.ParseDuration("5s")
	tmust(t, err)
	assert.GreaterOrEqual(t, epsilon, time.Now().UTC().Sub(*user.CreatedAt))
	assert.GreaterOrEqual(t, epsilon, time.Now().UTC().Sub(*user.UpdatedAt))
	assert.Equal(t, user.CreatedAt, user.UpdatedAt)
}

func TestUpdateUser(t *testing.T) {
	server := InitDB(t)

	var userId int
	tmust(t, server.DB.Get(&userId, "INSERT INTO users (name) VALUES ('Alice') RETURNING id"))
	t1s, err := time.ParseDuration("1s")
	tmust(t, err)
	time.Sleep(t1s)

	// Update the user
	reqBody := `{"name":"Bob"}`
	uri := fmt.Sprintf("/api/v1/users/%v", userId)
	req, err := http.NewRequest("POST", uri, strings.NewReader(reqBody))
	req.SetBasicAuth("admin", "secret")
	tmust(t, err)
	res := httptest.NewRecorder()
	server.ServeHTTP(res, req)

	assert.Equal(t, 200, res.Code)
	var user models.User
	tmust(t, json.Unmarshal(res.Body.Bytes(), &user))
	assert.LessOrEqual(t, 1, user.Id)
	assert.Equal(t, "Bob", user.Name)
	assert.Greater(t, *user.UpdatedAt, *user.CreatedAt)

	epsilon, err := time.ParseDuration("5s")
	tmust(t, err)
	assert.GreaterOrEqual(t, epsilon, time.Now().UTC().Sub(*user.CreatedAt))
	assert.GreaterOrEqual(t, epsilon, time.Now().UTC().Sub(*user.UpdatedAt))
}

func TestDeleteUser(t *testing.T) {
	server := InitDB(t)

	var userId int
	tmust(t, server.DB.Get(&userId, "INSERT INTO users (name) VALUES ('Alice') RETURNING id"))

	// Delete the user
	uri := fmt.Sprintf("/api/v1/users/%v", userId)
	req, err := http.NewRequest("DELETE", uri, nil)
	req.SetBasicAuth("admin", "secret")
	tmust(t, err)
	res := httptest.NewRecorder()
	server.ServeHTTP(res, req)
	assert.Equal(t, 204, res.Code)

	var userCount int
	tmust(t, server.DB.Get(&userCount, "SELECT COUNT(id) FROM users"))
	assert.Equal(t, 0, userCount)
}
