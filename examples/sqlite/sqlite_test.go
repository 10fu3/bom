//go:build moderncsqlite

package sqlite_test

import (
	"context"
	"database/sql"
	"testing"

	"bom/examples/sqlite/generated"
	"bom/pkg/opt"
	_ "modernc.org/sqlite"
)

func setupSQLiteTestDB(t *testing.T) (context.Context, *sql.DB) {
	t.Helper()
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

	return ctx, db
}

func TestGeneratedQueriesExecuteOnSQLite(t *testing.T) {
	ctx, db := setupSQLiteTestDB(t)

	runRelationQueryAssertions(t, ctx, &sqlQuerier{db: db})
	runCreateAssertions(t, ctx, &sqlQuerier{db: db})
	runClauseInjectionAssertions(t, ctx, &sqlQuerier{db: db})
	runMutationInjectionAssertions(t, ctx, &sqlQuerier{db: db})
}

func TestGeneratedQueriesBlockSQLInjection(t *testing.T) {
	ctx, db := setupSQLiteTestDB(t)
	querier := &sqlQuerier{db: db}

	payload := "alice@example.com' OR 1=1 --"
	rows, err := generated.FindManyAuthor[generated.Author](ctx, querier, generated.AuthorFindMany{
		Where: &generated.AuthorWhereInput{
			Email: opt.OVal(payload),
		},
		Select: generated.AuthorSelect{
			generated.AuthorFieldId,
		},
	})
	if err != nil {
		t.Fatalf("FindManyAuthor with malicious payload failed: %v", err)
	}
	if len(rows) != 0 {
		t.Fatalf("expected no matches for malicious payload, got %#v", rows)
	}

	safe, err := generated.FindManyAuthor[generated.Author](ctx, querier, generated.AuthorFindMany{
		Where: &generated.AuthorWhereInput{
			Email: opt.OVal("alice@example.com"),
		},
		Select: generated.AuthorSelect{
			generated.AuthorFieldId,
			generated.AuthorFieldEmail,
		},
	})
	if err != nil {
		t.Fatalf("FindManyAuthor for legitimate payload failed: %v", err)
	}
	if len(safe) != 1 || safe[0].Email != "alice@example.com" {
		t.Fatalf("expected Alice record intact, got %#v", safe)
	}
}
