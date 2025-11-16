package mysql_test

import (
	"context"
	"database/sql"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

func TestGeneratedQueriesExecuteOnMySQL(t *testing.T) {
	dsn := os.Getenv("TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("set TEST_MYSQL_DSN to run MySQL integration test")
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("open mysql: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()

	if _, err := db.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS=0"); err != nil {
		t.Fatalf("disable fk checks: %v", err)
	}
	execScript(t, db, mysqlDropTablesSQL)
	execScript(t, db, mysqlSchemaSQL)
	if _, err := db.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS=1"); err != nil {
		t.Fatalf("enable fk checks: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(context.Background(), "SET FOREIGN_KEY_CHECKS=0")
		execScript(t, db, mysqlDropTablesSQL)
		db.ExecContext(context.Background(), "SET FOREIGN_KEY_CHECKS=1")
	})

	for _, stmt := range testFixtures {
		if _, err := db.ExecContext(ctx, stmt.sql, stmt.args...); err != nil {
			t.Fatalf("fixture %q failed: %v", stmt.sql, err)
		}
	}

	runRelationQueryAssertions(t, ctx, &sqlQuerier{db: db})
}
