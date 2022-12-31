package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

func TestListUsersNoAuth(t *testing.T) {
	server := InitDB(t)

	res := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/api/v1/users", nil)
	tmust(t, err)
	server.ServeHTTP(res, req)

	assert.Equal(t, 401, res.Code)
}

func TestListUsersEmptyDB(t *testing.T) {
	server := InitDB(t)

	res := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/api/v1/users", nil)
	req.SetBasicAuth("admin", "secret")
	tmust(t, err)
	server.ServeHTTP(res, req)

	assert.Equal(t, 200, res.Code)
	assert.Equal(t, "[]", res.Body.String())
}

func TestListUsersWithOneUser(t *testing.T) {
	server := InitDB(t)

	_, err := server.DB.Exec("INSERT INTO users (name) VALUES ('Alice')")
	tmust(t, err)

	res := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/api/v1/users", nil)
	req.SetBasicAuth("admin", "secret")
	tmust(t, err)
	server.ServeHTTP(res, req)

	assert.Equal(t, 200, res.Code)
	var users []models.User
	tmust(t, json.Unmarshal(res.Body.Bytes(), &users))
	assert.Equal(t, 1, len(users))
	user := users[0]
	assert.Equal(t, 1, user.Id)
	assert.Equal(t, "Alice", user.Name)

	epsilon, err := time.ParseDuration("5s")
	tmust(t, err)
	assert.GreaterOrEqual(t, epsilon, time.Now().UTC().Sub(*user.CreatedAt))
	assert.GreaterOrEqual(t, epsilon, time.Now().UTC().Sub(*user.UpdatedAt))
	assert.Equal(t, user.CreatedAt, user.UpdatedAt)
}
