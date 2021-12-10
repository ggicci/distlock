package distlock

import (
	"database/sql"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
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

func TestPostgreSQLProvider(t *testing.T) {
	db, err := sql.Open(
		"postgres",
		"user=root host=127.0.0.1 port=5432 dbname=test sslmode=disable",
	)
	if err != nil {
		t.Fatal(err)
	}

	provider, err := NewPostgreSQLProvider(db, "distlocks")
	if err != nil {
		t.Fatalf("could not create provider: %s", err)
	}

	runBasicLockTests(t, provider)
}