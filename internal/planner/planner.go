package planner

import (
	"fmt"
	"strings"
)

// Projection represents a single SELECT expression and optional alias.
type Projection struct {
	Expr  string
	Alias string
}

// FindManyInput captures the pieces needed to render a SELECT query.
type FindManyInput struct {
	Table       string
	Alias       string
	Projections []Projection
	Joins       []string
	Distinct    []string
	Where       string
	Args        []any
	OrderBy     []string
	Limit       *int64
	Offset      *int64
	JSONArray   bool
}

// BuildFindMany builds a SQL statement for the supplied input tuple.
func BuildFindMany(d Dialect, in FindManyInput) (string, []any, error) {
	if in.Table == "" {
		return "", nil, fmt.Errorf("table name required")
	}
	if len(in.Projections) == 0 {
		return "", nil, fmt.Errorf("at least one projection required")
	}
	alias := in.Alias
	if alias == "" {
		alias = "t0"
	}
	var selectExprs []string
	for _, p := range in.Projections {
		expr := p.Expr
		if p.Alias != "" {
			expr = fmt.Sprintf("%s AS %s", expr, d.QuoteIdent(p.Alias))
		}
		selectExprs = append(selectExprs, expr)
	}
	var prefix string
	if len(in.Distinct) > 0 {
		if d.Cap().DistinctOn {
			prefix = fmt.Sprintf("DISTINCT ON (%s) ", strings.Join(in.Distinct, ", "))
		} else {
			prefix = "DISTINCT "
		}
	}
	base := strings.Builder{}
	base.WriteString("SELECT ")
	base.WriteString(prefix)
	base.WriteString(strings.Join(selectExprs, ", "))
	base.WriteString(" FROM ")
	base.WriteString(d.QuoteIdent(in.Table))
	base.WriteString(" AS ")
	base.WriteString(d.QuoteIdent(alias))
	for _, join := range in.Joins {
		if strings.TrimSpace(join) == "" {
			continue
		}
		base.WriteString(" ")
		base.WriteString(join)
	}
	if in.Where != "" {
		base.WriteString(" WHERE ")
		base.WriteString(in.Where)
	}
	if len(in.OrderBy) > 0 {
		base.WriteString(" ORDER BY ")
		base.WriteString(strings.Join(in.OrderBy, ", "))
	}
	if lim := d.LimitOffset(in.Limit, in.Offset); lim != "" {
		base.WriteString(" ")
		base.WriteString(lim)
	}
	sqlStr := base.String()
	if !in.JSONArray {
		return sqlStr, in.Args, nil
	}
	jsonCol := fmt.Sprintf("%s.%s", d.QuoteIdent("r"), d.QuoteIdent("__bom_json"))
	aggExpr := d.CoalesceJSONAgg(d.JSONArrayAgg(jsonCol), d.JSONArrayEmpty())
	aggExpr = d.JSONValue(aggExpr)
	fromFmt := "SELECT %s AS %s FROM (%s) AS %s"
	if d.Name() == "postgres" {
		fromFmt = "SELECT %s AS %s FROM LATERAL (%s) AS %s"
	}
	agg := fmt.Sprintf(fromFmt, aggExpr, d.QuoteIdent("__bom_json"), sqlStr, d.QuoteIdent("r"))
	return agg, in.Args, nil
}
