package distlock

import (
	"database/sql"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

const TestTableName = "ggicci_distlock_test"

func cleanupMySQL(db *sql.DB) {
	_, _ = db.Exec(formatSQL("DROP TABLE IF EXISTS %s", TestTableName))
}

func cleanupPostgreSQL(db *sql.DB) {
	_, _ = db.Exec(formatSQL("DROP TABLE IF EXISTS %s", TestTableName))
}

func TestMySQLProvider(t *testing.T) {
	db, err := sql.Open("mysql", "root@tcp(localhost:3306)/test")
	if err != nil {
		t.Fatal(err)
	}
	cleanupMySQL(db)

	provider, err := NewMySQLProvider(db, TestTableName)
	if err != nil {
		t.Fatalf("could not create provider: %s", err)
	}
	runLockTestsWithoutLifetime(t, provider)
	runLockTestsWithLifetime(t, provider)
}

func TestPostgreSQLProvider(t *testing.T) {
	db, err := sql.Open(
		"postgres",
		"user=root host=127.0.0.1 port=5432 dbname=test sslmode=disable",
	)
	if err != nil {
		t.Fatal(err)
	}

	cleanupPostgreSQL(db)

	provider, err := NewPostgreSQLProvider(db, TestTableName)
	if err != nil {
		t.Fatalf("could not create provider: %s", err)
	}
	runLockTestsWithoutLifetime(t, provider)
	runLockTestsWithLifetime(t, provider)
}
