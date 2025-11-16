package planner

import (
	"strings"
	"testing"

	"bom/pkg/dialect/mysql"
)

func TestBuildFindManyBasic(t *testing.T) {
	d := mysql.New()
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
	d := mysql.New()
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
	d := mysql.New()
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
