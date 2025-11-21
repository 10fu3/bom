package postgres_test

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"bom/examples/postgres/generated"
	"bom/pkg/opt"
	_ "github.com/lib/pq"
)

func setupPostgresTestDB(t *testing.T) (context.Context, *sql.DB) {
	t.Helper()
	dsn := os.Getenv("TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("set TEST_POSTGRES_DSN to run Postgres integration test")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open postgres: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()

	if _, err := db.ExecContext(ctx, "SET session_replication_role = replica"); err != nil {
		t.Fatalf("disable fk checks: %v", err)
	}
	execScript(t, db, postgresDropTablesSQL)
	execScript(t, db, postgresSchemaSQL)
	if _, err := db.ExecContext(ctx, "SET session_replication_role = origin"); err != nil {
		t.Fatalf("enable fk checks: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(context.Background(), "SET session_replication_role = replica")
		execScript(t, db, postgresDropTablesSQL)
		db.ExecContext(context.Background(), "SET session_replication_role = origin")
	})

	for _, stmt := range postgresFixtures {
		if _, err := db.ExecContext(ctx, stmt.sql, stmt.args...); err != nil {
			t.Fatalf("fixture %q failed: %v", stmt.sql, err)
		}
	}
	return ctx, db
}

func TestGeneratedQueriesExecuteOnPostgres(t *testing.T) {
	ctx, db := setupPostgresTestDB(t)

	runRelationQueryAssertions(t, ctx, &sqlQuerier{db: db})
	runCreateAssertions(t, ctx, &sqlQuerier{db: db})
	runClauseInjectionAssertions(t, ctx, &sqlQuerier{db: db})
	runMutationInjectionAssertions(t, ctx, &sqlQuerier{db: db})
}

func TestGeneratedQueriesBlockSQLInjection(t *testing.T) {
	ctx, db := setupPostgresTestDB(t)
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
