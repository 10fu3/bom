package planner

import (
	"strings"
	"testing"
)

func TestBuildFindManyBasic(t *testing.T) {
	d := newMySQLDialect()
	limit := int64(5)
	input := FindManyInput{
		Table: "video",
		Alias: "t0",
		Projections: []Projection{
			{Expr: "t0.`id`", Alias: "id"},
			{Expr: "t0.`title`", Alias: "title"},
		},
		Where:   "`id` = ?",
		Args:    []any{1},
		OrderBy: []string{"`title` ASC"},
		Limit:   &limit,
	}

	sql, args, err := BuildFindMany(d, input)
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}
	expect := "SELECT t0.`id` AS `id`, t0.`title` AS `title` FROM `video` AS `t0` WHERE `id` = ? ORDER BY `title` ASC LIMIT 5"
	if sql != expect {
		t.Fatalf("unexpected sql:\n got %s\nwant %s", sql, expect)
	}
	if len(args) != 1 || args[0] != 1 {
		t.Fatalf("unexpected args: %#v", args)
	}
}

func TestBuildFindManyDistinctOnFallback(t *testing.T) {
	d := newMySQLDialect()
	input := FindManyInput{
		Table: "author",
		Alias: "a0",
		Projections: []Projection{
			{Expr: "a0.`email`", Alias: "email"},
		},
		Distinct: []string{"`email`"},
	}
	sql, _, err := BuildFindMany(d, input)
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}
	if !strings.HasPrefix(sql, "SELECT DISTINCT ") {
		t.Fatalf("expected DISTINCT prefix, got %s", sql)
	}
}

func TestBuildFindManyJSONArray(t *testing.T) {
	d := newMySQLDialect()
	input := FindManyInput{
		Table: "video",
		Alias: "t0",
		Projections: []Projection{
			{Expr: "t0.`id`", Alias: "__bom_json"},
		},
		JSONArray: true,
	}
	sql, _, err := BuildFindMany(d, input)
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}
	expect := "SELECT COALESCE(JSON_ARRAYAGG(`r`.`__bom_json`), JSON_ARRAY()) AS `__bom_json` FROM (SELECT t0.`id` AS `__bom_json` FROM `video` AS `t0`) AS `r`"
	if sql != expect {
		t.Fatalf("unexpected sql:\n got %s\nwant %s", sql, expect)
	}
}

func TestBuildFindManyJSONArrayPostgres(t *testing.T) {
	d := newPostgresDialect()
	input := FindManyInput{
		Table: "video",
		Alias: "t0",
		Projections: []Projection{
			{Expr: `"t0"."id"`, Alias: "__bom_json"},
		},
		JSONArray: true,
	}
	sql, _, err := BuildFindMany(d, input)
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}
	expect := `SELECT (COALESCE(JSON_AGG("r"."__bom_json"), '[]'::json))::json AS "__bom_json" FROM LATERAL (SELECT "t0"."id" AS "__bom_json" FROM "video" AS "t0") AS "r"`
	if sql != expect {
		t.Fatalf("unexpected sql:\n got %s\nwant %s", sql, expect)
	}
}

func TestBuildFindManyWithJoin(t *testing.T) {
	d := newPostgresDialect()
	input := FindManyInput{
		Table: "author",
		Alias: "t0",
		Projections: []Projection{
			{Expr: `"t0"."id"`, Alias: "id"},
		},
		Joins: []string{
			`LEFT JOIN LATERAL (SELECT 1 AS "__bom_json") AS "j0" ON true`,
		},
	}
	sql, _, err := BuildFindMany(d, input)
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}
	expect := `SELECT "t0"."id" AS "id" FROM "author" AS "t0" LEFT JOIN LATERAL (SELECT 1 AS "__bom_json") AS "j0" ON true`
	if sql != expect {
		t.Fatalf("unexpected sql:\n got %s\nwant %s", sql, expect)
	}
}
