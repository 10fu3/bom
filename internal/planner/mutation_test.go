package planner

import (
	"testing"

	"bom/pkg/dialect/mysql"
	"bom/pkg/dialect/postgres"
)

func TestBuildInsertBasic(t *testing.T) {
	d := mysql.New()
	sql, err := BuildInsert(d, InsertInput{
		Table:   "video",
		Columns: []string{"id", "title"},
		Values:  []string{"?", "?"},
	})
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}
	expect := "INSERT INTO `video` (`id`, `title`) VALUES (?, ?)"
	if sql != expect {
		t.Fatalf("unexpected sql: %s", sql)
	}
}

func TestBuildInsertReturning(t *testing.T) {
	d := postgres.New()
	sql, err := BuildInsert(d, InsertInput{
		Table:     "author",
		Columns:   []string{"id"},
		Values:    []string{"$1"},
		Returning: []string{"id"},
	})
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}
	expect := `INSERT INTO "author" ("id") VALUES ($1) RETURNING "id"`
	if sql != expect {
		t.Fatalf("unexpected sql: %s", sql)
	}
}

func TestBuildInsertDefaultValues(t *testing.T) {
	d := postgres.New()
	sql, err := BuildInsert(d, InsertInput{Table: "tag"})
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}
	expect := `INSERT INTO "tag" DEFAULT VALUES`
	if sql != expect {
		t.Fatalf("unexpected sql: %s", sql)
	}
}

func TestBuildUpdate(t *testing.T) {
	d := mysql.New()
	sql, err := BuildUpdate(d, UpdateInput{
		Table:      "video",
		SetClauses: []string{"`title` = ?"},
		Where:      "`id` = ?",
	})
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}
	expect := "UPDATE `video` SET `title` = ? WHERE `id` = ?"
	if sql != expect {
		t.Fatalf("unexpected sql: %s", sql)
	}
}

func TestBuildDelete(t *testing.T) {
	d := mysql.New()
	sql, err := BuildDelete(d, DeleteInput{
		Table: "video",
		Where: "`id` = ?",
	})
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}
	expect := "DELETE FROM `video` WHERE `id` = ?"
	if sql != expect {
		t.Fatalf("unexpected sql: %s", sql)
	}
}
