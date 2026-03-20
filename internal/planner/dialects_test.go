package planner

import (
	"fmt"
	"strings"
)

type mysqlDialect struct{}

type postgresDialect struct{}

func newMySQLDialect() Dialect    { return mysqlDialect{} }
func newPostgresDialect() Dialect { return postgresDialect{} }

func (mysqlDialect) Name() string { return "mysql" }
func (mysqlDialect) Cap() Capabilities {
	return Capabilities{
		DistinctOn:      false,
		CaseInsensitive: LowerLike,
		Placeholder:     "?",
		InsertReturning: false,
		MaxParameters:   65535,
	}
}
func (mysqlDialect) QuoteIdent(id string) string {
	return "`" + strings.ReplaceAll(id, "`", "``") + "`"
}
func (mysqlDialect) Placeholder(int) string { return "?" }
func (mysqlDialect) Eq(l, r string) string  { return fmt.Sprintf("%s = %s", l, r) }
func (mysqlDialect) And(preds ...string) string {
	return "(" + strings.Join(preds, " AND ") + ")"
}
func (mysqlDialect) Or(preds ...string) string {
	return "(" + strings.Join(preds, " OR ") + ")"
}
func (mysqlDialect) Not(pred string) string { return fmt.Sprintf("NOT (%s)", pred) }
func (mysqlDialect) InsensitiveLike(lhs, ph string) string {
	return fmt.Sprintf("LOWER(%s) LIKE LOWER(%s)", lhs, ph)
}
func (mysqlDialect) JSONBuildObject(pairs ...string) string {
	return fmt.Sprintf("JSON_OBJECT(%s)", strings.Join(pairs, ", "))
}
func (mysqlDialect) JSONArrayAgg(expr string) string { return fmt.Sprintf("JSON_ARRAYAGG(%s)", expr) }
func (mysqlDialect) JSONArrayEmpty() string          { return "JSON_ARRAY()" }
func (mysqlDialect) CoalesceJSONAgg(expr, empty string) string {
	return fmt.Sprintf("COALESCE(%s, %s)", expr, empty)
}
func (mysqlDialect) JSONValue(expr string) string { return expr }
func (mysqlDialect) LimitOffset(limit, offset *int64) string {
	switch {
	case limit == nil && offset == nil:
		return ""
	case limit != nil && offset != nil:
		return fmt.Sprintf("LIMIT %d OFFSET %d", *limit, *offset)
	case limit != nil:
		return fmt.Sprintf("LIMIT %d", *limit)
	default:
		return fmt.Sprintf("LIMIT 18446744073709551615 OFFSET %d", *offset)
	}
}
func (mysqlDialect) DistinctProjection(cols []string) (string, bool) { return "", false }

func (postgresDialect) Name() string { return "postgres" }
func (postgresDialect) Cap() Capabilities {
	return Capabilities{
		DistinctOn:      true,
		CaseInsensitive: ILike,
		Placeholder:     "$n",
		InsertReturning: true,
		MaxParameters:   65535,
	}
}
func (postgresDialect) QuoteIdent(id string) string {
	return "\"" + strings.ReplaceAll(id, "\"", "\"\"") + "\""
}
func (postgresDialect) Placeholder(n int) string { return fmt.Sprintf("$%d", n) }
func (postgresDialect) Eq(l, r string) string    { return fmt.Sprintf("%s = %s", l, r) }
func (postgresDialect) And(preds ...string) string {
	return "(" + strings.Join(preds, " AND ") + ")"
}
func (postgresDialect) Or(preds ...string) string {
	return "(" + strings.Join(preds, " OR ") + ")"
}
func (postgresDialect) Not(pred string) string { return fmt.Sprintf("NOT (%s)", pred) }
func (postgresDialect) InsensitiveLike(lhs, ph string) string {
	return fmt.Sprintf("%s ILIKE %s", lhs, ph)
}
func (postgresDialect) JSONBuildObject(pairs ...string) string {
	return fmt.Sprintf("JSON_BUILD_OBJECT(%s)", strings.Join(pairs, ", "))
}
func (postgresDialect) JSONArrayAgg(expr string) string {
	return fmt.Sprintf("JSON_AGG(%s)", expr)
}
func (postgresDialect) JSONArrayEmpty() string { return "'[]'::json" }
func (postgresDialect) CoalesceJSONAgg(expr, empty string) string {
	return fmt.Sprintf("COALESCE(%s, %s)", expr, empty)
}
func (postgresDialect) JSONValue(expr string) string {
	return fmt.Sprintf("(%s)::json", expr)
}
func (postgresDialect) LimitOffset(limit, offset *int64) string {
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
func (postgresDialect) DistinctProjection(cols []string) (string, bool) {
	if len(cols) == 0 {
		return "", false
	}
	return fmt.Sprintf("DISTINCT ON (%s)", strings.Join(cols, ", ")), true
}
