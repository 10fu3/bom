package mysql_test

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/10fu3/bom/examples/mysql/generated"
	"github.com/10fu3/bom/pkg/opt"
	_ "github.com/go-sql-driver/mysql"
)

func setupMySQLTestDB(t *testing.T) (context.Context, *sql.DB) {
	t.Helper()
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
	return ctx, db
}

func TestGeneratedQueriesExecuteOnMySQL(t *testing.T) {
	ctx, db := setupMySQLTestDB(t)

	runRelationQueryAssertions(t, ctx, &sqlQuerier{db: db})
	runCreateAssertions(t, ctx, &sqlQuerier{db: db})
	runClauseInjectionAssertions(t, ctx, &sqlQuerier{db: db})
	runMutationInjectionAssertions(t, ctx, &sqlQuerier{db: db})
}

func TestGeneratedQueriesBlockSQLInjection(t *testing.T) {
	ctx, db := setupMySQLTestDB(t)
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
