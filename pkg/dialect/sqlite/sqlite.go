package sqlite

import (
	"fmt"
	"strings"

	"bom/pkg/dialect"
)

type sqliteDialect struct{}

func New() dialect.Dialect {
	return &sqliteDialect{}
}

func (sqliteDialect) Name() string { return "sqlite" }

func (sqliteDialect) Cap() dialect.Capabilities {
	return dialect.Capabilities{
		DistinctOn:      false,
		CaseInsensitive: dialect.CollateCI,
		Placeholder:     "?",
		InsertReturning: false,
		MaxParameters:   999,
	}
}

func (sqliteDialect) QuoteIdent(id string) string {
	return `"` + strings.ReplaceAll(id, `"`, `""`) + `"`
}

func (sqliteDialect) Placeholder(_ int) string { return "?" }

func (sqliteDialect) Eq(l, r string) string { return fmt.Sprintf("%s = %s", l, r) }
func (sqliteDialect) And(preds ...string) string {
	return "(" + strings.Join(preds, " AND ") + ")"
}
func (sqliteDialect) Or(preds ...string) string {
	return "(" + strings.Join(preds, " OR ") + ")"
}
func (sqliteDialect) Not(pred string) string { return fmt.Sprintf("NOT (%s)", pred) }

func (sqliteDialect) InsensitiveLike(lhs, ph string) string {
	return fmt.Sprintf("%s LIKE %s COLLATE NOCASE", lhs, ph)
}

func (sqliteDialect) JSONBuildObject(pairs ...string) string {
	return fmt.Sprintf("json_object(%s)", strings.Join(pairs, ", "))
}

func (sqliteDialect) JSONArrayAgg(expr string) string {
	return fmt.Sprintf("json_group_array(%s)", expr)
}

func (sqliteDialect) JSONArrayEmpty() string {
	return "json('[]')"
}

func (sqliteDialect) CoalesceJSONAgg(expr, empty string) string {
	return fmt.Sprintf("COALESCE(%s, %s)", expr, empty)
}

func (sqliteDialect) JSONValue(expr string) string {
	return fmt.Sprintf("json(%s)", expr)
}

func (sqliteDialect) LimitOffset(limit, offset *int64) string {
	switch {
	case limit == nil && offset == nil:
		return ""
	case limit != nil && offset != nil:
		return fmt.Sprintf("LIMIT %d OFFSET %d", *limit, *offset)
	case limit != nil:
		return fmt.Sprintf("LIMIT %d", *limit)
	default:
		return fmt.Sprintf("OFFSET %d", *offset)
	}
}

func (sqliteDialect) DistinctProjection(cols []string) (string, bool) {
	return "", false
}
