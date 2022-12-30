package db

import (
	"testing"

	"github.com/maerics/golog"
	"github.com/maerics/goutil"
	"github.com/stretchr/testify/assert"
)

const Env_TEST_DATABASE_URL = "TEST_DATABASE_URL"

func MustConnectTestDB() *DB {
	return golog.Must1(Connect(goutil.MustEnv(Env_TEST_DATABASE_URL)))
}

func TestDbIsSqlDatabase(t *testing.T) {
	testdb := MustConnectTestDB()

	var one int
	if err := testdb.Get(&one, "SELECT 1"); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 1, one)
}
