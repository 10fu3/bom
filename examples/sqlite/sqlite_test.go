//go:build moderncsqlite

package sqlite_test

import (
	"context"
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func TestGeneratedQueriesExecuteOnSQLite(t *testing.T) {
	ctx := context.Background()
	db, err := sql.Open("sqlite", "file:examples_sqlite.sqlite?mode=memory&_pragma=foreign_keys(ON)")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	execScript(t, db, sqliteSchemaSQL)
	for _, stmt := range testFixtures {
		if _, err := db.ExecContext(ctx, stmt.sql, stmt.args...); err != nil {
			t.Fatalf("fixture %q failed: %v", stmt.sql, err)
		}
	}

	runRelationQueryAssertions(t, ctx, &sqlQuerier{db: db})
	runCreateAssertions(t, ctx, &sqlQuerier{db: db})
}
