package postgres_test

import (
    "context"
    "database/sql"
    "os"
    "testing"

    _ "github.com/lib/pq"
)

func TestGeneratedQueriesExecuteOnPostgres(t *testing.T) {
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

    runRelationQueryAssertions(t, ctx, &sqlQuerier{db: db})
}
