package distlock

import (
	"database/sql"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

func TestMySQLProvider(t *testing.T) {
	db, err := sql.Open("mysql", "root@tcp(localhost:3306)/test")
	if err != nil {
		t.Fatal(err)
	}

	provider, err := NewMySQLProvider(db, "distlocks")
	if err != nil {
		t.Fatalf("could not create provider: %s", err)
	}

	runBasicLockTests(t, provider)
}
