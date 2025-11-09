package mysql

import (
	"fmt"
	"strings"

	"bom/pkg/dialect"
)

type mysqlDialect struct{}

func New() dialect.Dialect {
	return &mysqlDialect{}
}

func (mysqlDialect) Name() string { return "mysql" }

func (mysqlDialect) Cap() dialect.Capabilities {
	return dialect.Capabilities{
		DistinctOn:      false,
		CaseInsensitive: dialect.LowerLike,
		Placeholder:     "?",
	}
}

func (mysqlDialect) QuoteIdent(id string) string {
	return "`" + strings.ReplaceAll(id, "`", "``") + "`"
}

func (mysqlDialect) Placeholder(_ int) string { return "?" }

func (mysqlDialect) Eq(l, r string) string { return fmt.Sprintf("%s = %s", l, r) }
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

func (mysqlDialect) JSONArrayAgg(expr string) string {
	return fmt.Sprintf("JSON_ARRAYAGG(%s)", expr)
}

func (mysqlDialect) JSONArrayEmpty() string {
	return "JSON_ARRAY()"
}

func (mysqlDialect) CoalesceJSONAgg(expr, empty string) string {
	return fmt.Sprintf("COALESCE(%s, %s)", expr, empty)
}

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

func (mysqlDialect) DistinctProjection(cols []string) (string, bool) {
	return "", false
}
