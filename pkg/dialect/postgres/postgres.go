package postgres

import (
	"fmt"
	"strings"

	"bom/pkg/dialect"
)

type postgresDialect struct{}

func New() dialect.Dialect {
	return &postgresDialect{}
}

func (postgresDialect) Name() string { return "postgres" }

func (postgresDialect) Cap() dialect.Capabilities {
	return dialect.Capabilities{
		DistinctOn:      true,
		CaseInsensitive: dialect.ILike,
		Placeholder:     "$n",
	}
}

func (postgresDialect) QuoteIdent(id string) string {
	return `"` + strings.ReplaceAll(id, `"`, `""`) + `"`
}

func (postgresDialect) Placeholder(n int) string {
	return fmt.Sprintf("$%d", n)
}

func (postgresDialect) Eq(l, r string) string { return fmt.Sprintf("%s = %s", l, r) }
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

func (postgresDialect) JSONArrayEmpty() string {
	return "'[]'::jsonb"
}

func (postgresDialect) CoalesceJSONAgg(expr, empty string) string {
	return fmt.Sprintf("COALESCE(%s, %s)", expr, empty)
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
