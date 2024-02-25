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

const (
	ContentTypeHeaderValue = "content-type"
	ContentTypeTextJSON    = "text/json"
)

func InitTestServer(t *testing.T) *Server {
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
	server := InitTestServer(t)

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
	server := InitTestServer(t)

	req, err := http.NewRequest("GET", "/api/v1/users", nil)
	req.SetBasicAuth("admin", "secret")
	req.Header.Add(ContentTypeHeaderValue, ContentTypeTextJSON)
	tmust(t, err)
	res := httptest.NewRecorder()
	server.ServeHTTP(res, req)

	assert.Equal(t, 200, res.Code)
	assert.Equal(t, "[]", res.Body.String())
}

func TestListUsersWithOneUser(t *testing.T) {
	server := InitTestServer(t)

	_, err := server.DB.Exec("INSERT INTO users (email,password) VALUES ('foo','bar')")
	tmust(t, err)

	req, err := http.NewRequest("GET", "/api/v1/users", nil)
	req.SetBasicAuth("admin", "secret")
	req.Header.Add(ContentTypeHeaderValue, ContentTypeTextJSON)
	tmust(t, err)
	res := httptest.NewRecorder()
	server.ServeHTTP(res, req)

	assert.Equal(t, 200, res.Code)
	var users []models.User
	tmust(t, json.Unmarshal(res.Body.Bytes(), &users))
	assert.Equal(t, 1, len(users))
	user := users[0]
	assert.LessOrEqual(t, 1, user.Id)
	assert.Equal(t, "foo", user.Email)
	assert.Equal(t, "bar", user.Password)

	epsilon, err := time.ParseDuration("5s")
	tmust(t, err)
	assert.GreaterOrEqual(t, epsilon, time.Since(*user.CreatedAt))
	assert.GreaterOrEqual(t, epsilon, time.Since(*user.UpdatedAt))
	assert.Equal(t, user.CreatedAt, user.UpdatedAt)
}

func TestCreateUser(t *testing.T) {
	server := InitTestServer(t)

	reqBody := `{"email":"alice","password":"secret"}`
	req, err := http.NewRequest("PUT", "/api/v1/users", strings.NewReader(reqBody))
	req.SetBasicAuth("admin", "secret")
	req.Header.Add(ContentTypeHeaderValue, ContentTypeTextJSON)
	tmust(t, err)
	res := httptest.NewRecorder()
	server.ServeHTTP(res, req)

	assert.Equal(t, 200, res.Code)
	var user models.User
	tmust(t, json.Unmarshal(res.Body.Bytes(), &user))
	assert.LessOrEqual(t, 1, user.Id)
	assert.Equal(t, "alice", user.Email)
	assert.Equal(t, "", user.Password)

	epsilon, err := time.ParseDuration("5s")
	tmust(t, err)
	assert.GreaterOrEqual(t, epsilon, time.Since(*user.CreatedAt))
	assert.GreaterOrEqual(t, epsilon, time.Since(*user.UpdatedAt))
	assert.Equal(t, user.CreatedAt, user.UpdatedAt)
}

func TestGetUser(t *testing.T) {
	server := InitTestServer(t)

	// Create the new user
	reqBody := `{"email":"alice","password":"password"}`
	req, err := http.NewRequest("PUT", "/api/v1/users", strings.NewReader(reqBody))
	req.SetBasicAuth("admin", "secret")
	req.Header.Add(ContentTypeHeaderValue, ContentTypeTextJSON)
	tmust(t, err)
	res := httptest.NewRecorder()
	server.ServeHTTP(res, req)

	assert.Equal(t, 200, res.Code)
	var user models.User
	tmust(t, json.Unmarshal(res.Body.Bytes(), &user))

	// Fetch the new user by id
	req2, err := http.NewRequest("GET", fmt.Sprintf("/api/v1/users/%v", user.Id), nil)
	req2.SetBasicAuth("admin", "secret")
	req2.Header.Add(ContentTypeHeaderValue, ContentTypeTextJSON)
	tmust(t, err)
	res2 := httptest.NewRecorder()
	server.ServeHTTP(res2, req2)
	var gotUser models.User
	tmust(t, json.Unmarshal(res2.Body.Bytes(), &gotUser))
	assert.Equal(t, user, gotUser)
}

func TestUpdateUser(t *testing.T) {
	server := InitTestServer(t)

	// Create the new user
	reqBody := `{"email":"alice","password":"password"}`
	req, err := http.NewRequest("PUT", "/api/v1/users", strings.NewReader(reqBody))
	req.SetBasicAuth("admin", "secret")
	req.Header.Add(ContentTypeHeaderValue, ContentTypeTextJSON)
	tmust(t, err)
	res := httptest.NewRecorder()
	server.ServeHTTP(res, req)

	assert.Equal(t, 200, res.Code)
	var user models.User
	tmust(t, json.Unmarshal(res.Body.Bytes(), &user))

	// Update the user
	reqBody = `{"email":"bob","password":"123456"}`
	uri := fmt.Sprintf("/api/v1/users/%v", user.Id)
	req2, err := http.NewRequest("POST", uri, strings.NewReader(reqBody))
	req2.SetBasicAuth("admin", "secret")
	req2.Header.Add(ContentTypeHeaderValue, ContentTypeTextJSON)
	tmust(t, err)
	res2 := httptest.NewRecorder()
	server.ServeHTTP(res2, req2)
	assert.Equal(t, 200, res2.Code)
	var updatedUser models.User
	tmust(t, json.Unmarshal(res2.Body.Bytes(), &updatedUser))
	assert.Equal(t, "bob", updatedUser.Email)
	assert.True(t, strings.HasPrefix(updatedUser.Password, "$2a$10"))

	// TODO: Fetch the user
}

func TestDeleteUser(t *testing.T) {
	server := InitTestServer(t)

	var userId int
	tmust(t, server.DB.Get(&userId, "INSERT INTO users (email,password) VALUES ('alice','password') RETURNING id"))

	// Delete the user
	uri := fmt.Sprintf("/api/v1/users/%v", userId)
	req, err := http.NewRequest("DELETE", uri, nil)
	req.SetBasicAuth("admin", "secret")
	req.Header.Add(ContentTypeHeaderValue, ContentTypeTextJSON)
	tmust(t, err)
	res := httptest.NewRecorder()
	server.ServeHTTP(res, req)
	assert.Equal(t, 204, res.Code)

	var userCount int
	tmust(t, server.DB.Get(&userCount, "SELECT COUNT(id) FROM users"))
	assert.Equal(t, 0, userCount)
}
